package main

import (
	"fmt"
)

type MsgBufferChange struct {
	Sender int    `json:"-"`
	ID     int    `json:"id"`
	Value  string `json:"value"`
}

type Client struct {
	ID     int
	Input  chan interface{}
	Output chan interface{}
}

// Hub manages connected clients and takes care of routing messages
type Hub struct {
	register   chan *Client
	unregister chan *Client
	clients    map[int]map[*Client]bool
	broadcast  chan interface{}
}

func newHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan interface{}),
		clients:    make(map[int]map[*Client]bool),
	}
}

func (h *Hub) RegisterClient(id int) *Client {
	c := &Client{
		ID:     id,
		Input:  make(chan interface{}),
		Output: h.broadcast,
	}
	h.register <- c
	return c
}

func (h *Hub) UnregisterClient(c *Client) {
	h.unregister <- c
}

func (h *Hub) SendToUser(id int, msg interface{}) {
	fmt.Println("sending...", id)
	if l, ok := h.clients[id]; ok {
		for c := range l {
			c.Input <- msg
		}
	}
}

func (h *Hub) Run() {
	for {
		select {
		case msg := <-h.broadcast:
			switch msg := msg.(type) {
			case MsgBufferChange:
				fmt.Println(msg.Value)
				h.SendToUser(msg.ID, MsgBufferChange{
					ID:    msg.Sender,
					Value: msg.Value,
				})
			}

		case c := <-h.register:
			if _, ok := h.clients[c.ID]; !ok {
				h.clients[c.ID] = make(map[*Client]bool)
			}

			h.clients[c.ID][c] = true
			fmt.Println("registered", c.ID, h.clients)

		case c := <-h.unregister:
			if _, ok := h.clients[c.ID]; ok {
				delete(h.clients[c.ID], c)
				if len(h.clients[c.ID]) == 0 {
					delete(h.clients, c.ID)
				}
			}
			fmt.Println("deregistered", c.ID, h.clients)
		}
	}
}
