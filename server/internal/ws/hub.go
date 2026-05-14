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
	EventScreenShareStarted EventType = "screen_share_started"
	EventScreenShareStopped EventType = "screen_share_stopped"
	EventAudioShareStart    EventType = "audio_share_start"
	EventAudioShareStop     EventType = "audio_share_stop"
	EventAudioShareStarted  EventType = "audio_share_started"
	EventAudioShareStopped  EventType = "audio_share_stopped"
	EventChatMessage        EventType = "chat_message"
	EventParticipantUpdate  EventType = "participant_update"
	EventWebRTCOffer        EventType = "webrtc_offer"
	EventWebRTCAnswer       EventType = "webrtc_answer"
	EventWebRTCIceCandidate EventType = "webrtc_ice_candidate"
	EventWebRTCSignal       EventType = "webrtc_signal"
	EventUserOnline         EventType = "user_online"
	EventUserOffline        EventType = "user_offline"
	// EventHeartbeat 心跳事件，客户端可发送此事件更新活跃时间
	EventHeartbeat EventType = "heartbeat"
)

// Hub 相关常量
const (
	// DefaultHeartbeatTimeout 默认心跳超时时间，超过此时间未收到客户端响应则断开连接
	// 建议设置为前端心跳间隔的2-3倍，允许丢失1-2次心跳
	DefaultHeartbeatTimeout = 60 * time.Second
	// HeartbeatCheckInterval 心跳检测间隔，定期检查客户端活跃状态
	HeartbeatCheckInterval = 15 * time.Second
	// RegisterChannelBuffer 注册通道缓冲区大小
	RegisterChannelBuffer = 256
	// UnregisterChannelBuffer 注销通道缓冲区大小
	UnregisterChannelBuffer = 256
	// BroadcastChannelBuffer 广播通道缓冲区大小
	BroadcastChannelBuffer = 1024
)

// Event WebSocket事件结构
type Event struct {
	Event EventType `json:"event"`
	Data  any       `json:"data"`
}

// Client WebSocket客户端
type Client struct {
	ID       string
	UserID   uint64
	Username string
	RoomID   uint64
	Conn     any // 实际类型为 *websocket.Conn
	Send     chan []byte
	Hub      *Hub
	mu       sync.Mutex
	// lastActiveTime 最后活跃时间，用于心跳检测
	// 每次收到客户端消息或Pong响应时更新
	lastActiveTime time.Time
}

// UpdateLastActiveTime 更新最后活跃时间
// 线程安全，在收到客户端消息或Pong响应时调用
func (c *Client) UpdateLastActiveTime() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastActiveTime = time.Now()
}

// GetLastActiveTime 获取最后活跃时间
// 线程安全，用于心跳检测
func (c *Client) GetLastActiveTime() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lastActiveTime
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
	// heartbeatTimeout 心跳超时时间，超过此时间未收到客户端响应则断开连接
	heartbeatTimeout time.Duration
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
	Data    any
	Exclude string // 排除的客户端ID
}

// NewHub 创建Hub实例
// heartbeatTimeout 心跳超时时间，超过此时间未收到客户端响应则断开连接
// 建议设置为前端心跳间隔的2-3倍，允许丢失1-2次心跳
func NewHub(userService *service.UserService, heartbeatTimeout time.Duration) *Hub {
	// 设置默认心跳超时时间
	if heartbeatTimeout <= 0 {
		heartbeatTimeout = DefaultHeartbeatTimeout
	}

	return &Hub{
		rooms:            make(map[uint64]*Room),
		clients:          make(map[string]*Client),
		register:         make(chan *Client, RegisterChannelBuffer),
		unregister:       make(chan *Client, UnregisterChannelBuffer),
		broadcast:        make(chan *BroadcastMessage, BroadcastChannelBuffer),
		userService:      userService,
		heartbeatTimeout: heartbeatTimeout,
	}
}

// Run 运行Hub
func (h *Hub) Run() {
	ticker := time.NewTicker(HeartbeatCheckInterval)
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
	participants := make([]map[string]any, 0, len(room.Clients))
	for _, client := range room.Clients {
		participants = append(participants, map[string]any{
			"userId":   client.UserID,
			"username": client.Username,
		})
	}
	room.mu.RUnlock()

	h.broadcast <- &BroadcastMessage{
		RoomID: roomID,
		Event:  EventParticipantUpdate,
		Data: map[string]any{
			"participants": participants,
		},
	}
}

// checkHeartbeat 检查心跳，清理超时连接
// 实现机制：
// 1. 遍历所有客户端，检查最后活跃时间
// 2. 超过心跳超时时间的客户端视为断连，加入待清理列表
// 3. 异步清理超时连接，避免阻塞Hub主循环
//
// 性能考虑：
// - 使用快照遍历，避免长时间持有锁
// - 清理操作异步执行，不影响消息处理
// - 心跳检测间隔为 HeartbeatCheckInterval，与Run()中的ticker一致
func (h *Hub) checkHeartbeat() {
	// 获取客户端快照，避免长时间持有锁
	h.mu.RLock()
	clientCount := len(h.clients)
	if clientCount == 0 {
		h.mu.RUnlock()
		return
	}

	// 复制客户端列表，减少锁持有时间
	clients := make([]*Client, 0, clientCount)
	for _, client := range h.clients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	// 检查每个客户端的活跃时间
	now := time.Now()
	timeoutClients := make([]*Client, 0)

	for _, client := range clients {
		lastActive := client.GetLastActiveTime()
		// 超过心跳超时时间未活跃，视为断连
		if now.Sub(lastActive) > h.heartbeatTimeout {
			timeoutClients = append(timeoutClients, client)
		}
	}

	// 异步清理超时连接，避免阻塞Hub主循环
	// 清理操作会修改Hub状态，需要通过unregister通道进行
	for _, client := range timeoutClients {
		select {
		case h.unregister <- client:
			// 成功发送到注销通道，Hub会处理清理
		default:
			// 通道满，跳过本次清理，下次检测时再处理
			// 这种情况极少发生，因为通道容量为 UnregisterChannelBuffer
		}
	}
}

// Broadcast 广播消息
func (h *Hub) Broadcast(roomID uint64, event EventType, data any) {
	h.broadcast <- &BroadcastMessage{
		RoomID: roomID,
		Event:  event,
		Data:   data,
	}
}

// BroadcastExcept 广播消息（排除指定客户端）
func (h *Hub) BroadcastExcept(roomID uint64, event EventType, data any, excludeClientID string) {
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

// GetClientByUserID 根据用户ID获取客户端
func (h *Hub) GetClientByUserID(roomID, userID uint64) *Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	room, exists := h.rooms[roomID]
	if !exists {
		return nil
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	for _, client := range room.Clients {
		if client.UserID == userID {
			return client
		}
	}
	return nil
}

// SendToClient 向特定客户端发送消息
func (h *Hub) SendToClient(clientID string, event EventType, data any) {
	h.mu.RLock()
	client, exists := h.clients[clientID]
	h.mu.RUnlock()

	if !exists {
		return
	}

	dataBytes, err := json.Marshal(Event{
		Event: event,
		Data:  data,
	})
	if err != nil {
		return
	}

	select {
	case client.Send <- dataBytes:
	default:
		// 发送失败
	}
}

// SendToUser 向特定用户发送消息（在指定房间内）
func (h *Hub) SendToUser(roomID, userID uint64, event EventType, data any) {
	client := h.GetClientByUserID(roomID, userID)
	if client == nil {
		return
	}

	dataBytes, err := json.Marshal(Event{
		Event: event,
		Data:  data,
	})
	if err != nil {
		return
	}

	select {
	case client.Send <- dataBytes:
	default:
		// 发送失败
	}
}
