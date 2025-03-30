package main

import (
	"net/http"
)

func (app *application) GetGroups(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusOK, "Groups fetched successfully", nil)
}

func (app *application) GetGroup(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusOK, "Group fetched successfully", nil)
}
