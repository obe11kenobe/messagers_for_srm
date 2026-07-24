package messager

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"nhooyr.io/websocket"
)

type Hub struct {
	mu    sync.Mutex
	conns map[int][]*websocket.Conn
}

func NewHub() *Hub {
	return &Hub{conns: make(map[int][]*websocket.Conn)}
}

func (h *Hub) subscribe(convID int, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[convID] = append(h.conns[convID], c)
}

func (h *Hub) unsubscribe(convID int, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	conns := h.conns[convID]
	for i, conn := range conns {
		if conn == c {
			h.conns[convID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
}

func (h *Hub) broadcast(convID int, msg Message) {
	h.mu.Lock()
	conns := h.conns[convID]
	h.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("broadcast marshal:", err)
		return
	}

	for _, c := range conns {
		err := c.Write(context.Background(), websocket.MessageText, data)
		if err != nil {
			log.Println("broadcast write:", err)
			continue
		}
	}
}
