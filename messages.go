package main

import (
	"fmt"
	"net/http"

	"encoding/json"

	"github.com/irth/radchat-server/models"
	"github.com/vattle/sqlboiler/queries/qm"
)

func (a *App) registerMessageHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/send", a.handleSend)
}

type ReqMessageSend struct {
	AuthTokenRequest
	Target  int    `json:"target"`
	Message string `json:"message"`
}

type ResMessageSend struct {
	Success bool   `json:"success"`
	ID      int    `json:"id,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (a *App) handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		errorResponse(w, "Not found", http.StatusNotFound)
		return
	}

	var req ReqMessageSend
	err := decodeJSON(w, r, &req)
	if err != nil {
		return
	}
	u, err := a.verifyAuthToken(req.Token)
	if err != nil {
		errorResponse(w, "Unauthorized", http.StatusUnauthorized)
	}

	isFriend, err := u.FriendshipsG(qm.Where("friend_id=?", req.Target)).Exists()

	fmt.Println(req, isFriend, err)
	if err != nil || !isFriend {
		errorResponse(w, "Unauthorized (you're not friends)", http.StatusUnauthorized)
		return
	}

	m := models.Message{
		SenderID: u.ID,
		TargetID: req.Target,
		Content:  req.Message,
	}

	err = m.InsertG()
	if err != nil {
		errorResponse(w, "Couldn't send the message", http.StatusInternalServerError)
		return
	}

	a.Hub.SendToUser(req.Target, JSON{"type": "message", "id": m.ID, "sender": u.ID, "message": req.Message})
	json.NewEncoder(w).Encode(ResMessageSend{Success: true, ID: m.ID})
}
