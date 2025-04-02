package store

import (
	"context"
	"database/sql"
)

type GroupJoinRequestsStore struct {
	db *sql.DB
}

type GroupJoinRequest struct {
	ID      int    `json:"id"`
	UserID  int    `json:"user_id"`
	GroupID int    `json:"group_id"`
	Status  string `json:"status"`
}

func (s *GroupJoinRequestsStore) JoinRequest(ctx context.Context, groupID int, userID int) error {
	query := `
		INSERT INTO join_requests (user_id, group_id, status)
		VALUES ($1, $2, 'pending')
	`

	_, err := s.db.ExecContext(ctx, query, userID, groupID)
	if err != nil {
		return err
	}

	return nil
}

func (s *GroupJoinRequestsStore) GetJoinRequests(ctx context.Context, groupID int) ([]GroupJoinRequest, error) {
	query := `
		SELECT * FROM join_requests WHERE group_id = $1
	`

	rows, err := s.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, err
	}

	var joinRequests []GroupJoinRequest
	for rows.Next() {
		var joinRequest GroupJoinRequest
		err := rows.Scan(&joinRequest.ID, &joinRequest.UserID, &joinRequest.GroupID, &joinRequest.Status)
		if err != nil {
			return nil, err
		}

		joinRequests = append(joinRequests, joinRequest)
	}

	return joinRequests, nil
}

func (s *GroupJoinRequestsStore) ApproveJoinRequest(ctx context.Context, groupID int, userID int) error {
	query := `
		DELETE FROM join_requests WHERE group_id = $1 AND user_id = $2
	`

	_, err := s.db.ExecContext(ctx, query, groupID, userID)
	if err != nil {
		return err
	}

	query = `
		INSERT INTO membership (user_id, group_id, status, role)
		VALUES ($1, $2, 'member', 'member')
	`

	_, err = s.db.ExecContext(ctx, query, userID, groupID)
	if err != nil {
		return err
	}

	return nil
}

func (s *GroupJoinRequestsStore) RejectJoinRequest(ctx context.Context, groupID int, userID int) error {
	query := `
		DELETE FROM join_requests WHERE group_id = $1 AND user_id = $2
	`

	_, err := s.db.ExecContext(ctx, query, groupID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *GroupJoinRequestsStore) IsJoinRequested(ctx context.Context, groupID int, userID int) (bool, error) {
	query := `
		SELECT EXISTS(SELECT 1 FROM join_requests WHERE group_id = $1 AND user_id = $2)
	`

	row := s.db.QueryRowContext(ctx, query, groupID, userID)

	var exists bool
	err := row.Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
