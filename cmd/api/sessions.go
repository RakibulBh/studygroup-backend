package main

import (
	"net/http"
)

func (app *application) GetSessions(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusOK, "Sessions fetched successfully", nil)
}

func (app *application) GetSession(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusOK, "Session fetched successfully", nil)
}
