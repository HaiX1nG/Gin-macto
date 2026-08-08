package ws

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yourorg/livemix/internal/service"
)

// EventType WebSocket事件类型
type EventType string

const (
	// 频道订阅
	EventJoinChannel  EventType = "join_channel"
	EventLeaveChannel EventType = "leave_channel"

	// 聊天
	EventChatMessage    EventType = "chat_message"
	EventMessageDelete  EventType = "message_delete"
	EventMessageUpdate  EventType = "message_update"
	EventReactionAdd    EventType = "reaction_add"
	EventReactionRemove EventType = "reaction_remove"
	EventTyping         EventType = "typing"

	// 语音
	EventVoiceJoin        EventType = "voice_join"
	EventVoiceLeave       EventType = "voice_leave"
	EventVoiceUserJoined  EventType = "voice_user_joined"
	EventVoiceUserLeft    EventType = "voice_user_left"
	EventVoiceStateUpdate EventType = "voice_state_update"

	// 屏幕共享
	EventScreenShareStart EventType = "screen_share_start"
	EventScreenShareStop  EventType = "screen_share_stop"

	// WebRTC
	EventWebRTCSignal EventType = "webrtc_signal"

	// 成员
	EventMemberJoined      EventType = "member_joined"
	EventMemberLeft        EventType = "member_left"
	EventParticipantUpdate EventType = "participant_update"

	// 用户状态
	EventUserOnline  EventType = "user_online"
	EventUserOffline EventType = "user_offline"

	// 心跳
	EventHeartbeat EventType = "heartbeat"
	EventPing      EventType = "ping"
	EventPong      EventType = "pong"
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
// per-app 单例连接，不绑定任何频道。通过 join_channel/leave_channel 订阅多个频道。
type Client struct {
	ID       string
	UserID   uint64
	Username string
	// Channels 该客户端订阅的频道集合
	Channels map[uint64]bool
	// Conn WebSocket连接（强类型）
	Conn *websocket.Conn
	Send chan []byte
	Hub  *Hub
	mu   sync.Mutex
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

// Hub WebSocket Hub，管理所有频道订阅和连接
type Hub struct {
	channels     map[uint64]*ChannelGroup
	clients      map[string]*Client
	register     chan *Client
	unregister   chan *Client
	broadcast    chan *BroadcastMessage
	userService  *service.UserService
	mu           sync.RWMutex
	// heartbeatTimeout 心跳超时时间，超过此时间未收到客户端响应则断开连接
	heartbeatTimeout time.Duration
}

// ChannelGroup 频道订阅组（替代旧的 Room 类型）
type ChannelGroup struct {
	ID      uint64
	Clients map[string]*Client
	mu      sync.RWMutex
}

// BroadcastMessage 广播消息
type BroadcastMessage struct {
	ChannelID uint64
	Event     EventType
	Data      any
	Exclude   string // 排除的客户端ID
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
		channels:         make(map[uint64]*ChannelGroup),
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
			h.broadcastToChannel(msg)

		case <-ticker.C:
			// 心跳检测
			h.checkHeartbeat()
		}
	}
}

// registerClient 注册客户端
// 仅加入全局 clients map，不自动加入任何频道，不广播 participant_update
// 客户端需主动发送 join_channel 事件订阅频道
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 添加到全局客户端列表
	h.clients[client.ID] = client

	// 设置用户在线状态
	if h.userService != nil {
		go h.userService.SetOnline(context.Background(), client.UserID)
	}
}

// unregisterClient 注销客户端
// 从全局 clients 移除 + 从其订阅的所有 channels 移除
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()

	// 从全局列表移除
	delete(h.clients, client.ID)

	// 设置用户离线状态
	if h.userService != nil {
		go h.userService.SetOffline(context.Background(), client.UserID)
	}

	// 从客户端订阅的所有频道中移除
	client.mu.Lock()
	subscribedChannels := make([]uint64, 0, len(client.Channels))
	for chID := range client.Channels {
		subscribedChannels = append(subscribedChannels, chID)
	}
	client.mu.Unlock()

	for _, chID := range subscribedChannels {
		if ch, exists := h.channels[chID]; exists {
			ch.mu.Lock()
			delete(ch.Clients, client.ID)
			empty := len(ch.Clients) == 0
			ch.mu.Unlock()

			// 频道组为空则删除
			if empty {
				delete(h.channels, chID)
			}
		}
	}

	h.mu.Unlock()

	close(client.Send)
}

// JoinChannel 将客户端加入频道订阅
// 若频道组不存在则创建
func (h *Hub) JoinChannel(client *Client, channelID uint64) {
	h.mu.Lock()

	// 创建频道组（如不存在）
	ch, exists := h.channels[channelID]
	if !exists {
		ch = &ChannelGroup{
			ID:      channelID,
			Clients: make(map[string]*Client),
		}
		h.channels[channelID] = ch
	}
	h.mu.Unlock()

	// 加入频道组
	ch.mu.Lock()
	ch.Clients[client.ID] = client
	ch.mu.Unlock()

	// 记录客户端订阅
	client.mu.Lock()
	if client.Channels == nil {
		client.Channels = make(map[uint64]bool)
	}
	client.Channels[channelID] = true
	client.mu.Unlock()
}

// LeaveChannel 从频道订阅移除客户端
// 频道组空则删除
func (h *Hub) LeaveChannel(client *Client, channelID uint64) {
	h.mu.Lock()

	ch, exists := h.channels[channelID]
	if !exists {
		h.mu.Unlock()
		// 频道组不存在，仅从客户端订阅记录中移除
		client.mu.Lock()
		delete(client.Channels, channelID)
		client.mu.Unlock()
		return
	}

	ch.mu.Lock()
	delete(ch.Clients, client.ID)
	empty := len(ch.Clients) == 0
	ch.mu.Unlock()

	// 频道组为空则删除
	if empty {
		delete(h.channels, channelID)
	}

	h.mu.Unlock()

	// 从客户端订阅记录中移除
	client.mu.Lock()
	delete(client.Channels, channelID)
	client.mu.Unlock()
}

// BroadcastToChannel 向频道所有订阅者广播消息
func (h *Hub) BroadcastToChannel(channelID uint64, event EventType, data any) {
	h.broadcast <- &BroadcastMessage{
		ChannelID: channelID,
		Event:     event,
		Data:      data,
	}
}

// BroadcastToChannelExcept 向频道所有订阅者广播消息（排除指定客户端）
func (h *Hub) BroadcastToChannelExcept(channelID uint64, event EventType, data any, excludeClientID string) {
	h.broadcast <- &BroadcastMessage{
		ChannelID: channelID,
		Event:     event,
		Data:      data,
		Exclude:   excludeClientID,
	}
}

// broadcastToChannel 向频道广播消息（内部方法，通过 broadcast 通道调用）
func (h *Hub) broadcastToChannel(msg *BroadcastMessage) {
	h.mu.RLock()
	ch, exists := h.channels[msg.ChannelID]
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

	ch.mu.RLock()
	defer ch.mu.RUnlock()

	for id, client := range ch.Clients {
		if id == msg.Exclude {
			continue
		}
		select {
		case client.Send <- data:
		default:
			// 发送失败，关闭连接
			close(client.Send)
			delete(ch.Clients, id)
		}
	}
}

// notifyVoiceParticipants 通知语音频道订阅者参与者列表更新
// 仅语音频道用，由 handler 在 voice_join/voice_leave 时调用
func (h *Hub) notifyVoiceParticipants(channelID uint64) {
	h.mu.RLock()
	ch, exists := h.channels[channelID]
	h.mu.RUnlock()

	if !exists {
		return
	}

	ch.mu.RLock()
	participants := make([]map[string]any, 0, len(ch.Clients))
	for _, client := range ch.Clients {
		participants = append(participants, map[string]any{
			"userId":   client.UserID,
			"username": client.Username,
		})
	}
	ch.mu.RUnlock()

	h.broadcast <- &BroadcastMessage{
		ChannelID: channelID,
		Event:     EventParticipantUpdate,
		Data: map[string]any{
			"channelId":    channelID,
			"participants": participants,
		},
	}
}

// SendToUser 向用户发送消息（遍历所有该用户的连接，不绑频道）
// 用户可能有多个连接，全部发送
func (h *Hub) SendToUser(userID uint64, event EventType, data any) {
	dataBytes, err := json.Marshal(Event{
		Event: event,
		Data:  data,
	})
	if err != nil {
		return
	}

	h.mu.RLock()
	for _, client := range h.clients {
		if client.UserID == userID {
			select {
			case client.Send <- dataBytes:
			default:
				// 发送失败
			}
		}
	}
	h.mu.RUnlock()
}

// GetChannelClients 获取频道订阅者列表
func (h *Hub) GetChannelClients(channelID uint64) []*Client {
	h.mu.RLock()
	ch, exists := h.channels[channelID]
	h.mu.RUnlock()

	if !exists {
		return nil
	}

	ch.mu.RLock()
	defer ch.mu.RUnlock()

	clients := make([]*Client, 0, len(ch.Clients))
	for _, client := range ch.Clients {
		clients = append(clients, client)
	}
	return clients
}

// GetClientByUserID 按 userID 查找客户端（取第一个连接）
func (h *Hub) GetClientByUserID(userID uint64) *Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		if client.UserID == userID {
			return client
		}
	}
	return nil
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
