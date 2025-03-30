package main

import (
	"encoding/json"
	"net/http"
)

type jsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (app *application) writeJSON(w http.ResponseWriter, status int, message string, data any) error {
	response := jsonResponse{
		Error:   false,
		Message: message,
		Data:    data,
	}

	js, err := json.Marshal(response)
	if err != nil {
		return err
	}

	w.WriteHeader(status)

	_, err = w.Write(js)
	if err != nil {
		return err
	}

	return nil
}

func (app *application) readJSON(r *http.Request, data any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(data)
	if err != nil {
		return err
	}

	return nil
}

func (app *application) errorJSON(w http.ResponseWriter, err error, statusCode int) error {
	response := jsonResponse{
		Error:   true,
		Message: err.Error(),
	}

	js, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.WriteHeader(statusCode)

	_, err = w.Write(js)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}
