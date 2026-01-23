package matchmaking

import (
	"encoding/json"
	"sync"

	"github.com/leconcepteur/cloud-cicd-poc/internal/models"
)

type EventHub struct {
	clients map[string]chan []byte
	mu      sync.RWMutex
}

func NewEventHub() *EventHub {
	return &EventHub{
		clients: make(map[string]chan []byte),
	}
}

func (h *EventHub) Subscribe(userID string) chan []byte {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Close existing channel if any
	if existing, ok := h.clients[userID]; ok {
		close(existing)
	}

	ch := make(chan []byte, 10)
	h.clients[userID] = ch
	return ch
}

func (h *EventHub) Unsubscribe(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if ch, ok := h.clients[userID]; ok {
		close(ch)
		delete(h.clients, userID)
	}
}

func (h *EventHub) SendToUser(userID string, event models.SSEEvent) {
	h.mu.RLock()
	ch, ok := h.clients[userID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	select {
	case ch <- data:
	default:
		// Channel full, skip
	}
}

func (h *EventHub) SendToUsers(userIDs []string, event models.SSEEvent) {
	for _, userID := range userIDs {
		h.SendToUser(userID, event)
	}
}

func (h *EventHub) IsConnected(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}
