package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type GroupInvitation struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	GroupID   int       `json:"group_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type GroupInvitationsStore struct {
	db *sql.DB
}

func (s *GroupInvitationsStore) InviteUserToGroup(ctx context.Context, groupID int, userID int) error {
	query := `
		DELETE FROM group_invitations WHERE user_id = $1 AND group_id = $2
	`
	_, err := s.db.ExecContext(ctx, query, userID, groupID)
	if err != nil {
		return err
	}

	expiry := time.Now().Add(time.Hour * 23)

	query = `
		INSERT INTO group_invitations (user_id, group_id, expires_at)
		VALUES ($1, $2, $3)
	`

	_, err = s.db.ExecContext(ctx, query, userID, groupID, expiry)
	if err != nil {
		return err
	}

	return nil
}

func (s *GroupInvitationsStore) AcceptInvitation(ctx context.Context, userID int, groupID int) error {
	// Check if invitation is expired
	query := `
		SELECT EXISTS(SELECT 1 FROM group_invitations WHERE user_id = $1 AND group_id = $2 AND expiry > NOW())
	`

	row := s.db.QueryRowContext(ctx, query, userID, groupID)

	var exists bool
	err := row.Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("invitation expired")
	}

	// Delete invitation
	query = `
		DELETE FROM group_invitations WHERE user_id = $1 AND group_id = $2
	`
	_, err = s.db.ExecContext(ctx, query, userID, groupID)
	if err != nil {
		return err
	}

	// Insert user
	query = `
		INSERT INTO membership (user_id, group_id, role)
		VALUES ($1, $2, 'member')
	`

	_, err = s.db.ExecContext(ctx, query, userID, groupID)
	if err != nil {
		return err
	}

	return nil
}

func (s *GroupInvitationsStore) GetInvitations(ctx context.Context, userID int) ([]GroupInvitation, error) {
	query := `
		SELECT * FROM group_invitations WHERE user_id = $1
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	var invitations []GroupInvitation
	for rows.Next() {
		var invitation GroupInvitation
		err := rows.Scan(&invitation.ID, &invitation.UserID, &invitation.GroupID, &invitation.ExpiresAt)
		if err != nil {
			return nil, err
		}

		invitations = append(invitations, invitation)
	}

	return invitations, nil
}
