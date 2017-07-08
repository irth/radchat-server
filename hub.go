package main

import (
	"fmt"

	"github.com/irth/radchat-server/models"
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
		ID:    id,
		Input: make(chan interface{}),
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

func (h *Hub) IsConnected(id int) bool {
	_, ok := h.clients[id]
	return ok
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			if _, ok := h.clients[c.ID]; !ok {
				h.clients[c.ID] = make(map[*Client]bool)
			}

			h.clients[c.ID][c] = true
			fmt.Println("registered", c.ID, h.clients)
			go h.SendToFriendsWithUser(
				c.ID,
				func(u *models.User) interface{} {
					return JSON{"type": "statusUpdate", "id": u.ID, "status": u.Status}
				},
			)

		case c := <-h.unregister:
			if _, ok := h.clients[c.ID]; ok {
				delete(h.clients[c.ID], c)
				if len(h.clients[c.ID]) == 0 {
					delete(h.clients, c.ID)
				}
			}
			fmt.Println("deregistered", c.ID, h.clients)
			go h.SendToFriendsWithUser(
				c.ID,
				func(u *models.User) interface{} {
					return JSON{"type": "statusUpdate", "id": u.ID, "status": models.StatusUnavailable}
				},
			)
			go h.SendToFriends(c.ID, JSON{"type": "inputBufferUpdate", "id": c.ID, "value": ""})
		}
	}
}
