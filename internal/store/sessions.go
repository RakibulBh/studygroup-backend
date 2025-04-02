package store

import (
	"context"
	"database/sql"
	"time"
)

type SessionStore struct {
	db *sql.DB
}

type StudySession struct {
	ID          int       `json:"id"`
	GroupID     int       `json:"group_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	CreatedAt   time.Time `json:"created_at"`
}

func (s *SessionStore) CreateStudySession(ctx context.Context, session *StudySession) (int, error) {
	query := `
		INSERT INTO study_sessions (group_id, title, description, location, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := s.db.QueryRowContext(ctx, query, session.GroupID, session.Title, session.Description, session.Location, session.StartTime, session.EndTime).Scan(&session.ID)
	if err != nil {
		return 0, err
	}

	return session.ID, nil
}

func (s *SessionStore) GetUserStudySessions(ctx context.Context, userID int) ([]StudySession, error) {
	query := `
		SELECT s.id, s.group_id, s.title, s.description, s.location, s.start_time, s.end_time, s.created_at
		FROM study_sessions s
		INNER JOIN groups g ON s.group_id = g.id
		INNER JOIN membership m ON s.group_id = m.group_id
		WHERE m.user_id = $1 AND m.role IN ('member', 'admin') AND s.end_time > NOW()
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var sessions []StudySession

	for rows.Next() {
		var session StudySession
		err := rows.Scan(&session.ID, &session.GroupID, &session.Title, &session.Description, &session.Location, &session.StartTime, &session.EndTime, &session.CreatedAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (s *SessionStore) GetGroupStudySessions(ctx context.Context, groupID int) ([]StudySession, error) {
	query := `
		SELECT id, group_id, title, description, location, start_time, end_time, created_at
		FROM study_sessions
		WHERE group_id = $1
	`

	rows, err := s.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var sessions []StudySession

	for rows.Next() {
		var session StudySession
		err := rows.Scan(&session.ID, &session.GroupID, &session.Title, &session.Description, &session.Location, &session.StartTime, &session.EndTime, &session.CreatedAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}
