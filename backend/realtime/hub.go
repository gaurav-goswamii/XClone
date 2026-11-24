package realtime 

import ( 
	"sync"
    "backend/models"
)

type Hub struct {
	//lock that protects shared data.
	mu sync.RWMutex
	//stores websocket clients grouped by chat id
	byChat map[string]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		byChat: make(map[string]map[*Client]struct{}),
	}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.byChat[client.ChatID] == nil {
		h.byChat[client.ChatID] = make(map[*Client]struct{})
	}

	h.byChat[client.ChatID][client] = struct{}{}
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
    defer h.mu.Unlock()

	clients := h.byChat[client.ChatID]
	if _, exists := clients[client]; !exists {
		return
	}

	delete(clients, client)
	if len(clients) == 0 {
		delete(h.byChat, client.ChatID)
	}
	client.shutdown()
}

func (h *Hub) Broadcast(chatID string, message models.Message) {
	var slowClients []*Client

	h.mu.RLock()
	for client := range h.byChat[chatID] {
		select {
		case client.Outbound <- message:
		default:
			slowClients = append(slowClients, client)
		}
	}
	h.mu.RUnlock()

	for _, client := range slowClients {
		h.Unregister(client)
	}
}

func (h *Hub) Publish(message models.Message) {
	h.Broadcast(message.ChatID, message)
}