package main

import (
	"encoding/json"
	"net/http"

	"github.com/lib/pq"
	"github.com/pkg/errors"

	null "gopkg.in/nullbio/null.v6"
)

func (a *App) registerProfileHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/profile/me", a.handleOwnProfile)
}

type UserUpdateRequest struct {
	AuthTokenRequest
	DisplayName null.String `json:"display_name"`
	Username    null.String `json:"username"`
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
			return
		}

		if data.DisplayName.Valid {
			u.DisplayName = data.DisplayName.String
		}

		if data.Username.Valid {
			u.Username = data.Username
		}

		err = u.Update(a.DB)
		if err != nil {
			status := http.StatusInternalServerError
			switch err := errors.Cause(err).(type) {
			case *pq.Error:
				if err.Code.Class().Name() == "integrity_constraint_violation" {
					status = http.StatusUnprocessableEntity
					// TODO: figure out what caused the constraint and report it
				}
			}
			errorResponse(w, err.Error(), status)
			return
		}

		json.NewEncoder(w).Encode(u)
	}

}
