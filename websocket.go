package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func (a *App) registerWebsocketHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/socket", a.handleWebsocket)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func readPump(conn *websocket.Conn, ch chan MsgBufferChange) {
	for {
		var msg MsgBufferChange
		err := conn.ReadJSON(&msg)

		if err != nil {
			close(ch)
			return
		}

		ch <- msg
	}
}

type JSON map[string]interface{}

func (a *App) handleWebsocket(w http.ResponseWriter, r *http.Request) {
	u, err := a.requireUser(w, r)
	if err != nil {
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	defer conn.Close()
	if err != nil {
		log.Println(err)
		return
	}

	client := a.Hub.RegisterClient(u.ID)
	defer a.Hub.UnregisterClient(client)
	readChannel := make(chan MsgBufferChange)
	go readPump(conn, readChannel)

	for {
		select {
		case msg, more := <-readChannel:
			if !more {
				return
			}
			msg.Sender = u.ID
			client.Output <- msg

		case msg := <-client.Input:
			conn.WriteJSON(msg)
		}
	}
}
