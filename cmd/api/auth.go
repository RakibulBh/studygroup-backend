package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/RakibulBh/studygroup-backend/internal/store"
	"github.com/golang-jwt/jwt/v5"
)

type RegisterRequest struct {
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Email           string `json:"email"`
	University      string `json:"university"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
}

func (app *application) Register(w http.ResponseWriter, r *http.Request) {

	// Parse the request
	var payload RegisterRequest
	err := app.readJSON(r, &payload)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	// Check if email exists already
	_, err = app.store.User.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNoRows):
			// do nothing
		default:
			app.internalServerErrorResponse(w, r, err)
			return
		}
	}

	// Validate the length of each field
	if len(payload.FirstName) < 2 || len(payload.LastName) < 2 || len(payload.Email) < 5 || len(payload.Password) < 8 || payload.Password != payload.PasswordConfirm {
		app.badRequestResponse(w, r, errors.New("invalid request payload"))
		return
	}

	// validate password matches
	if payload.Password != payload.PasswordConfirm {
		app.badRequestResponse(w, r, errors.New("password does not match"))
		return
	}

	// Hash the passowrd
	hash, err := app.store.Auth.HashPassword(payload.Password)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	// store the user in the database
	err = app.store.Auth.Register(ctx, store.RegisterRequest{FirstName: payload.FirstName, LastName: payload.LastName, Email: payload.Email, University: payload.University, PasswordHash: hash})
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusCreated, "registered successfully", nil)
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *application) Login(w http.ResponseWriter, r *http.Request) {

	// Parse the request
	var payload LoginRequest
	err := app.readJSON(r, &payload)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	// Fetch user from the database
	user, err := app.store.User.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNoRows):
			app.badRequestResponse(w, r, errors.New("user does not exist"))
			return
		default:
			app.internalServerErrorResponse(w, r, err)
			return
		}
	}

	// Verify password matches with the database hash
	passwordMatches, err := app.store.Auth.VerifyPassword(payload.Password, user.PasswordHash)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	if !passwordMatches {
		app.badRequestResponse(w, r, errors.New("invalid credentials"))
		return
	}

	// Generate a JWT token
	accessToken, err := app.store.Auth.GenerateJWT(user.ID, time.Now().Add(app.config.auth.exp), app.config.auth.jwtSecret)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	// Generate a Refesh JWT token
	refreshToken, err := app.store.Auth.GenerateJWT(user.ID, time.Now().Add(app.config.auth.refreshExp), app.config.auth.jwtSecret)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	// Store the refresh token in the database
	err = app.store.Auth.StoreRefreshToken(ctx, user.ID, refreshToken, time.Now().Add(app.config.auth.refreshExp))
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusAccepted, "logged in successfully", map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (app *application) Refresh(w http.ResponseWriter, r *http.Request) {

	refreshToken := r.Header.Get("Authorization")
	if refreshToken == "" {
		app.badRequestResponse(w, r, errors.New("refresh token is required"))
		return
	}

	parts := strings.Split(refreshToken, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		app.badRequestResponse(w, r, errors.New("invalid refresh token"))
		return
	}

	refreshToken = parts[1]

	// Validate if the token is valid
	jwtToken, err := app.store.Auth.VerifyToken(refreshToken, app.config.auth.jwtSecret)
	if err != nil {
		app.unauthorizedResponse(w, r, errors.New("invalid token"))
		return
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		app.unauthorizedResponse(w, r, errors.New("invalid token"))
		return
	}

	userID, err := strconv.ParseInt(fmt.Sprintf("%v", claims["user_id"]), 10, 64)
	if err != nil {
		app.unauthorizedResponse(w, r, errors.New("invalid token"))
		return
	}

	ctx := r.Context()

	// Verify the refresh token and regenerate
	accessToken, refreshToken, err := app.store.Auth.RefreshToken(ctx, int(userID), refreshToken, app.config.auth.jwtSecret, app.config.auth.refreshExp, app.config.auth.exp)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusAccepted, "refreshed successfully", map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
	}
}
