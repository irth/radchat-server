package main

import (
	"encoding/json"
	"net/http"
)

func (a *App) registerProfileHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/profile/me", a.handleOwnProfile)
}

func (a *App) handleOwnProfile(w http.ResponseWriter, r *http.Request) {
	u, err := a.requireUser(w, r)
	if err != nil {
		return
	}

	json.NewEncoder(w).Encode(u)
}
