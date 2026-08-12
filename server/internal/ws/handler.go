package ws

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/yourorg/livemix/internal/middleware"
	"github.com/yourorg/livemix/pkg/jwt"
	"github.com/yourorg/livemix/pkg/logger"
	"github.com/yourorg/livemix/pkg/util"
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
	hub *Hub
}

// NewHandler 创建WebSocket处理器实例
func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

// HandleWebSocket 处理WebSocket连接
// per-app 单例连接，不要求 room_id 参数
// 客户端连接后通过 join_channel/leave_channel 事件订阅频道
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

	// 创建客户端（不绑定任何频道，Channels 初始为空）
	client := &Client{
		ID:             generateClientID(),
		UserID:         claims.UserID,
		Username:       claims.Username,
		Channels:       make(map[uint64]bool),
		Conn:           conn,
		Send:           make(chan []byte, 256),
		Hub:            h.hub,
		lastActiveTime: time.Now(), // 初始化活跃时间
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
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
	}()

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

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 批量发送
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
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
	switch event.Event {
	case EventHeartbeat:
		// 心跳事件，仅更新活跃时间，无需响应
		// 活跃时间已在readPump中更新，此处无需额外处理
		// 客户端可发送空心跳事件保持连接活跃

	case EventJoinChannel:
		// 订阅频道
		// data: { channelId: number }
		data, ok := event.Data.(map[string]any)
		if !ok {
			return
		}
		channelID, ok := parseChannelID(data["channelId"])
		if !ok {
			return
		}

		c.Hub.JoinChannel(c, channelID)

		// 广播 member_joined 到该频道（排除自己）
		c.Hub.BroadcastToChannelExcept(channelID, EventMemberJoined, map[string]any{
			"channelId": channelID,
			"userId":    c.UserID,
			"username":  c.Username,
		}, c.ID)

	case EventLeaveChannel:
		// 取消订阅频道
		// data: { channelId: number }
		data, ok := event.Data.(map[string]any)
		if !ok {
			return
		}
		channelID, ok := parseChannelID(data["channelId"])
		if !ok {
			return
		}

		c.Hub.LeaveChannel(c, channelID)

		// 广播 member_left 到该频道（排除自己）
		c.Hub.BroadcastToChannelExcept(channelID, EventMemberLeft, map[string]any{
			"channelId": channelID,
			"userId":    c.UserID,
		}, c.ID)

	case EventChatMessage:
		// XSS过滤：对聊天消息内容进行HTML转义
		sanitizedData := sanitizeEventData(event.Data)

		// data 含 channelId，解析后定向广播到频道
		data, ok := event.Data.(map[string]any)
		if !ok {
			return
		}
		channelID, ok := parseChannelID(data["channelId"])
		if !ok {
			return
		}

		c.Hub.BroadcastToChannelExcept(channelID, EventChatMessage, map[string]any{
			"channelId":   channelID,
			"senderId":    c.UserID,
			"senderName":  c.Username,
			"data":        sanitizedData,
		}, c.ID)

	case EventWebRTCSignal:
		// 统一的WebRTC信令处理
		// 期望格式: { type: "offer|answer|ice-candidate", targetId?: number, payload: any }
		data, ok := event.Data.(map[string]any)
		if !ok {
			return
		}

		signalType, _ := data["type"].(string)
		targetID, _ := data["targetId"].(float64) // JSON数字默认解析为float64

		// 构建符合前端期望的格式
		signalData := map[string]any{
			"fromUserId":   c.UserID,
			"fromUsername": c.Username,
			"signal": map[string]any{
				"type":     signalType,
				"targetId": uint64(targetID),
				"payload":  data["payload"],
			},
		}

		// 如果指定了目标用户，全局发给该用户（不绑频道）
		if targetID > 0 {
			c.Hub.SendToUser(uint64(targetID), EventWebRTCSignal, signalData)
		}
		// targetId <= 0 时跳过（前端始终应指定 targetId）

	case EventVoiceJoin:
		// 加入语音频道
		// data: { channelId: number }
		data, ok := event.Data.(map[string]any)
		if !ok {
			return
		}
		channelID, ok := parseChannelID(data["channelId"])
		if !ok {
			return
		}

		// 广播 voice_user_joined 到该频道（排除自己）
		c.Hub.BroadcastToChannelExcept(channelID, EventVoiceUserJoined, map[string]any{
			"channelId": channelID,
			"userId":    c.UserID,
			"username":  c.Username,
		}, c.ID)

	case EventVoiceLeave:
		// 离开语音频道
		// data: { channelId: number }
		data, ok := event.Data.(map[string]any)
		if !ok {
			return
		}
		channelID, ok := parseChannelID(data["channelId"])
		if !ok {
			return
		}

		// 广播 voice_user_left 到该频道（排除自己）
		c.Hub.BroadcastToChannelExcept(channelID, EventVoiceUserLeft, map[string]any{
			"channelId": channelID,
			"userId":    c.UserID,
		}, c.ID)

	case EventTyping:
		// 输入状态
		// data: { channelId: number, isTyping: boolean }
		data, ok := event.Data.(map[string]any)
		if !ok {
			return
		}
		channelID, ok := parseChannelID(data["channelId"])
		if !ok {
			return
		}
		isTyping, _ := data["isTyping"].(bool)

		// 广播到频道（排除发送者）
		c.Hub.BroadcastToChannelExcept(channelID, EventTyping, map[string]any{
			"channelId": channelID,
			"userId":    c.UserID,
			"username":  c.Username,
			"isTyping":  isTyping,
		}, c.ID)

	case EventScreenShareStart:
		// 用户开始屏幕共享
		// data: { channelId: number }
		data, ok := event.Data.(map[string]any)
		if !ok {
			return
		}
		channelID, ok := parseChannelID(data["channelId"])
		if !ok {
			return
		}

		// 广播 screen_share_start 到频道（排除自己）
		c.Hub.BroadcastToChannelExcept(channelID, EventScreenShareStart, map[string]any{
			"channelId": channelID,
			"userId":    c.UserID,
			"username":  c.Username,
		}, c.ID)

	case EventScreenShareStop:
		// 用户停止屏幕共享
		// data: { channelId: number }
		data, ok := event.Data.(map[string]any)
		if !ok {
			return
		}
		channelID, ok := parseChannelID(data["channelId"])
		if !ok {
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
	}
}

// generateClientID 生成客户端ID
func generateClientID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

// parseChannelID 从 any 值解析频道ID
// JSON数字默认解析为 float64，需要安全转换
func parseChannelID(v any) (uint64, bool) {
	switch val := v.(type) {
	case float64:
		return uint64(val), true
	case int:
		return uint64(val), true
	case int64:
		return uint64(val), true
	case uint64:
		return val, true
	default:
		return 0, false
	}
}

// sanitizeEventData 对事件数据进行XSS过滤
// 递归处理map中的字符串字段，确保所有用户输入的内容都被HTML转义
func sanitizeEventData(data any) any {
	switch v := data.(type) {
	case string:
		// 对字符串直接进行HTML转义
		return util.TrimAndEscape(v)
	case map[string]any:
		// 对map递归处理每个值
		result := make(map[string]any)
		for key, value := range v {
			result[key] = sanitizeEventData(value)
		}
		return result
	case []any:
		// 对数组递归处理每个元素
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = sanitizeEventData(item)
		}
		return result
	default:
		// 其他类型（数字、布尔等）无需过滤
		return data
	}
}
