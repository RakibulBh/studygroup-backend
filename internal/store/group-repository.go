package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type GroupRepository struct {
	db *sql.DB
}

type Group struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	HasMemberLimit bool      `json:"has_member_limit"`
	MemberLimit    int       `json:"member_limit"`
	Subject        string    `json:"subject"`
	Description    string    `json:"description"`
	Location       string    `json:"location"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Visibility     string    `json:"visibility"`
}

func (s *GroupRepository) CreateGroup(ctx context.Context, group *Group) (int, error) {
	var query string
	var id int

	if group.MemberLimit == 0 {
		query = `
			INSERT INTO groups (name, description, has_member_limit, subject, location, visibility)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`
		err := s.db.QueryRowContext(ctx, query, group.Name, group.Description, group.HasMemberLimit, group.Subject, group.Location, group.Visibility).Scan(&id)
		if err != nil {
			return 0, err
		}
	} else {
		query = `
			INSERT INTO groups (name, description, has_member_limit, member_limit, subject, location, visibility)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id
		`
		err := s.db.QueryRowContext(ctx, query, group.Name, group.Description, group.HasMemberLimit, group.MemberLimit, group.Subject, group.Location, group.Visibility).Scan(&id)
		if err != nil {
			return 0, err
		}
	}

	return id, nil
}

func (s *GroupRepository) GetGroupByID(ctx context.Context, id int) (Group, error) {
	query := `
		SELECT id, name, description, has_member_limit, member_limit, subject, location, visibility
		FROM groups
		WHERE id = $1
	`

	row := s.db.QueryRowContext(ctx, query, id)

	var group Group
	var memberLimit sql.NullInt64
	err := row.Scan(&group.ID, &group.Name, &group.Description, &group.HasMemberLimit, &memberLimit, &group.Subject, &group.Location, &group.Visibility)
	if err != nil {
		return Group{}, err
	}

	if memberLimit.Valid {
		group.MemberLimit = int(memberLimit.Int64)
	} else {
		group.MemberLimit = 0
	}

	return group, nil
}

func (s *GroupRepository) GetUserGroups(ctx context.Context, userID int) ([]Group, error) {
	query := `
		SELECT g.id, g.name, g.description, g.has_member_limit, g.member_limit, g.subject, g.location, g.visibility, g.created_at
		FROM groups g
		JOIN membership m ON g.id = m.group_id
		WHERE m.user_id = $1 AND m.role = 'admin'
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	var groups []Group
	for rows.Next() {
		var group Group
		var memberLimit sql.NullInt64
		err := rows.Scan(&group.ID, &group.Name, &group.Description, &group.HasMemberLimit, &memberLimit, &group.Subject, &group.Location, &group.Visibility, &group.CreatedAt)
		if err != nil {
			return nil, err
		}

		if memberLimit.Valid {
			group.MemberLimit = int(memberLimit.Int64)
		} else {
			group.MemberLimit = 0
		}

		groups = append(groups, group)
	}

	return groups, nil
}

func (s *GroupRepository) SearchGroup(ctx context.Context, searchQuery string) ([]Group, error) {
	query := `
       SELECT * FROM groups WHERE name ILIKE $1 || '%' AND visibility = 'public' ORDER BY name ASC`

	rows, err := s.db.QueryContext(ctx, query, searchQuery)
	if err != nil {
		return nil, err
	}

	var groups []Group
	for rows.Next() {
		var group Group
		var memberLimit sql.NullInt64
		err := rows.Scan(&group.ID, &group.Name, &group.HasMemberLimit, &memberLimit, &group.Description, &group.Subject, &group.Location, &group.CreatedAt, &group.UpdatedAt, &group.Visibility)
		if err != nil {
			return nil, err
		}

		if memberLimit.Valid {
			group.MemberLimit = int(memberLimit.Int64)
		} else {
			group.MemberLimit = 0
		}

		groups = append(groups, group)
	}

	return groups, nil
}

func (s *GroupRepository) GetJoinedGroups(ctx context.Context, userID int) ([]Group, error) {
	query := `
		SELECT g.id, g.name, g.description, g.has_member_limit, g.member_limit, g.subject, g.location, g.visibility
		FROM groups g
		JOIN membership m ON g.id = m.group_id
		WHERE m.user_id = $1 AND m.role != 'admin'
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	var groups []Group
	for rows.Next() {
		var group Group
		var memberLimit sql.NullInt64
		err := rows.Scan(&group.ID, &group.Name, &group.Description, &group.HasMemberLimit, &memberLimit, &group.Subject, &group.Location, &group.Visibility)
		if err != nil {
			return nil, err
		}

		if memberLimit.Valid {
			group.MemberLimit = int(memberLimit.Int64)
		} else {
			group.MemberLimit = 0
		}

		groups = append(groups, group)
	}

	return groups, nil
}

func (s *GroupRepository) GetAllGroups(ctx context.Context) ([]Group, error) {
	query := `
		SELECT g.id, g.name, g.description, g.has_member_limit, g.member_limit, g.subject, g.location, g.visibility
		FROM groups g
		WHERE g.visibility = 'public'
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var groups []Group
	for rows.Next() {
		var group Group
		var memberLimit sql.NullInt64
		err := rows.Scan(&group.ID, &group.Name, &group.Description, &group.HasMemberLimit, &memberLimit, &group.Subject, &group.Location, &group.Visibility)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		if memberLimit.Valid {
			group.MemberLimit = int(memberLimit.Int64)
		} else {
			group.MemberLimit = 0
		}

		groups = append(groups, group)
	}

	return groups, nil
}

func (s *GroupRepository) DeleteGroup(ctx context.Context, groupID int) error {
	query := `
		DELETE FROM groups WHERE id = $1
	`
	_, err := s.db.ExecContext(ctx, query, groupID)
	return err
}
