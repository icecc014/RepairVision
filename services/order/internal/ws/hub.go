package ws

import (
	"encoding/json"
	"log"
	"sync"

	"golang.org/x/net/websocket"
)

type OrderEvent struct {
	Type       string `json:"type"`
	OrderId    int64  `json:"orderId"`
	OrderNo    string `json:"orderNo"`
	BuildingId int64  `json:"buildingId"`
	WorkerId   int64  `json:"workerId,omitempty"`
	Status     int64  `json:"status"`
}

type client struct {
	conn *websocket.Conn
	send chan []byte
	keys []string
}

type Hub struct {
	mu      sync.Mutex
	clients map[*client]struct{}
	byKey   map[string]map[*client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*client]struct{}),
		byKey:   make(map[string]map[*client]struct{}),
	}
}

func (h *Hub) register(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = struct{}{}
	for _, key := range c.keys {
		if h.byKey[key] == nil {
			h.byKey[key] = make(map[*client]struct{})
		}
		h.byKey[key][c] = struct{}{}
	}
}

func (h *Hub) unregister(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, c)
	for _, key := range c.keys {
		if set := h.byKey[key]; set != nil {
			delete(set, c)
			if len(set) == 0 {
				delete(h.byKey, key)
			}
		}
	}
	close(c.send)
}

func (h *Hub) PublishOrder(event OrderEvent) {
	body, _ := json.Marshal(event)
	keys := []string{"all", "building:" + intToString(event.BuildingId)}
	if event.WorkerId > 0 {
		keys = append(keys, "worker:"+intToString(event.WorkerId))
	}
	h.mu.Lock()
	seen := make(map[*client]struct{})
	for _, key := range keys {
		for c := range h.byKey[key] {
			seen[c] = struct{}{}
		}
	}
	h.mu.Unlock()
	for c := range seen {
		select {
		case c.send <- body:
		default:
		}
	}
}

func (h *Hub) pump(c *client) {
	for msg := range c.send {
		if err := writeMessage(c, msg); err != nil {
			log.Printf("ws write error: %v", err)
			return
		}
	}
}

func intToString(v int64) string {
	if v == 0 {
		return "0"
	}
	negative := v < 0
	u := uint64(v)
	if negative {
		u = uint64(-v)
	}
	var buf [20]byte
	i := len(buf)
	for u >= 10 {
		i--
		buf[i] = byte('0' + u%10)
		u /= 10
	}
	i--
	buf[i] = byte('0' + u)
	if negative {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func writeMessage(c *client, body []byte) error {
	return websocket.Message.Send(c.conn, string(body))
}
