package main

import (
	"errors"
	"net/http"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrDuplicateEmail     = errors.New("duplicate email")
	ErrDuplicateUsername  = errors.New("duplicate username")
)

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorJSON(w, err, http.StatusBadRequest)
}

func (app *application) internalServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorJSON(w, err, http.StatusInternalServerError)
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorJSON(w, err, http.StatusNotFound)
}

func (app *application) unauthorizedResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorJSON(w, err, http.StatusUnauthorized)
}
