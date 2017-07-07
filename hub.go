package main

import (
	"container/list"
	"fmt"
)

type MsgBufferChange struct {
	Sender int    `json:"-"`
	ID     int    `json:"id"`
	Value  string `json:"value"`
}

type Client struct {
	ID         int
	Input      chan interface{}
	Output     chan interface{}
	Unregister func()
}

// Hub manages connected clients and takes care of routing messages
type Hub struct {
	register   chan *Client
	unregister chan *list.Element
	clients    map[int]*list.List
	broadcast  chan interface{}
}

func newHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *list.Element),
		broadcast:  make(chan interface{}),
		clients:    make(map[int]*list.List),
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

func (h *Hub) SendToUser(id int, msg interface{}) {
	fmt.Println("sending...", id)
	if l, ok := h.clients[id]; ok {
		for c := l.Front(); c != nil; c = c.Next() {
			c.Value.(*Client).Input <- msg
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
				h.clients[c.ID] = list.New()
			}

			el := h.clients[c.ID].PushFront(c)
			c.Unregister = func() {
				h.unregister <- el
			}
			fmt.Println("registered", c.ID, h.clients)

		case e := <-h.unregister:
			id := e.Value.(*Client).ID
			if l, ok := h.clients[id]; ok {
				l.Remove(e)
				if l.Len() == 0 {
					delete(h.clients, id)
				}
			}
			fmt.Println("unregistered", id, h.clients)
		}
	}
}
