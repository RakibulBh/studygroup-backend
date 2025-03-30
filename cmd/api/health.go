package main

import (
	"net/http"
)

type envelope map[string]any

func (app *application) Healthcheck(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusOK, "success", envelope{"status": "ok"})
}
