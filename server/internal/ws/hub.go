package ws

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/yourorg/livemix/internal/service"
)

// EventType WebSocket事件类型
type EventType string

const (
	EventVoiceJoin          EventType = "voice_join"
	EventVoiceLeave         EventType = "voice_leave"
	EventMusicSync          EventType = "music_sync"
	EventScreenShareStart   EventType = "screen_share_start"
	EventScreenShareStop    EventType = "screen_share_stop"
	EventChatMessage        EventType = "chat_message"
	EventParticipantUpdate  EventType = "participant_update"
	EventWebRTCOffer        EventType = "webrtc_offer"
	EventWebRTCAnswer       EventType = "webrtc_answer"
	EventWebRTCIceCandidate EventType = "webrtc_ice_candidate"
	EventUserOnline         EventType = "user_online"
	EventUserOffline        EventType = "user_offline"
)

// Event WebSocket事件结构
type Event struct {
	Event EventType   `json:"event"`
	Data  interface{} `json:"data"`
}

// Client WebSocket客户端
type Client struct {
	ID       string
	UserID   uint64
	Username string
	RoomID   uint64
	Conn     interface{} // 实际类型为 *websocket.Conn
	Send     chan []byte
	Hub      *Hub
	mu       sync.Mutex
}

// Hub WebSocket Hub，管理所有房间和连接
type Hub struct {
	rooms       map[uint64]*Room
	clients     map[string]*Client
	register    chan *Client
	unregister  chan *Client
	broadcast   chan *BroadcastMessage
	userService *service.UserService
	mu          sync.RWMutex
}

// Room 房间
type Room struct {
	ID      uint64
	Clients map[string]*Client
	mu      sync.RWMutex
}

// BroadcastMessage 广播消息
type BroadcastMessage struct {
	RoomID  uint64
	Event   EventType
	Data    interface{}
	Exclude string // 排除的客户端ID
}

// NewHub 创建Hub实例
func NewHub(userService *service.UserService) *Hub {
	return &Hub{
		rooms:       make(map[uint64]*Room),
		clients:     make(map[string]*Client),
		register:    make(chan *Client, 256),
		unregister:  make(chan *Client, 256),
		broadcast:   make(chan *BroadcastMessage, 1024),
		userService: userService,
	}
}

// Run 运行Hub
func (h *Hub) Run() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case msg := <-h.broadcast:
			h.broadcastToRoom(msg)

		case <-ticker.C:
			// 心跳检测
			h.checkHeartbeat()
		}
	}
}

// registerClient 注册客户端
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 添加到全局客户端列表
	h.clients[client.ID] = client

	// 添加到房间
	room, exists := h.rooms[client.RoomID]
	if !exists {
		room = &Room{
			ID:      client.RoomID,
			Clients: make(map[string]*Client),
		}
		h.rooms[client.RoomID] = room
	}

	room.mu.Lock()
	room.Clients[client.ID] = client
	room.mu.Unlock()

	// 设置用户在线状态
	if h.userService != nil {
		go h.userService.SetOnline(context.Background(), client.UserID)
	}

	// 广播用户加入事件
	h.notifyParticipantUpdate(client.RoomID)
}

// unregisterClient 注销客户端
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 从全局列表移除
	delete(h.clients, client.ID)

	// 设置用户离线状态
	if h.userService != nil {
		go h.userService.SetOffline(context.Background(), client.UserID)
	}

	// 从房间移除
	if room, exists := h.rooms[client.RoomID]; exists {
		room.mu.Lock()
		delete(room.Clients, client.ID)
		room.mu.Unlock()

		// 如果房间为空，删除房间
		if len(room.Clients) == 0 {
			delete(h.rooms, client.RoomID)
		} else {
			// 广播用户离开事件
			h.notifyParticipantUpdate(client.RoomID)
		}
	}

	close(client.Send)
}

// broadcastToRoom 向房间广播消息
func (h *Hub) broadcastToRoom(msg *BroadcastMessage) {
	h.mu.RLock()
	room, exists := h.rooms[msg.RoomID]
	h.mu.RUnlock()

	if !exists {
		return
	}

	data, err := json.Marshal(Event{
		Event: msg.Event,
		Data:  msg.Data,
	})
	if err != nil {
		return
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	for id, client := range room.Clients {
		if id == msg.Exclude {
			continue
		}
		select {
		case client.Send <- data:
		default:
			// 发送失败，关闭连接
			close(client.Send)
			delete(room.Clients, id)
		}
	}
}

// notifyParticipantUpdate 通知参与者更新
func (h *Hub) notifyParticipantUpdate(roomID uint64) {
	room, exists := h.rooms[roomID]
	if !exists {
		return
	}

	room.mu.RLock()
	participants := make([]map[string]interface{}, 0, len(room.Clients))
	for _, client := range room.Clients {
		participants = append(participants, map[string]interface{}{
			"userId":   client.UserID,
			"username": client.Username,
		})
	}
	room.mu.RUnlock()

	h.broadcast <- &BroadcastMessage{
		RoomID: roomID,
		Event:  EventParticipantUpdate,
		Data: map[string]interface{}{
			"participants": participants,
		},
	}
}

// checkHeartbeat 检查心跳
func (h *Hub) checkHeartbeat() {
	// 简化实现，实际应该在连接层面处理心跳
}

// Broadcast 广播消息
func (h *Hub) Broadcast(roomID uint64, event EventType, data interface{}) {
	h.broadcast <- &BroadcastMessage{
		RoomID: roomID,
		Event:  event,
		Data:   data,
	}
}

// BroadcastExcept 广播消息（排除指定客户端）
func (h *Hub) BroadcastExcept(roomID uint64, event EventType, data interface{}, excludeClientID string) {
	h.broadcast <- &BroadcastMessage{
		RoomID:  roomID,
		Event:   event,
		Data:    data,
		Exclude: excludeClientID,
	}
}

// GetRoomClients 获取房间客户端列表
func (h *Hub) GetRoomClients(roomID uint64) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	room, exists := h.rooms[roomID]
	if !exists {
		return nil
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	clients := make([]*Client, 0, len(room.Clients))
	for _, client := range room.Clients {
		clients = append(clients, client)
	}
	return clients
}
