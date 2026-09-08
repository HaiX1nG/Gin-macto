package ws

import (
	"encoding/json"
	"testing"

	"github.com/yourorg/livemix/internal/dto"
	"github.com/yourorg/livemix/internal/middleware"
)

func testClient(id string, userID uint64) *Client {
	return &Client{
		ID:       id,
		UserID:   userID,
		Channels: make(map[uint64]bool),
		Send:     make(chan []byte, 8),
	}
}

func TestPublishChatMessageUsesCanonicalEnvelopeAndExcludesSenderUser(t *testing.T) {
	rateLimiter := middleware.NewWSRateLimiter(middleware.DefaultWSRateLimiterConfig())
	defer rateLimiter.Stop()
	hub := NewHub(nil, 0, 0, rateLimiter)
	senderConnection := testClient("sender-1", 10)
	senderSecondConnection := testClient("sender-2", 10)
	recipient := testClient("recipient", 20)

	for _, client := range []*Client{senderConnection, senderSecondConnection, recipient} {
		hub.registerClient(client)
		hub.JoinChannel(client, 7)
	}

	message := &dto.MessageResponse{
		ID:           99,
		ChannelID:    7,
		SenderUserID: 10,
		SenderName:   "sender",
		SenderAvatar: "avatar",
		Type:         1,
		Content:      "persisted",
		IsPinned:     false,
		Reactions:    []dto.ReactionResponse{},
		Attachments:  []dto.AttachmentResponse{},
		CreatedAt:    "2026-09-02 00:00:00",
	}

	hub.PublishChatMessage(7, 10, message)
	hub.broadcastToChannel(<-hub.broadcast)

	if got := len(senderConnection.Send); got != 0 {
		t.Fatalf("sender connection received %d messages", got)
	}
	if got := len(senderSecondConnection.Send); got != 0 {
		t.Fatalf("sender's second connection received %d messages", got)
	}
	if got := len(recipient.Send); got != 1 {
		t.Fatalf("recipient received %d messages, want 1", got)
	}

	var envelope struct {
		Event string `json:"event"`
		Data  struct {
			ChannelID uint64              `json:"channelId"`
			Message   dto.MessageResponse `json:"message"`
		} `json:"data"`
	}
	if err := json.Unmarshal(<-recipient.Send, &envelope); err != nil {
		t.Fatalf("decode broadcast: %v", err)
	}
	if envelope.Event != string(EventChatMessage) {
		t.Fatalf("event = %q, want %q", envelope.Event, EventChatMessage)
	}
	if envelope.Data.ChannelID != 7 || envelope.Data.Message.ID != message.ID {
		t.Fatalf("unexpected canonical data: %+v", envelope.Data)
	}
}

func TestLegacyChatIngressIsRejectedWithoutBroadcast(t *testing.T) {
	rateLimiter := middleware.NewWSRateLimiter(middleware.DefaultWSRateLimiterConfig())
	defer rateLimiter.Stop()
	hub := NewHub(nil, 0, 0, rateLimiter)
	client := testClient("sender", 10)
	client.Hub = hub
	hub.registerClient(client)

	client.handleEvent(Event{
		Event: EventChatMessage,
		Data: map[string]any{
			"channelId": float64(7),
			"content":   "legacy",
		},
	})

	if len(hub.broadcast) != 0 {
		t.Fatal("legacy chat ingress queued a broadcast")
	}
	if len(client.Send) != 1 {
		t.Fatalf("rejection response count = %d, want 1", len(client.Send))
	}
	var envelope struct {
		Event string `json:"event"`
	}
	if err := json.Unmarshal(<-client.Send, &envelope); err != nil {
		t.Fatalf("decode rejection: %v", err)
	}
	if envelope.Event != string(EventError) {
		t.Fatalf("rejection event = %q, want %q", envelope.Event, EventError)
	}
}
