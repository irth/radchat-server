package main

import "fmt"

type MsgBufferChange struct {
	Sender int    `json:"-"`
	ID     int    `json:"id"`
	Value  string `json:"value"`
}

type Client struct {
	Input  chan MsgBufferChange
	Output chan MsgBufferChange
}

// Hub manages connected clients and takes care of routing messages
type Hub struct {
	clients   map[int]*Client
	broadcast chan MsgBufferChange
}

func newHub() *Hub {
	return &Hub{
		broadcast: make(chan MsgBufferChange),
		clients:   make(map[int]*Client),
	}
}

func (h *Hub) RegisterClient(id int) *Client {
	c := &Client{
		Input:  make(chan MsgBufferChange),
		Output: h.broadcast,
	}

	h.clients[id] = c
	fmt.Println("registered", id, h.clients)
	return c
}

func (h *Hub) UnregisterClient(id int) {
	if _, ok := h.clients[id]; ok {
		delete(h.clients, id)
	}
	fmt.Println("unregistered", id, h.clients)
}

func (h *Hub) Run() {
	for {
		msg := <-h.broadcast
		if c, ok := h.clients[msg.ID]; ok {
			c.Input <- MsgBufferChange{
				ID:    msg.Sender,
				Value: msg.Value,
			}
		}
	}
}
