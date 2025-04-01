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

type GroupWithMetadata struct {
	Group
	Metadata GroupMetadata `json:"metadata"`
}

type GroupMetadata struct {
	JoinRequested bool `json:"join_requested"`
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

func (s *GroupStore) GetGroupMembers(ctx context.Context, groupID int) ([]User, error) {
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

func (s *GroupStore) GetUserGroups(ctx context.Context, userID int) ([]Group, error) {
	query := `
		SELECT g.id, g.name, g.description, g.has_member_limit, g.member_limit, g.subject, g.location, g.visibility
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

func (s *GroupStore) SearchGroup(ctx context.Context, searchQuery string, userID int) ([]GroupWithMetadata, error) {
	query := `
       SELECT
			g.id,
			g.name,
			g.description,
			g.has_member_limit,
			g.member_limit,
			g.subject,
			g.location,
			g.visibility
		FROM
			groups g
		LEFT JOIN membership m ON g.id = m.group_id AND m.user_id = $1
		WHERE
			g.name ILIKE '%' || $2 || '%'
			AND g.visibility = 'public'
			AND m.user_id IS NULL
		ORDER BY
			g.name ASC
    `

	rows, err := s.db.QueryContext(ctx, query, userID, searchQuery)
	if err != nil {
		return nil, err
	}

	var groups []GroupWithMetadata
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

		// Verify if the user requested for the group already
		joinRequested, err := s.IsJoinRequested(ctx, group.ID, userID)
		if err != nil {
			return nil, err
		}

		groups = append(groups, GroupWithMetadata{
			Group: group,
			Metadata: GroupMetadata{
				JoinRequested: joinRequested,
			},
		})
	}

	return groups, nil
}

func (s *GroupStore) IsJoinRequested(ctx context.Context, groupID int, userID int) (bool, error) {
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

func (s *GroupStore) GetAllGroups(ctx context.Context, userID int) ([]Group, error) {
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

// ROLE CHECKS

func (s *GroupStore) IsAdmin(ctx context.Context, groupID int, userID int) (bool, error) {
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

func (s *GroupStore) IsMember(ctx context.Context, groupID int, userID int) (bool, error) {
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

// Join and leaving a group

type JoinRequest struct {
	ID      int    `json:"id"`
	UserID  int    `json:"user_id"`
	GroupID int    `json:"group_id"`
	Status  string `json:"status"`
}

func (s *GroupStore) JoinRequest(ctx context.Context, groupID int, userID int) error {
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

func (s *GroupStore) GetJoinRequests(ctx context.Context, groupID int) ([]JoinRequest, error) {
	query := `
		SELECT * FROM join_requests WHERE group_id = $1
	`

	rows, err := s.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, err
	}

	var joinRequests []JoinRequest
	for rows.Next() {
		var joinRequest JoinRequest
		err := rows.Scan(&joinRequest.ID, &joinRequest.UserID, &joinRequest.GroupID, &joinRequest.Status)
		if err != nil {
			return nil, err
		}

		joinRequests = append(joinRequests, joinRequest)
	}

	return joinRequests, nil
}

func (s *GroupStore) ApproveJoinRequest(ctx context.Context, groupID int, userID int) error {
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

func (s *GroupStore) RejectJoinRequest(ctx context.Context, groupID int, userID int) error {
	query := `
		DELETE FROM join_requests WHERE group_id = $1 AND user_id = $2
	`

	_, err := s.db.ExecContext(ctx, query, groupID, userID)
	if err != nil {
		return err
	}

	return nil
}
