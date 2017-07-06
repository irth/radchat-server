package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/irth/radchat-server/models"
)

var ErrEmptyBody = errors.New("Empty request body")

func decodeJSON(w http.ResponseWriter, r *http.Request, i interface{}) error {
	if r.Body == nil {
		log.Print("Request body was empty")
		http.Error(w, "Request body empty", http.StatusBadRequest)
		return ErrEmptyBody
	}

	err := json.NewDecoder(r.Body).Decode(i)
	if err != nil {
		log.Print("Failed to decode the JSON request body:", err)
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return err
	}

	return nil
}

type AuthTokenRequest struct {
	Token string `json:"auth_token"`
}

func (a *App) requireUser(w http.ResponseWriter, r *http.Request) (*models.User, error) {
	var authRequest AuthTokenRequest

	if err := decodeJSON(w, r, &authRequest); err != nil {
		return nil, err
	}

	u, err := a.verifyAuthToken(authRequest.Token)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil, err
	}

	return u, nil
}
