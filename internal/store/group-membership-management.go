package store

import (
	"context"
	"database/sql"
)

type GroupMembershipManagementStore struct {
	db *sql.DB
}

func (s *GroupMembershipManagementStore) MakeAdmin(ctx context.Context, groupID int, userID int) error {
	query := `
		INSERT INTO membership (user_id, group_id, role)
		VALUES ($1, $2, 'admin')
	`

	_, err := s.db.ExecContext(ctx, query, userID, groupID)
	if err != nil {
		return err
	}

	return nil

}

func (s *GroupMembershipManagementStore) LeaveGroup(ctx context.Context, groupID int, userID int) error {
	query := `
		DELETE FROM membership WHERE user_id = $1 AND group_id = $2
	`

	_, err := s.db.ExecContext(ctx, query, userID, groupID)
	if err != nil {
		return err
	}

	return nil
}

func (s *GroupMembershipManagementStore) KickUserFromGroup(ctx context.Context, groupID int, userID int) error {
	query := `
		DELETE FROM membership WHERE user_id = $1 AND group_id = $2
	`

	_, err := s.db.ExecContext(ctx, query, userID, groupID)
	if err != nil {
		return err
	}

	return nil
}
