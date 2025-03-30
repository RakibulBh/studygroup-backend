package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Errors
var (
	ErrNotFound = errors.New("not found")
	ErrNoRows   = errors.New("user not found")
	ErrConflict = errors.New("conflict")
	ErrInternal = errors.New("internal server error")
	ErrInvalid  = errors.New("invalid input")
)

type Storage struct {
	Auth interface {
		HashPassword(password string) (string, error)
		Register(ctx context.Context, request RegisterRequest) error
		VerifyPassword(password string, hash string) (bool, error)
		GenerateJWT(userID int, expiresAt time.Time, secret string) (string, error)
		VerifyToken(tokenString string, secret string) (*jwt.Token, error)
		StoreRefreshToken(ctx context.Context, userID int, token string, expiresAt time.Time) error
		RefreshToken(ctx context.Context, userID int, tokenString string, secret string, refreshExp time.Duration, accessExp time.Duration) (string, string, error)
	}
	User interface {
		GetUserByID(ctx context.Context, id int) (User, error)
		GetUserByEmail(ctx context.Context, email string) (UserData, error)
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Auth: &AuthStore{db: db},
		User: &UserStore{db: db},
	}
}
