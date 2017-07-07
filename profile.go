package main

import (
	"encoding/json"
	"fmt"
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
	Status      null.String `json:"status"`
}

func (a *App) handleOwnProfile(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Method)
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
		json.NewEncoder(w).Encode(u)
	}

}
