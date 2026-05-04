package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/yourorg/livemix/pkg/jwt"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源
	},
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
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	// 创建客户端
	client := &Client{
		ID:       generateClientID(),
		UserID:   claims.UserID,
		Username: claims.Username,
		RoomID:   roomID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Hub:      h.hub,
	}

	// 注册客户端
	h.hub.register <- client

	// 启动读写协程
	go client.writePump()
	go client.readPump()
}

// readPump 读取客户端消息
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
	}()

	conn := c.Conn.(*websocket.Conn)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
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
func (c *Client) handleEvent(event Event) {
	switch event.Event {
	case EventChatMessage:
		// 广播聊天消息
		c.Hub.Broadcast(c.RoomID, EventChatMessage, map[string]interface{}{
			"senderId":   c.UserID,
			"senderName": c.Username,
			"data":       event.Data,
		})

	case EventWebRTCSignal:
		// 统一的WebRTC信令处理
		// 期望格式: { type: "offer|answer|ice-candidate", targetId?: number, payload: any }
		data, ok := event.Data.(map[string]interface{})
		if !ok {
			return
		}

		signalType, _ := data["type"].(string)
		targetID, _ := data["targetId"].(float64) // JSON数字默认解析为float64

		signalData := map[string]interface{}{
			"fromUserId":   c.UserID,
			"fromUsername": c.Username,
			"signal": map[string]interface{}{
				"type":    signalType,
				"payload": data["payload"],
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
		c.Hub.BroadcastExcept(c.RoomID, event.Event, map[string]interface{}{
			"fromUserId":   c.UserID,
			"fromUsername": c.Username,
			"data":         event.Data,
		}, c.ID)

	case EventVoiceJoin:
		c.Hub.Broadcast(c.RoomID, EventVoiceJoin, map[string]interface{}{
			"userId":   c.UserID,
			"username": c.Username,
		})

	case EventVoiceLeave:
		c.Hub.Broadcast(c.RoomID, EventVoiceLeave, map[string]interface{}{
			"userId":   c.UserID,
			"username": c.Username,
		})

	case EventScreenShareStart:
		// 屏幕共享开始
		c.Hub.BroadcastExcept(c.RoomID, EventScreenShareStart, map[string]interface{}{
			"userId":   c.UserID,
			"username": c.Username,
		}, c.ID)

	case EventScreenShareStop:
		// 屏幕共享停止
		c.Hub.BroadcastExcept(c.RoomID, EventScreenShareStop, map[string]interface{}{
			"userId":   c.UserID,
			"username": c.Username,
		}, c.ID)
	}
}

// generateClientID 生成客户端ID
func generateClientID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
