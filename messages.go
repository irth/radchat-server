package main

import (
	"fmt"
	"net/http"

	"github.com/vattle/sqlboiler/queries/qm"
)

func (a *App) registerMessageHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/send", a.handleSend)
}

type ReqMessageSend struct {
	AuthTokenRequest
	Target  int    `json:"id"`
	Message string `json:"message"`
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
	if err != nil {
		errorResponse(w, "Unauthorized (you're not friends)", http.StatusUnauthorized)
		return
	}

	if isFriend {
		a.Hub.SendToUser(req.Target, JSON{"type": "message", "from": u.ID, "message": req.Message})
	}
}
