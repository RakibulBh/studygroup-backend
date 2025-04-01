package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/RakibulBh/studygroup-backend/internal/store"
	"github.com/go-chi/chi/v5"
)

func (app *application) GetGroups(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusOK, "Groups fetched successfully", nil)
}

func (app *application) GetGroup(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusOK, "Group fetched successfully", nil)
}

type CreateGroupRequest struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	HasMemberLimit bool   `json:"has_member_limit"`
	MemberLimit    int    `json:"member_limit"`
	Subject        string `json:"subject"`
	Location       string `json:"location"`
	Visibility     string `json:"visibility"`
}

func (app *application) CreateGroup(w http.ResponseWriter, r *http.Request) {

	// Get user id from context
	user := r.Context().Value(userCtx).(store.User)
	fmt.Println("User: ", user)

	// Parse the request
	var payload CreateGroupRequest
	err := app.readJSON(r, &payload)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Verify empty data
	if payload.Name == "" || payload.Description == "" || payload.Subject == "" || payload.Location == "" {
		app.badRequestResponse(w, r, errors.New("invalid request"))
		return
	}

	// Verify values
	if payload.Visibility != "public" && payload.Visibility != "private" {
		app.badRequestResponse(w, r, errors.New("invalid visibility value"))
		return
	}

	// Verify lengths
	if len(payload.Name) > 100 || len(payload.Description) > 500 || len(payload.Subject) > 100 || len(payload.Location) > 100 {
		app.badRequestResponse(w, r, errors.New("invalid character limit"))
		return
	}

	// Check if member limit being true or false
	var memberLimit int
	if payload.HasMemberLimit {
		memberLimit = payload.MemberLimit
	} else {
		memberLimit = 0
	}

	ctx := r.Context()

	group := &store.Group{
		Name:           payload.Name,
		Description:    payload.Description,
		HasMemberLimit: payload.HasMemberLimit,
		MemberLimit:    memberLimit,
		Subject:        payload.Subject,
		Location:       payload.Location,
	}

	// Create group
	id, err := app.store.Group.CreateGroup(ctx, group)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	group.ID = id

	// Make admin
	err = app.store.Group.MakeAdmin(ctx, group.ID, user.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, "Group created successfully", nil)
}

func (app *application) GetUserGroups(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userCtx).(store.User)
	groups, err := app.store.Group.GetUserGroups(r.Context(), user.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	fmt.Println("Groups: ", groups)

	app.writeJSON(w, http.StatusOK, "Groups fetched successfully", groups)
}

func (app *application) SearchGroup(w http.ResponseWriter, r *http.Request) {

	// Get search query from path
	searchQuery := chi.URLParam(r, "search_query")

	ctx := r.Context()

	// Get user id from context
	user := r.Context().Value(userCtx).(store.User)

	// Search for group
	groups, err := app.store.Group.SearchGroup(ctx, searchQuery, user.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, "Group searched successfully", groups)
}

func (app *application) JoinGroup(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusOK, "Group joined successfully", nil)
}

func (app *application) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusOK, "Group left successfully", nil)
}

// Get the user's joined groups
func (app *application) GetJoinedGroups(w http.ResponseWriter, r *http.Request) {

	// Get user id from context
	user := r.Context().Value(userCtx).(store.User)

	// Get group id from path
	ctx := r.Context()

	// Get joined groups
	groups, err := app.store.Group.GetJoinedGroups(ctx, user.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, "Joined groups fetched successfully", groups)
}

func (app *application) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusOK, "Group updated successfully", nil)
}

func (app *application) GetAllGroups(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userCtx).(store.User)

	ctx := r.Context()

	groups, err := app.store.Group.GetAllGroups(ctx, user.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, "All groups fetched successfully", groups)
}
