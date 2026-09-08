package ws

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/yourorg/livemix/config"
	"github.com/yourorg/livemix/internal/middleware"
	"github.com/yourorg/livemix/internal/model"
	"github.com/yourorg/livemix/internal/service"
	"github.com/yourorg/livemix/pkg/errcode"
	"github.com/yourorg/livemix/pkg/jwt"
	"github.com/yourorg/livemix/pkg/logger"
	"go.uber.org/zap"
)

// upgrader WebSocket升级器
// 安全策略：CheckOrigin验证请求来源是否在CORS白名单中
// 非白名单来源的WebSocket连接将被拒绝，防止CSRF攻击
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     getCheckOriginFunc(),
}

// getCheckOriginFunc 获取Origin验证函数
// 优先使用全局Origin验证器，未初始化时拒绝所有非同源请求
// 安全降级说明：如果验证器未初始化（如测试环境），仅允许无Origin头的同源请求
func getCheckOriginFunc() func(r *http.Request) bool {
	return func(r *http.Request) bool {
		validator := middleware.GetOriginValidator()
		if validator != nil {
			return validator.IsAllowed(r.Header.Get("Origin"))
		}
		// 验证器未初始化时的安全降级策略：仅允许无Origin头的同源请求
		// 浏览器同源请求不会携带Origin头，因此这类请求可以安全放行
		origin := r.Header.Get("Origin")
		return origin == ""
	}
}

// Handler WebSocket处理器
type Handler struct {
	hub        *Hub
	channelSvc *service.ChannelService
	serverSvc  *service.ServerService
}

// NewHandler 创建WebSocket处理器实例
func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

// SetServices wires the channel authorization service used for validating
// channel-scoped events. Keeping this separate preserves the lightweight Hub-only
// constructor used by tests and embedders.
func (h *Handler) SetServices(channelSvc *service.ChannelService, serverSvc *service.ServerService) {
	h.channelSvc = channelSvc
	h.serverSvc = serverSvc
	h.hub.setChannelService(channelSvc)
}

// HandleWebSocket 处理WebSocket连接
// per-app 单例连接，不要求 room_id 参数
// 客户端连接后通过 join_channel/leave_channel 事件订阅频道
// 安全策略：JWT认证 + Origin验证 + 消息大小限制 + 速率限制
func (h *Handler) HandleWebSocket(c *gin.Context) {
	// 从查询参数获取token
	tokenString := c.Query("token")

	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 40003, "message": "缺少认证信息"})
		return
	}

	// 验证token
	claims, err := jwt.ParseAccessToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 40003, "message": "Token无效"})
		return
	}

	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Warn("WebSocket升级失败",
			logger.WithError(err),
			zap.String("clientIP", c.ClientIP()),
			zap.String("path", c.Request.URL.Path),
		)
		return
	}

	// 设置消息大小限制（防止 DoS 攻击）
	cfg := config.Get()
	maxMessageSize := cfg.WebSocket.MaxMessageSize
	if maxMessageSize <= 0 {
		maxMessageSize = 65536 // 默认 64KB
	}
	conn.SetReadLimit(maxMessageSize)

	// 创建客户端（不绑定任何频道，Channels 初始为空）
	client := &Client{
		ID:             generateClientID(),
		UserID:         claims.UserID,
		Username:       claims.Username,
		Channels:       make(map[uint64]bool),
		Conn:           conn,
		Send:           make(chan []byte, 256),
		Hub:            h.hub,
		channelSvc:     h.channelSvc,
		serverSvc:      h.serverSvc,
		lastActiveTime: time.Now(),
	}

	// 注册客户端
	h.hub.register <- client

	// 启动读写协程
	go client.writePump()
	go client.readPump()
}

// readPump 读取客户端消息
// 心跳机制：
// - 每次收到消息时更新客户端活跃时间
// - 收到Pong响应时更新活跃时间并重置读超时
// - 超时未收到消息或Pong，连接将被Hub的心跳检测关闭
// 安全策略：速率限制，防止 DoS 攻击
func (c *Client) readPump() {
	defer c.requestUnregister()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		// 收到Pong响应，更新活跃时间并重置读超时
		c.UpdateLastActiveTime()
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		// 收到消息，更新活跃时间
		c.UpdateLastActiveTime()

		// 速率限制检查
		if c.Hub.rateLimiter != nil && !c.Hub.rateLimiter.AllowMessage(c.ID) {
			c.Hub.SendToClient(c.ID, EventError, map[string]any{
				"code":    40503,
				"message": "消息发送频率过高，请稍后再试",
			})
			continue
		}

		// 解析消息
		var event Event
		if err := json.Unmarshal(message, &event); err != nil {
			continue
		}

		// 处理事件
		c.handleEvent(event)
	}
}

// writePump 向客户端写入消息
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Each queued message was marshaled as one complete JSON document.
			// Keep one document per text frame; newline-delimited batching makes
			// WebSocket frames ambiguous and breaks strict clients.
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleEvent 处理客户端事件
// 安全策略：所有用户输入的消息内容在广播前必须进行XSS过滤，防止反射型XSS攻击
func (c *Client) handleEvent(event Event) {
	// Return protocol errors to the sender without exposing internal details.
	reject := func(err error) {
		message := "无权执行此操作"
		if typed, ok := err.(*errcode.Error); ok {
			message = typed.Message
		}
		c.Hub.SendToClient(c.ID, EventError, map[string]any{
			"event":   event.Event,
			"message": message,
		})
	}

	validateChannel := func(channelID uint64, permission int64) (*model.Channel, bool) {
		if channelID == 0 || c.channelSvc == nil {
			reject(errcode.ErrInvalidParam)
			return nil, false
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		channel, err := c.channelSvc.ValidateAccess(ctx, channelID, c.UserID, permission)
		if err != nil {
			reject(err)
			return nil, false
		}
		return channel, true
	}

	switch event.Event {
	case EventHeartbeat:
		// 心跳事件，仅更新活跃时间，无需响应
		// 活跃时间已在readPump中更新，此处无需额外处理
		// 客户端可发送空心跳事件保持连接活跃

	case EventJoinChannel:
		// 订阅频道，必须具备频道查看权限。
		data, ok := event.Data.(map[string]any)
		if !ok {
			reject(errcode.ErrInvalidParam)
			return
		}
		channelID, ok := parseChannelID(data["channelId"])
		if !ok {
			reject(errcode.ErrInvalidParam)
			return
		}
		channel, ok := validateChannel(channelID, model.PermViewChannel)
		if !ok {
			return
		}
		if c.IsSubscribed(channelID) {
			return
		}

		c.Hub.JoinChannel(c, channelID)

		member := map[string]any{
			"serverId": channel.ServerID,
			"userId":   c.UserID,
			"username": c.Username,
		}
		if c.serverSvc != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			memberResponse, err := c.serverSvc.GetMember(ctx, channel.ServerID, c.UserID)
			cancel()
			if err == nil {
				member = map[string]any{
					"id":        memberResponse.ID,
					"serverId":  memberResponse.ServerID,
					"userId":    memberResponse.UserID,
					"username":  memberResponse.Username,
					"avatarUrl": memberResponse.AvatarURL,
					"nickname":  memberResponse.Nickname,
					"joinedAt":  memberResponse.JoinedAt,
					"roles":     memberResponse.Roles,
				}
			}
		}

		// 广播 member_joined 到该频道（排除自己）。 The member object
		// matches the current frontend contract; the client refreshes the full
		// member list using serverId.
		c.Hub.BroadcastToChannelExcept(channelID, EventMemberJoined, map[string]any{
			"serverId": channel.ServerID,
			"member":   member,
		}, c.ID)

	case EventLeaveChannel:
		// 离开频道前仍需验证用户属于该服务器且有频道查看权限。
		data, ok := event.Data.(map[string]any)
		if !ok {
			reject(errcode.ErrInvalidParam)
			return
		}
		channelID, ok := parseChannelID(data["channelId"])
		if !ok {
			reject(errcode.ErrInvalidParam)
			return
		}
		if _, ok = validateChannel(channelID, model.PermViewChannel); !ok {
			return
		}
		if !c.IsSubscribed(channelID) {
			reject(errcode.ErrBadRequest.WithMessage("未订阅该频道"))
			return
		}

		c.Hub.LeaveChannel(c, channelID)

		// 广播 member_left 到该频道（排除自己）
		c.Hub.BroadcastToChannelExcept(channelID, EventMemberLeft, map[string]any{
			"channelId": channelID,
			"userId":    c.UserID,
		}, c.ID)

	case EventChatMessage:
		// Chat persistence and publication are owned by the REST endpoint. The
		// legacy client-to-server event is intentionally rejected so it cannot
		// create a second non-persistent broadcast path.
		reject(errcode.ErrBadRequest.WithMessage("聊天消息请使用 REST API"))

	case EventWebRTCSignal:
		// WebRTC signals are only relayed between users sharing an authorized
		// channel. The optional mediaType is copied to both the envelope and the
		// nested signal for clients using either protocol shape.
		data, ok := event.Data.(map[string]any)
		if !ok {
			reject(errcode.ErrInvalidParam)
			return
		}
		signalType, ok := data["type"].(string)
		if !ok || (signalType != "offer" && signalType != "answer" && signalType != "ice-candidate") {
			reject(errcode.ErrInvalidParam.WithMessage("无效的信令类型"))
			return
		}
		targetID, ok := parseChannelID(data["targetId"])
		if !ok || targetID == c.UserID {
			reject(errcode.ErrInvalidParam.WithMessage("目标用户无效"))
			return
		}
		payload, exists := data["payload"]
		if !exists || payload == nil {
			reject(errcode.ErrInvalidParam.WithMessage("信令负载不能为空"))
			return
		}

		mediaType, hasMediaType := "", false
		if rawMediaType, exists := data["mediaType"]; exists {
			var ok bool
			mediaType, ok = rawMediaType.(string)
			if !ok {
				reject(errcode.ErrInvalidParam.WithMessage("无效的媒体类型"))
				return
			}
			hasMediaType = true
			if mediaType != "voice" && mediaType != "screen" {
				reject(errcode.ErrInvalidParam.WithMessage("无效的媒体类型"))
				return
			}
		}

		channelID, ok := c.Hub.SharedChannelWithUserID(c, targetID)
		if !ok {
			reject(errcode.ErrForbidden.WithMessage("目标用户不在同一频道"))
			return
		}
		channel, ok := validateChannel(channelID, model.PermViewChannel)
		if !ok {
			return
		}
		if channel.Type != model.ChannelTypeVoice {
			reject(errcode.ErrBadRequest.WithMessage("信令频道必须是语音频道"))
			return
		}
		// Legacy signals omitted mediaType and historically represented voice.
		// Authorize both peers for the media capability, not just channel view.
		peerPermission := model.PermConnectVoice
		if mediaType == "screen" {
			peerPermission = model.PermScreenShare
		}
		if _, ok = validateChannel(channelID, peerPermission); !ok {
			return
		}

		nestedSignal := map[string]any{
			"type":     signalType,
			"targetId": targetID,
			"payload":  payload,
		}
		if hasMediaType {
			nestedSignal["mediaType"] = mediaType
		}
		signalData := map[string]any{
			"fromUserId":   c.UserID,
			"fromUsername": c.Username,
			"signal":       nestedSignal,
		}
		if hasMediaType {
			signalData["mediaType"] = mediaType
		}
		// Re-check the recipient against the current membership/permission state;
		// a stale subscription must not keep receiving signals after access is revoked.
		if c.channelSvc == nil {
			reject(errcode.ErrInternalServer.WithMessage("频道服务不可用"))
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_, targetErr := c.channelSvc.ValidateAccess(ctx, channelID, targetID, peerPermission)
		cancel()
		if targetErr != nil {
			reject(errcode.ErrForbidden.WithMessage("目标用户无权访问该频道"))
			return
		}

		c.Hub.SendToUserInChannel(channelID, targetID, EventWebRTCSignal, signalData)

	case EventVoiceJoin:
		// 加入语音频道需要语音频道类型和连接语音权限。
		data, ok := event.Data.(map[string]any)
		if !ok {
			reject(errcode.ErrInvalidParam)
			return
		}
		channelID, ok := parseChannelID(data["channelId"])
		if !ok {
			reject(errcode.ErrInvalidParam)
			return
		}
		channel, ok := validateChannel(channelID, model.PermConnectVoice)
		if !ok {
			return
		}
		if channel.Type != model.ChannelTypeVoice {
			reject(errcode.ErrBadRequest.WithMessage("该频道不是语音频道"))
			return
		}

		// A voice event is meaningful only after the connection has subscribed
		// to the channel. Subscription and voice permission are separate checks.
		if !c.IsSubscribed(channelID) {
			reject(errcode.ErrForbidden.WithMessage("未订阅该频道"))
			return
		}

		// 广播 voice_user_joined 到该频道（排除自己）。 Keep the
		// participant under `user`, matching the current frontend contract.
		user := map[string]any{
			"id":       c.UserID,
			"userId":   c.UserID,
			"username": c.Username,
		}
		c.Hub.BroadcastToChannelExcept(channelID, EventVoiceUserJoined, map[string]any{
			"channelId": channelID,
			"user":      user,
		}, c.ID)

	case EventVoiceLeave:
		// 离开语音频道需要验证频道访问权限，并且必须是当前订阅者。
		data, ok := event.Data.(map[string]any)
		if !ok {
			reject(errcode.ErrInvalidParam)
			return
		}
		channelID, ok := parseChannelID(data["channelId"])
		if !ok {
			reject(errcode.ErrInvalidParam)
			return
		}
		channel, ok := validateChannel(channelID, model.PermConnectVoice)
		if !ok {
			return
		}
		if channel.Type != model.ChannelTypeVoice {
			reject(errcode.ErrBadRequest.WithMessage("该频道不是语音频道"))
			return
		}
		if !c.IsSubscribed(channelID) {
			reject(errcode.ErrBadRequest.WithMessage("未订阅该频道"))
			return
		}

		// 广播 voice_user_left 到该频道（排除自己）
		c.Hub.BroadcastToChannelExcept(channelID, EventVoiceUserLeft, map[string]any{
			"channelId": channelID,
			"userId":    c.UserID,
		}, c.ID)

	case EventTyping:
		// 输入状态仅允许发送到已授权且已订阅的频道。
		// data: { channelId: number, isTyping: boolean }
		data, ok := event.Data.(map[string]any)
		if !ok {
			reject(errcode.ErrInvalidParam)
			return
		}
		channelID, ok := parseChannelID(data["channelId"])
		if !ok {
			reject(errcode.ErrInvalidParam)
			return
		}
		if _, ok = validateChannel(channelID, model.PermViewChannel); !ok {
			return
		}
		if !c.IsSubscribed(channelID) {
			reject(errcode.ErrForbidden.WithMessage("未订阅该频道"))
			return
		}
		isTyping, ok := data["isTyping"].(bool)
		if !ok {
			reject(errcode.ErrInvalidParam)
			return
		}

		// 广播到频道（排除发送者）
		c.Hub.BroadcastToChannelExcept(channelID, EventTyping, map[string]any{
			"channelId": channelID,
			"userId":    c.UserID,
			"username":  c.Username,
			"isTyping":  isTyping,
		}, c.ID)

	case EventScreenShareStart:
		// 用户开始屏幕共享。屏幕共享使用语音频道的订阅通道。
		// data: { channelId: number }
		data, ok := event.Data.(map[string]any)
		if !ok {
			reject(errcode.ErrInvalidParam)
			return
		}
		channelID, ok := parseChannelID(data["channelId"])
		if !ok {
			reject(errcode.ErrInvalidParam)
			return
		}
		channel, ok := validateChannel(channelID, model.PermViewChannel)
		if !ok {
			return
		}
		if channel.Type != model.ChannelTypeVoice {
			reject(errcode.ErrBadRequest.WithMessage("该频道不是语音频道"))
			return
		}
		if _, ok = validateChannel(channelID, model.PermScreenShare); !ok {
			return
		}
		if !c.IsSubscribed(channelID) {
			reject(errcode.ErrForbidden.WithMessage("未订阅该频道"))
			return
		}

		// 广播 screen_share_start 到频道（排除自己）
		c.Hub.BroadcastToChannelExcept(channelID, EventScreenShareStart, map[string]any{
			"channelId": channelID,
			"userId":    c.UserID,
			"username":  c.Username,
		}, c.ID)

	case EventScreenShareStop:
		// 用户停止屏幕共享。校验与开始共享相同，避免绕过频道权限。
		// data: { channelId: number }
		data, ok := event.Data.(map[string]any)
		if !ok {
			reject(errcode.ErrInvalidParam)
			return
		}
		channelID, ok := parseChannelID(data["channelId"])
		if !ok {
			reject(errcode.ErrInvalidParam)
			return
		}
		channel, ok := validateChannel(channelID, model.PermViewChannel)
		if !ok {
			return
		}
		if channel.Type != model.ChannelTypeVoice {
			reject(errcode.ErrBadRequest.WithMessage("该频道不是语音频道"))
			return
		}
		if _, ok = validateChannel(channelID, model.PermScreenShare); !ok {
			return
		}
		if !c.IsSubscribed(channelID) {
			reject(errcode.ErrForbidden.WithMessage("未订阅该频道"))
			return
		}

		// 广播 screen_share_stop 到频道（排除自己）
		c.Hub.BroadcastToChannelExcept(channelID, EventScreenShareStop, map[string]any{
			"channelId": channelID,
			"userId":    c.UserID,
		}, c.ID)

	case EventPing:
		// 应用层心跳：回送 pong 给发送者
		c.Hub.SendToUser(c.UserID, EventPong, map[string]any{
			"timestamp": time.Now().Unix(),
		})

	case EventFriendRequest:
		// 客户端确认收到好友请求通知，无需额外处理
		// 可用于未来扩展：标记通知已送达

	case EventFriendRequestResult:
		// 客户端确认收到请求结果通知
		// 可用于未来扩展：标记通知已送达
	}
}

// generateClientID 生成客户端ID
func generateClientID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

// parseChannelID 从 any 值解析频道ID
// JSON数字默认解析为 float64，需要拒绝负数、非有限数和溢出值。
func parseChannelID(v any) (uint64, bool) {
	switch val := v.(type) {
	case float64:
		if val <= 0 || math.IsNaN(val) || math.IsInf(val, 0) || val != math.Trunc(val) || val >= math.Pow(2, 64) {
			return 0, false
		}
		return uint64(val), true
	case int:
		if val <= 0 {
			return 0, false
		}
		return uint64(val), true
	case int64:
		if val <= 0 {
			return 0, false
		}
		return uint64(val), true
	case uint64:
		if val == 0 {
			return 0, false
		}
		return val, true
	case uint:
		if val == 0 {
			return 0, false
		}
		return uint64(val), true
	default:
		return 0, false
	}
}
