package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"encoding/json"

	"github.com/irth/radchat-server/models"
	"github.com/vattle/sqlboiler/queries/qm"
)

func (a *App) registerMessageHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/send", a.handleSend)
	mux.HandleFunc("/history", a.handleHistory)
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

func (a *App) handleHistory(w http.ResponseWriter, r *http.Request) {
	u, err := a.requireUser(w, r)
	if err != nil {
		return
	}

	friendIDStr := r.URL.Query().Get("friend")
	friendID, err := strconv.Atoi(friendIDStr)

	if err != nil {
		errorResponse(w, "Unknown or invalid friend ID.", http.StatusNotFound)
		return
	}

	count, err := strconv.Atoi(r.URL.Query().Get("count"))

	if err != nil || count > 30 || count < 0 {
		count = 10
	}

	queryMods := []qm.QueryMod{}
	queryMods = append(queryMods, qm.Where("((sender_id=? AND target_id=?) OR (sender_id=? AND target_id=?))", u.ID, friendID, friendID, u.ID))

	timestampStr := r.URL.Query().Get("before")
	if len(timestampStr) > 0 {
		timestamp, err := time.Parse("2006-01-02T15:04:05.999999999Z", timestampStr)
		if err != nil {
			errorResponse(w, "Invalid date.", http.StatusUnprocessableEntity)
			return
		}
		queryMods = append(queryMods, qm.And("created_at < ?", timestamp))
	}

	queryMods = append(queryMods, qm.OrderBy("created_at DESC"), qm.Limit(count))

	messages, err := models.MessagesG(queryMods...).All()
	if err != nil {
		errorResponse(w, "Database error", http.StatusInternalServerError)
		return
	}

	if messages == nil {
		messages = models.MessageSlice{}
	}

	// reverse the messages
	// https://github.com/golang/go/wiki/SliceTricks
	for left, right := 0, len(messages)-1; left < right; left, right = left+1, right-1 {
		messages[left], messages[right] = messages[right], messages[left]
	}

	json.NewEncoder(w).Encode(messages)
}
