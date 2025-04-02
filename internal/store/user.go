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
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type User struct {
	ID         int    `json:"id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	University string `json:"university"`
}

func (s *UserStore) GetUserByID(ctx context.Context, id int) (User, error) {

	query := `
	SELECT id, first_name, last_name, email, university
	FROM users
	WHERE id = $1
	`

	var fetchedUser User
	err := s.db.QueryRowContext(ctx, query, id).Scan(&fetchedUser.ID, &fetchedUser.FirstName, &fetchedUser.LastName, &fetchedUser.Email, &fetchedUser.University)

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
	SELECT id, first_name, last_name, email, university, password_hash
	FROM users
	WHERE email = $1
	`

	var fecthedUser UserData
	err := s.db.QueryRowContext(ctx, query, email).Scan(&fecthedUser.ID, &fecthedUser.FirstName, &fecthedUser.LastName, &fecthedUser.Email, &fecthedUser.University, &fecthedUser.PasswordHash)

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
