package main

import "fmt"

type MsgStatusChange struct {
	Sender int    `json:"id"`
	Status string `json:"status"`
}

type MsgBufferChange struct {
	Sender int    `json:"-"`
	ID     int    `json:"id"`
	Value  string `json:"value"`
}

type Client struct {
	Input  chan interface{}
	Output chan interface{}
}

// Hub manages connected clients and takes care of routing messages
type Hub struct {
	clients   map[int]*Client
	broadcast chan interface{}
}

func newHub() *Hub {
	return &Hub{
		broadcast: make(chan interface{}),
		clients:   make(map[int]*Client),
	}
}

func (h *Hub) RegisterClient(id int) *Client {
	c := &Client{
		Input:  make(chan interface{}),
		Output: h.broadcast,
	}

	h.clients[id] = c
	fmt.Println("registered", id, h.clients)
	return c
}

func (h *Hub) SendToClient(id int, msg interface{}) {
	fmt.Println("sending...", id)
	if c, ok := h.clients[id]; ok {
		c.Input <- msg
	}
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
		switch msg := msg.(type) {
		case MsgBufferChange:
			if c, ok := h.clients[msg.ID]; ok {
				c.Input <- MsgBufferChange{
					ID:    msg.Sender,
					Value: msg.Value,
				}
			}
		}
	}
}
