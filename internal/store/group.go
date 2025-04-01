package store

import (
	"context"
	"database/sql"
)

type GroupStore struct {
	db *sql.DB
}

type Group struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	HasMemberLimit bool   `json:"has_member_limit"`
	MemberLimit    int    `json:"member_limit"`
	Subject        string `json:"subject"`
	Visibility     string `json:"visibility"`
	Location       string `json:"location"`
}

func (s *GroupStore) CreateGroup(ctx context.Context, group *Group) (string, error) {
	var query string
	var id string

	if group.MemberLimit == 0 {
		query = `
			INSERT INTO groups (name, description, has_member_limit, subject, location, visibility)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`
		err := s.db.QueryRowContext(ctx, query, group.Name, group.Description, group.HasMemberLimit, group.Subject, group.Location, group.Visibility).Scan(&id)
		if err != nil {
			return "", err
		}
	} else {
		query = `
			INSERT INTO groups (name, description, has_member_limit, member_limit, subject, location, visibility)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id
		`
		err := s.db.QueryRowContext(ctx, query, group.Name, group.Description, group.HasMemberLimit, group.MemberLimit, group.Subject, group.Location, group.Visibility).Scan(&id)
		if err != nil {
			return "", err
		}
	}

	return id, nil
}

func (s *GroupStore) GetGroupByID(ctx context.Context, id string) (Group, error) {
	query := `
		SELECT id, name, description, has_member_limit, member_limit, subject, location
		FROM groups
		WHERE id = $1
	`

	row := s.db.QueryRowContext(ctx, query, id)

	var group Group
	err := row.Scan(&group.ID, &group.Name, &group.Description, &group.HasMemberLimit, &group.MemberLimit, &group.Subject, &group.Location)
	if err != nil {
		return Group{}, err
	}

	return group, nil
}

func (s *GroupStore) MakeAdmin(ctx context.Context, groupID string, userID int) error {
	query := `
		INSERT INTO membership (user_id, group_id, status, role)
		VALUES ($1, $2, 'member', 'admin')
	`

	_, err := s.db.ExecContext(ctx, query, userID, groupID)
	if err != nil {
		return err
	}

	return nil
}

func (s *GroupStore) GetUserGroups(ctx context.Context, userID int) ([]Group, error) {
	query := `
		SELECT g.id, g.name, g.description, g.has_member_limit, g.member_limit, g.subject, g.location
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
		err := rows.Scan(&group.ID, &group.Name, &group.Description, &group.HasMemberLimit, &memberLimit, &group.Subject, &group.Location)
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

func (s *GroupStore) SearchGroup(ctx context.Context, searchQuery string, userID int) ([]Group, error) {
	query := `
        SELECT 
            groups.id, 
            groups.name, 
            groups.description, 
            groups.has_member_limit, 
            groups.member_limit, 
            groups.subject, 
            groups.location,
            groups.visibility
        FROM groups
		JOIN membership m ON groups.id = m.group_id
        WHERE groups.name ILIKE '%' || $1 || '%' 
          AND visibility = 'public'
		  AND NOT EXISTS (
			SELECT 1 FROM membership 
			WHERE group_id = groups.id AND user_id = $2
		  )
        ORDER BY 
            name ASC
    `

	rows, err := s.db.QueryContext(ctx, query, searchQuery, userID)
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

func (s *GroupStore) JoinGroup(ctx context.Context, groupID string, userID int) error {
	query := `
		INSERT INTO membership (user_id, group_id, status, role)
		VALUES ($1, $2, 'member', 'member')
	`

	_, err := s.db.ExecContext(ctx, query, userID, groupID)
	if err != nil {
		return err
	}

	return nil
}

func (s *GroupStore) LeaveGroup(ctx context.Context, groupID string, userID int) error {
	query := `
		DELETE FROM membership WHERE user_id = $1 AND group_id = $2
	`

	_, err := s.db.ExecContext(ctx, query, userID, groupID)
	if err != nil {
		return err
	}

	return nil
}

func (s *GroupStore) GetJoinedGroups(ctx context.Context, userID int) ([]Group, error) {
	query := `
		SELECT g.id, g.name, g.description, g.has_member_limit, g.member_limit, g.subject, g.location
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
		err := rows.Scan(&group.ID, &group.Name, &group.Description, &group.HasMemberLimit, &memberLimit, &group.Subject, &group.Location)
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
