package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/RakibulBh/studygroup-backend/internal/store"
	"github.com/go-chi/chi/v5"
)

type StudySession struct {
	ID          int       `json:"id"`
	GroupID     int       `json:"group_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	CreatedAt   time.Time `json:"created_at"`
}

func (app *application) GetGroupStudySessions(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Getting")

	// Get group ID from URL param
	groupID := chi.URLParam(r, "groupID")

	groupIDInt, err := strconv.Atoi(groupID)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	sessions, err := app.store.Session.GetGroupStudySessions(ctx, groupIDInt)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, "Group study sessions fetched successfully", sessions)
}

type CreateStudySessionRequest struct {
	GroupID     int       `json:"group_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
}

func (app *application) CreateStudySession(w http.ResponseWriter, r *http.Request) {

	var payload CreateStudySessionRequest
	err := app.readJSON(r, &payload)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	groupID := chi.URLParam(r, "groupID")

	groupIDInt, err := strconv.Atoi(groupID)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := r.Context().Value(userCtx).(store.User)
	ctx := r.Context()

	// Check if user is admin
	isAdmin, err := app.store.GroupMembership.IsAdmin(ctx, groupIDInt, user.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
	if !isAdmin {
		app.badRequestResponse(w, r, errors.New("user is not admin"))
		return
	}

	id, err := app.store.Session.CreateStudySession(ctx, &store.StudySession{
		GroupID:     groupIDInt,
		Title:       payload.Title,
		Description: payload.Description,
		Location:    payload.Location,
		StartTime:   payload.StartTime,
		EndTime:     payload.EndTime,
	})
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, "Study session created successfully", id)
}

func (app *application) GetUserStudySessions(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userCtx).(store.User)
	ctx := r.Context()

	sessions, err := app.store.Session.GetUserStudySessions(ctx, user.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, "User study sessions fetched successfully", sessions)
}

func (app *application) DeleteStudySession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")

	sessionIDInt, err := strconv.Atoi(sessionID)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := r.Context().Value(userCtx).(store.User)
	ctx := r.Context()

	// Check if session id exists
	session, err := app.store.Session.GetStudySessionByID(ctx, sessionIDInt)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			app.notFoundResponse(w, r, errors.New("study session does not exist"))
		default:
			app.internalServerErrorResponse(w, r, err)
		}
		return
	}

	// Check if user is admin
	isAdmin, err := app.store.GroupMembership.IsAdmin(ctx, session.GroupID, user.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
	if !isAdmin {
		app.badRequestResponse(w, r, errors.New("user is not admin"))
		return
	}

	err = app.store.Session.DeleteStudySession(ctx, session.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, "Study session deleted successfully", nil)
}
