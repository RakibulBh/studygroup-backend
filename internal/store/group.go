package store

import (
	"context"
	"database/sql"
	"fmt"
)

type GroupStore struct {
	db *sql.DB
}

type Group struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	HasMemberLimit bool   `json:"has_member_limit"`
	MemberLimit    int    `json:"member_limit"`
	Subject        string `json:"subject"`
	Visibility     string `json:"visibility"`
	Location       string `json:"location"`
}

func (s *GroupStore) CreateGroup(ctx context.Context, group *Group) (int, error) {
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

func (s *GroupStore) GetGroupByID(ctx context.Context, id int) (Group, error) {
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

func (s *GroupStore) MakeAdmin(ctx context.Context, groupID int, userID int) error {
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
			g.id,
			g.name,
			g.description,
			g.has_member_limit,
			g.member_limit,
			g.subject,
			g.location
		FROM
			groups g
		LEFT JOIN membership m ON g.id = m.group_id AND m.user_id = $1
		WHERE
			g.name ILIKE '%' || $2 || '%'
			AND g.visibility = 'public'
			AND m.user_id IS NULL  -- More efficient than NOT EXISTS
		ORDER BY
			g.name ASC
    `

	rows, err := s.db.QueryContext(ctx, query, userID, searchQuery)
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

func (s *GroupStore) JoinGroup(ctx context.Context, groupID int, userID int) error {
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

func (s *GroupStore) LeaveGroup(ctx context.Context, groupID int, userID int) error {
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

func (s *GroupStore) GetAllGroups(ctx context.Context, userID int) ([]Group, error) {
	query := `
		SELECT g.id, g.name, g.description, g.has_member_limit, g.member_limit, g.subject, g.location
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
		err := rows.Scan(&group.ID, &group.Name, &group.Description, &group.HasMemberLimit, &memberLimit, &group.Subject, &group.Location)
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
