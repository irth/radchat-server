package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/irth/radchat-server/models"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/vattle/sqlboiler/queries/qm"

	null "gopkg.in/nullbio/null.v6"
)

func (a *App) registerProfileHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/profile", a.handleOwnProfile)
	mux.HandleFunc("/friends", a.handleFriends)
}

type UserUpdateRequest struct {
	AuthTokenRequest
	DisplayName null.String `json:"display_name"`
	Username    null.String `json:"username"`
	Status      null.String `json:"status"`
}

func (a *App) handleFriends(w http.ResponseWriter, r *http.Request) {
	u, err := a.requireUser(w, r)
	if err != nil {
		return
	}

	friendsList := []*models.User{}
	friendships, err := u.Friendships(a.DB, qm.Load("Friend")).All()
	if err != nil {
		errorResponse(w, err.Error(), http.StatusInternalServerError)
	}

	for _, friendship := range friendships {
		f := friendship.R.Friend
		friendsList = append(friendsList, f)
	}

	json.NewEncoder(w).Encode(JSON{
		"friends": friendsList,
	})
}

func (a *App) handleOwnProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		u, err := a.requireUser(w, r)
		if err != nil {
			return
		}

		json.NewEncoder(w).Encode(u)
	}

	if r.Method == "PATCH" {
		var data UserUpdateRequest
		err := decodeJSON(w, r, &data)
		if err != nil {
			return
		}

		u, err := a.verifyAuthToken(data.Token)
		if err != nil {
			errorResponse(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if data.DisplayName.Valid {
			u.DisplayName = data.DisplayName.String
		}

		if data.Username.Valid {
			u.Username = data.Username
		}

		if data.Status.Valid {
			u.Status = data.Status.String
		}

		err = u.Update(a.DB)
		if err != nil {
			status := http.StatusInternalServerError
			switch err := errors.Cause(err).(type) {
			case *pq.Error:
				errClass := err.Code.Class().Name()
				if errClass == "integrity_constraint_violation" || errClass == "data_exception" {
					status = http.StatusUnprocessableEntity
					// TODO: figure out what caused the error and report it
				}
				fmt.Println(errClass)
			}
			errorResponse(w, err.Error(), status)
			return
		}

		if data.Status.Valid {
			a.Hub.SendToUser(u.ID, JSON{"type": "statusUpdate", "id": u.ID, "status": u.Status}) // notify other connected clients

			// notify user's friends
			friendships, err := u.Friendships(a.DB).All()
			if err == nil {
				for _, friendship := range friendships {
					a.Hub.SendToUser(friendship.FriendID.Int, JSON{"type": "statusUpdate", "id": u.ID, "status": u.Status})
				}
			}
		}
		json.NewEncoder(w).Encode(u)
	}

}
