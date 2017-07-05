package main

import (
	"fmt"
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

type Message struct {
	Type string `json:"type"`
	*AuthTokenRequest
}

type JSON map[string]interface{}

func (a *App) handleWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	defer conn.Close()
	if err != nil {
		log.Println(err)
		return
	}

	for {
		var msg Message
		err := conn.ReadJSON(&msg)

		if err != nil {
			return
		}
		switch msg.Type {
		case "auth":
			_, err := a.verifyAuthToken(msg.AuthTokenRequest.Token)
			fmt.Println("???", err)
			if err != nil {
				conn.WriteJSON(JSON{
					"type":    "auth",
					"success": false,
				})
				return
			}
			conn.WriteJSON(JSON{
				"type":    "auth",
				"success": true,
			})
		}
	}
}
