package hub

import (
    "encoding/json"
    "log"
    "sync"

    "github.com/sankhadeeproycbowdhury/go_monitor/internal/models"
	"github.com/sankhadeeproycbowdhury/go_monitor/internal/constants"
)

// Hub maintains the set of active clients and broadcasts messages.
type Hub struct {
    mu      sync.RWMutex
    clients map[*Client]struct{}
}

func New() *Hub {
    return &Hub{clients: make(map[*Client]struct{})}
}

func (h *Hub) Register(c *Client) {
    h.mu.Lock()
    h.clients[c] = struct{}{}
    h.mu.Unlock()
    log.Printf("hub: client connected (%d total)", h.Count())
}

func (h *Hub) Unregister(c *Client) {
    h.mu.Lock()
    delete(h.clients, c)
    h.mu.Unlock()
    log.Printf("hub: client disconnected (%d total)", h.Count())
}

func (h *Hub) Count() int {
    h.mu.RLock()
    defer h.mu.RUnlock()
    return len(h.clients)
}

// BroadcastMetrics wraps a snapshot in a WSMessage envelope and fans it out.
func (h *Hub) BroadcastMetrics(snap models.MetricsSnapshot) {
    h.broadcast(models.WSMessage{Type: constants.MsgMetrics, Payload: snap})
}

// BroadcastScaleEvent fans out a scaling notification.
func (h *Hub) BroadcastScaleEvent(ev models.ScaleEvent) {
    h.broadcast(models.WSMessage{Type: constants.MsgScaleEvent, Payload: ev})
}

func (h *Hub) broadcast(msg models.WSMessage) {
    data, err := json.Marshal(msg)
    if err != nil {
        log.Printf("hub: marshal error: %v", err)
        return
    }

    h.mu.RLock()
    clients := make([]*Client, 0, len(h.clients))
    for c := range h.clients {
        clients = append(clients, c)
    }
    h.mu.RUnlock()

    for _, c := range clients {
        // non-blocking send; if the client's buffer is full it'll be dropped
        select {
        case c.send <- data:
        default:
            log.Printf("hub: client send buffer full, dropping")
        }
    }
}