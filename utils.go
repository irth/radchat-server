package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/irth/radchat-server/models"
)

var ErrEmptyBody = errors.New("Empty request body")

func errorResponse(w http.ResponseWriter, err string, status int) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(JSON{
		"error": err,
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, i interface{}) error {
	if r.Body == nil {
		log.Print("Request body was empty")
		errorResponse(w, "Request body empty", http.StatusBadRequest)
		return ErrEmptyBody
	}

	err := json.NewDecoder(r.Body).Decode(i)
	if err != nil {
		log.Print("Failed to decode the JSON request body:", err)
		errorResponse(w, "Failed to decode JSON", http.StatusBadRequest)
		return err
	}

	return nil
}

type AuthTokenRequest struct {
	Token string `json:"auth_token"`
}

func (a *App) requireUser(w http.ResponseWriter, r *http.Request) (*models.User, error) {
	token := r.URL.Query().Get("auth_token")

	if len(token) == 0 {
		var authRequest AuthTokenRequest
		if err := decodeJSON(w, r, &authRequest); err != nil {
			return nil, err
		}
		token = authRequest.Token
	}

	u, err := a.verifyAuthToken(token)
	if err != nil {
		errorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return nil, err
	}

	return u, nil
}

func (h *Hub) SendToFriends(id int, f func(u *models.User) interface{}) {
	u, err := models.FindUserG(id)
	if err != nil {
		return
	}
	friendships, err := u.FriendshipsG().All()
	if err == nil {
		for _, friendship := range friendships {
			h.SendToUser(friendship.FriendID.Int, f(u))
		}
	}
}
