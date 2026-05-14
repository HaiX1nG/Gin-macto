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
	validator := middleware.GetOriginValidator()
	if validator != nil {
		return validator.CheckOriginFunc()
	}
	// 验证器未初始化时的安全降级策略：仅允许无Origin头的同源请求
	// 浏览器同源请求不会携带Origin头，因此这类请求可以安全放行
	return func(r *http.Request) bool {
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
func (h *Handler) HandleWebSocket(c *gin.Context) {
	// 从查询参数获取token和roomID
	tokenString := c.Query("token")
	roomIDStr := c.Query("room_id")

	if tokenString == "" || roomIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 40003, "message": "缺少认证信息"})
		return
	}

	// 验证token
	claims, err := jwt.ParseAccessToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 40003, "message": "Token无效"})
		return
	}

	// 解析房间ID
	roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40001, "message": "房间ID无效"})
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

	// 创建客户端
	client := &Client{
		ID:             generateClientID(),
		UserID:         claims.UserID,
		Username:       claims.Username,
		RoomID:         roomID,
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

	conn := c.Conn.(*websocket.Conn)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		// 收到Pong响应，更新活跃时间并重置读超时
		c.UpdateLastActiveTime()
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := conn.ReadMessage()
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
		conn := c.Conn.(*websocket.Conn)
		conn.Close()
	}()

	conn := c.Conn.(*websocket.Conn)

	for {
		select {
		case message, ok := <-c.Send:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := conn.NextWriter(websocket.TextMessage)
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
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
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

	case EventChatMessage:
		// XSS过滤：对聊天消息内容进行HTML转义
		// event.Data 可能是 string 或 map[string]any
		// 如果是map，需要递归过滤其中的字符串字段
		sanitizedData := sanitizeEventData(event.Data)
		c.Hub.Broadcast(c.RoomID, EventChatMessage, map[string]any{
			"senderId":   c.UserID,
			"senderName": c.Username,
			"data":       sanitizedData,
		})

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

		// 如果指定了目标用户，只发给该用户
		if targetID > 0 {
			c.Hub.SendToUser(c.RoomID, uint64(targetID), EventWebRTCSignal, signalData)
		} else {
			// 广播给房间其他成员
			c.Hub.BroadcastExcept(c.RoomID, EventWebRTCSignal, signalData, c.ID)
		}

	case EventWebRTCOffer, EventWebRTCAnswer, EventWebRTCIceCandidate:
		// 兼容旧格式：WebRTC信令，广播给房间其他成员
		c.Hub.BroadcastExcept(c.RoomID, event.Event, map[string]any{
			"fromUserId":   c.UserID,
			"fromUsername": c.Username,
			"data":         event.Data,
		}, c.ID)

	case EventVoiceJoin:
		c.Hub.Broadcast(c.RoomID, EventVoiceJoin, map[string]any{
			"userId":   c.UserID,
			"username": c.Username,
		})

	case EventVoiceLeave:
		c.Hub.Broadcast(c.RoomID, EventVoiceLeave, map[string]any{
			"userId":   c.UserID,
			"username": c.Username,
		})

	case EventScreenShareStart:
		// 用户开始屏幕共享（前端发送）
		c.Hub.BroadcastExcept(c.RoomID, EventScreenShareStarted, map[string]any{
			"userId":   c.UserID,
			"username": c.Username,
		}, c.ID)

	case EventScreenShareStop:
		// 用户停止屏幕共享（前端发送）
		c.Hub.BroadcastExcept(c.RoomID, EventScreenShareStopped, map[string]any{
			"userId":   c.UserID,
			"username": c.Username,
		}, c.ID)

	case EventAudioShareStart:
		// 用户开始音频分享（前端发送）
		c.Hub.BroadcastExcept(c.RoomID, EventAudioShareStarted, map[string]any{
			"userId":   c.UserID,
			"username": c.Username,
		}, c.ID)

	case EventAudioShareStop:
		// 用户停止音频分享（前端发送）
		c.Hub.BroadcastExcept(c.RoomID, EventAudioShareStopped, map[string]any{
			"userId":   c.UserID,
			"username": c.Username,
		}, c.ID)
	}
}

// generateClientID 生成客户端ID
func generateClientID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
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
