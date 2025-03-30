package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthStore struct {
	db *sql.DB
}

type RegisterRequest struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
}

func (s *AuthStore) Register(ctx context.Context, request RegisterRequest) error {

	query := `
		INSERT INTO users (first_name, last_name, email, password_hash)
		VALUES ($1, $2, $3, $4)
	`

	_, err := s.db.ExecContext(ctx, query, request.FirstName, request.LastName, request.Email, request.PasswordHash)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthStore) StoreRefreshToken(ctx context.Context, userID int, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`

	_, err := s.db.ExecContext(ctx, query, userID, token, expiresAt)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthStore) HashPassword(password string) (string, error) {

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func (s *AuthStore) VerifyPassword(password string, hash string) (bool, error) {

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		switch err {
		case bcrypt.ErrMismatchedHashAndPassword:
			return false, errors.New("invalid credentials")
		default:
			return false, err
		}
	}

	return true, nil
}

func (s *AuthStore) GenerateJWT(userID int, expiresAt time.Time, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     expiresAt.Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *AuthStore) VerifyToken(tokenString string, secret string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}

func (s *AuthStore) RefreshToken(ctx context.Context, userID int, tokenString string, secret string, refreshExp time.Duration, accessExp time.Duration) (string, string, error) {

	// Delete the old refresh token
	query := `
		DELETE FROM refresh_tokens
		WHERE token = $1 AND user_id = $2
	`
	_, err := s.db.ExecContext(ctx, query, tokenString, userID)
	if err != nil {
		return "", "", err
	}

	// Generate a new refresh token
	refreshToken, err := s.GenerateJWT(userID, time.Now().Add(refreshExp), secret)
	if err != nil {
		return "", "", err
	}

	// Store the new refresh token
	err = s.StoreRefreshToken(ctx, userID, refreshToken, time.Now().Add(refreshExp))
	if err != nil {
		return "", "", err
	}

	// Generate a new access token
	tokenString, err = s.GenerateJWT(userID, time.Now().Add(accessExp), secret)
	if err != nil {
		return "", "", err
	}

	return tokenString, refreshToken, nil
}
