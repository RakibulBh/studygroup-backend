package store

import (
	"context"
	"database/sql"
)

type GroupMembershipStore struct {
	db *sql.DB
}

func (s *GroupMembershipStore) GetGroupMembers(ctx context.Context, groupID int) ([]User, error) {
	query := `
		SELECT u.id, u.first_name, u.last_name, u.email
		FROM users u
		JOIN membership m ON u.id = m.user_id
		WHERE m.group_id = $1 AND (m.role = 'member' OR m.role = 'admin')
	`

	rows, err := s.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, err
	}

	var members []User
	for rows.Next() {
		var member User
		err := rows.Scan(&member.ID, &member.FirstName, &member.LastName, &member.Email)
		if err != nil {
			return nil, err
		}

		members = append(members, member)
	}

	return members, nil
}

func (s *GroupMembershipStore) GetMemberCount(ctx context.Context, groupID int) (int, error) {
	query := `
		SELECT COUNT(*) FROM membership WHERE group_id = $1
		`

	row := s.db.QueryRowContext(ctx, query, groupID)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *GroupMembershipStore) IsAdmin(ctx context.Context, groupID int, userID int) (bool, error) {
	query := `
		SELECT EXISTS(SELECT 1 FROM membership WHERE group_id = $1 AND user_id = $2 AND role = 'admin')
	`

	row := s.db.QueryRowContext(ctx, query, groupID, userID)

	var exists bool
	err := row.Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (s *GroupMembershipStore) IsMember(ctx context.Context, groupID int, userID int) (bool, error) {
	query := `
		SELECT EXISTS(SELECT 1 FROM membership WHERE group_id = $1 AND user_id = $2)
	`

	row := s.db.QueryRowContext(ctx, query, groupID, userID)

	var exists bool
	err := row.Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
