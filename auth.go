package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"database/sql"

	"github.com/irth/radchat-server/models"
	uuid "github.com/satori/go.uuid"
)

type AuthRequest struct {
	Token string `json:"authToken"`
}

func (a App) handleAuth(w http.ResponseWriter, r *http.Request) {
	log.Print("Google Sign-in authentication request")

	if r.Body == nil {
		log.Print("Request body was empty")
		http.Error(w, "Request body empty", http.StatusBadRequest)
		return
	}

	var authRequest AuthRequest

	err := json.NewDecoder(r.Body).Decode(&authRequest)
	if err != nil {
		log.Print("Failed to decode the JSON request body:", err)
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}

	googleAuthToken, err := a.Verifier.Verify(authRequest.Token)

	if googleAuthToken == nil || err != nil {
		// token is invalid
		http.Error(w, fmt.Sprint("Invalid token: ", err.Error()), http.StatusUnauthorized)
		return
	}

	ru, err := models.FindRemoteUser(a.DB, fmt.Sprintf("google:%s", googleAuthToken.Subject))

	var u *models.User
	if err == sql.ErrNoRows {
		u = &models.User{
			DisplayName: googleAuthToken.Name,
		}

		ru = &models.RemoteUser{RemoteID: fmt.Sprintf("google:%s", googleAuthToken.Subject)}

		err := ru.SetUser(a.DB, true, u)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		err = ru.Insert(a.DB)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	} else {
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		u, err = ru.User(a.DB).One()
		if err != nil {
			http.Error(w, "Couldn't find the user", http.StatusUnauthorized)
			return
		}
	}

	token := uuid.NewV4().String()
	err = u.AddAuthTokens(a.DB, true, &models.AuthToken{Token: token})
	if err != nil {
		http.Error(w, "Couldn't create a token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(AuthRequest{token})
}
