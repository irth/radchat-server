package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/irth/radchat-server/models"
	"github.com/vattle/sqlboiler/queries/qm"
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
	*MsgBufferChange
}

func readPump(conn *websocket.Conn, ch chan Message) {
	for {
		var msg Message
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
	defer a.Hub.UnregisterClient(u.ID)
	readChannel := make(chan Message)
	go readPump(conn, readChannel)

	for {
		select {
		case msg, more := <-readChannel:
			if !more {
				return
			}
			switch msg.Type {
			case "getFriends":
				friendsList := []*models.User{}
				friendships, err := u.Friendships(a.DB, qm.Load("Friend")).All()
				if err != nil {
					continue
				}
				for _, friendship := range friendships {
					f := friendship.R.Friend
					friendsList = append(friendsList, f)
				}
				conn.WriteJSON(JSON{
					"type":    "friendsList",
					"friends": friendsList,
				})
			case "inputBufferUpdate":
				msg.Sender = u.ID
				client.Output <- *msg.MsgBufferChange
			}
		case msg := <-client.Input:
			conn.WriteJSON(msg)
		}
	}
}
