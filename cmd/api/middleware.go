package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// UserIDKey is the key for the user ID in the request context
type contextKey string

const (
	userCtx contextKey = "user"
)

// Authentication middleware
func (app *application) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorizedResponse(w, r, errors.New("authorization header is required"))
			return
		}

		// Split the header and see if it is a bearer token
		parts := strings.Split(authHeader, " ")
		if parts[0] != "Bearer" || len(parts[1]) == 0 {
			app.unauthorizedResponse(w, r, errors.New("invalid authorization header"))
			return
		}

		token := parts[1]

		// Validate if the token is valid
		jwtToken, err := app.store.Auth.VerifyToken(token, app.config.auth.jwtSecret)
		if err != nil {
			app.unauthorizedResponse(w, r, errors.New("invalid token"))
			return
		}

		// Get the user from the token
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

		// Fetch the user
		user, err := app.store.User.GetUserByID(ctx, int(userID))
		if err != nil {
			app.unauthorizedResponse(w, r, errors.New("invalid user"))
			return
		}

		ctx = context.WithValue(ctx, userCtx, user)

		// If all checks pass the user is allowed to go through
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
