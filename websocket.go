package main

import (
	"fmt"
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
	*AuthTokenRequest
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
	conn, err := upgrader.Upgrade(w, r, nil)
	defer conn.Close()
	if err != nil {
		log.Println(err)
		return
	}

	var msg Message
	err = conn.ReadJSON(&msg)

	if err != nil {
		return
	}

	var u *models.User
	var client *Client
	switch msg.Type {
	case "auth":
		u, err = a.verifyAuthToken(msg.AuthTokenRequest.Token)
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
		client = a.Hub.RegisterClient(u.ID)
		defer a.Hub.UnregisterClient(u.ID)
	default:
		conn.WriteJSON(JSON{
			"type":  "error",
			"error": "Authorize first.",
		})
		return
	}

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
				friendsList := []JSON{}
				friendships, err := u.Friendships(a.DB, qm.Load("Friend")).All()
				if err != nil {
					continue
				}
				for _, friendship := range friendships {
					f := friendship.R.Friend
					friendsList = append(friendsList, JSON{
						"display_name": f.DisplayName,
						"id":           f.ID,
					})
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
