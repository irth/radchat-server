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

func (a *App) registerAuthHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/auth/google", a.handleGoogleAuth)
}

type AuthTokenResponse struct {
	Token     string       `json:"auth_token"`
	User      *models.User `json:"user"`
	FirstTime bool         `json:"first_time"`
}

func (a App) handleGoogleAuth(w http.ResponseWriter, r *http.Request) {
	log.Print("Google Sign-in authentication request")

	var authRequest AuthTokenRequest

	if err := decodeJSON(w, r, &authRequest); err != nil {
		return
	}

	googleAuthToken, err := a.Verifier.Verify(authRequest.Token)

	if googleAuthToken == nil || err != nil {
		// token is invalid
		errorResponse(w, fmt.Sprint("Invalid token: ", err.Error()), http.StatusUnauthorized)
		return
	}

	ru, err := models.FindRemoteUser(a.DB, fmt.Sprintf("google:%s", googleAuthToken.Subject))

	var u *models.User
	if err == sql.ErrNoRows {
		u = &models.User{
			DisplayName: googleAuthToken.Name,
			Status:      models.StatusUnavailable,
		}

		ru = &models.RemoteUser{RemoteID: fmt.Sprintf("google:%s", googleAuthToken.Subject)}

		err := ru.SetUser(a.DB, true, u)
		if err != nil {
			errorResponse(w, "Database error", http.StatusInternalServerError)
			return
		}

		err = ru.Insert(a.DB)
		if err != nil {
			errorResponse(w, "Database error", http.StatusInternalServerError)
			return
		}
	} else {
		if err != nil {
			errorResponse(w, "Database error", http.StatusInternalServerError)
			return
		}

		u, err = ru.User(a.DB).One()
		if err != nil {
			errorResponse(w, "Couldn't find the user", http.StatusUnauthorized)
			return
		}
	}

	token := uuid.NewV4().String()
	err = u.AddAuthTokens(a.DB, true, &models.AuthToken{Token: token})
	if err != nil {
		errorResponse(w, "Couldn't create a token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(AuthTokenResponse{token, u, u.Username.IsZero()})
}

func (a *App) verifyAuthToken(token string) (*models.User, error) {
	t, err := models.FindAuthToken(a.DB, token)
	if err != nil {
		return nil, err
	}

	u, err := t.User(a.DB).One()
	if err != nil {
		return nil, err
	}

	return u, nil
}
