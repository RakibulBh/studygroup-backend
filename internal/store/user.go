package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type UserStore struct {
	db *sql.DB
}

type UserData struct {
	ID           int
	FirstName    string
	LastName     string
	Email        string
	University   string
	PasswordHash string
	Subject      string
	ImageUrl     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type User struct {
	ID         int    `json:"id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	University string `json:"university"`
	Subject    string `json:"subject"`
	ImageUrl   string `json:"image_url"`
}

func (s *UserStore) GetUserByID(ctx context.Context, id int) (User, error) {

	query := `
	SELECT id, first_name, last_name, email, university, subject, image_url
	FROM users
	WHERE id = $1
	`

	var fetchedUser User
	err := s.db.QueryRowContext(ctx, query, id).Scan(&fetchedUser.ID, &fetchedUser.FirstName, &fetchedUser.LastName, &fetchedUser.Email, &fetchedUser.University, &fetchedUser.Subject, &fetchedUser.ImageUrl)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return User{}, ErrNoRows
		default:
			return User{}, err
		}
	}

	return fetchedUser, nil
}

func (s *UserStore) GetUserByEmail(ctx context.Context, email string) (UserData, error) {

	query := `
	SELECT id, first_name, last_name, email, university, password_hash, subject, image_url	
	FROM users
	WHERE email = $1
	`

	var fecthedUser UserData

	err := s.db.QueryRowContext(ctx, query, email).Scan(&fecthedUser.ID, &fecthedUser.FirstName, &fecthedUser.LastName, &fecthedUser.Email, &fecthedUser.University, &fecthedUser.PasswordHash, &fecthedUser.Subject, &fecthedUser.ImageUrl)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return UserData{}, ErrNoRows
		default:
			return UserData{}, err
		}
	}

	return fecthedUser, nil
}
