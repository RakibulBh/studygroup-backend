package store

import (
	"context"
	"database/sql"
	"math"
	"time"
)

type GroupRepository struct {
	db *sql.DB
}

type Group struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	HasMemberLimit bool      `json:"has_member_limit"`
	MemberLimit    *int      `json:"member_limit"`
	Description    string    `json:"description"`
	Subject        string    `json:"subject"`
	Visibility     string    `json:"visibility"`
	Location       string    `json:"location"`
	Latitude       *float64  `json:"latitude"`
	Longitude      *float64  `json:"longitude"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (s *GroupRepository) CreateGroup(ctx context.Context, group *Group) (int, error) {
	var query string
	var id int

	if group.MemberLimit == nil {
		query = `
			INSERT INTO groups (name, description, has_member_limit, subject, location, latitude, longitude, visibility)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id
		`
		err := s.db.QueryRowContext(ctx, query, group.Name, group.Description, group.HasMemberLimit, group.Subject, group.Location, group.Latitude, group.Longitude, group.Visibility).Scan(&id)
		if err != nil {
			return 0, err
		}
	} else {
		query = `
			INSERT INTO groups (name, description, has_member_limit, member_limit, subject, location, latitude, longitude, visibility)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id
		`
		err := s.db.QueryRowContext(ctx, query, group.Name, group.Description, group.HasMemberLimit, *group.MemberLimit, group.Subject, group.Location, group.Latitude, group.Longitude, group.Visibility).Scan(&id)
		if err != nil {
			return 0, err
		}
	}

	return id, nil
}

func (s *GroupRepository) GetGroupByID(ctx context.Context, id int) (Group, error) {
	query := `
		SELECT id, name, description, has_member_limit, member_limit, subject, location, visibility, latitude, longitude, created_at, updated_at
		FROM groups
		WHERE id = $1
	`

	row := s.db.QueryRowContext(ctx, query, id)

	var group Group
	var memberLimit sql.NullInt64
	var latitude, longitude sql.NullFloat64
	err := row.Scan(&group.ID, &group.Name, &group.Description, &group.HasMemberLimit, &memberLimit, &group.Subject, &group.Location, &group.Visibility, &latitude, &longitude, &group.CreatedAt, &group.UpdatedAt)
	if err != nil {
		return Group{}, err
	}

	if memberLimit.Valid {
		group.MemberLimit = new(int)
		*group.MemberLimit = int(memberLimit.Int64)
	} else {
		group.MemberLimit = nil
	}

	if latitude.Valid {
		val := latitude.Float64
		group.Latitude = &val
	}

	if longitude.Valid {
		val := longitude.Float64
		group.Longitude = &val
	}

	return group, nil
}

func (s *GroupRepository) GetUserGroups(ctx context.Context, userID int) ([]Group, error) {
	query := `
		SELECT g.id, g.name, g.description, g.has_member_limit, g.member_limit, g.subject, g.location, g.visibility, g.latitude, g.longitude, g.created_at
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
		var latitude, longitude sql.NullFloat64
		err := rows.Scan(&group.ID, &group.Name, &group.Description, &group.HasMemberLimit, &memberLimit, &group.Subject, &group.Location, &group.Visibility, &latitude, &longitude, &group.CreatedAt)
		if err != nil {
			return nil, err
		}

		if memberLimit.Valid {
			group.MemberLimit = new(int)
			*group.MemberLimit = int(memberLimit.Int64)
		} else {
			group.MemberLimit = nil
		}

		if latitude.Valid {
			val := latitude.Float64
			group.Latitude = &val
		}

		if longitude.Valid {
			val := longitude.Float64
			group.Longitude = &val
		}

		groups = append(groups, group)
	}

	return groups, nil
}

func (s *GroupRepository) SearchGroup(ctx context.Context, searchQuery string) ([]Group, error) {
	query := `
       SELECT id, name, has_member_limit, member_limit, description, subject, location, visibility, latitude, longitude, created_at, updated_at FROM groups WHERE name ILIKE $1 || '%' AND visibility = 'public' ORDER BY name ASC`

	rows, err := s.db.QueryContext(ctx, query, searchQuery)
	if err != nil {
		return nil, err
	}

	var groups []Group
	for rows.Next() {
		var group Group
		var memberLimit sql.NullInt64
		var latitude, longitude sql.NullFloat64
		err := rows.Scan(&group.ID, &group.Name, &group.HasMemberLimit, &memberLimit, &group.Description, &group.Subject, &group.Location, &group.Visibility, &latitude, &longitude, &group.CreatedAt, &group.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if memberLimit.Valid {
			group.MemberLimit = new(int)
			*group.MemberLimit = int(memberLimit.Int64)
		} else {
			group.MemberLimit = nil
		}

		if latitude.Valid {
			val := latitude.Float64
			group.Latitude = &val
		}

		if longitude.Valid {
			val := longitude.Float64
			group.Longitude = &val
		}

		groups = append(groups, group)
	}

	return groups, nil
}

func (s *GroupRepository) GetJoinedGroups(ctx context.Context, userID int) ([]Group, error) {
	query := `
		SELECT g.id, g.name, g.description, g.has_member_limit, g.member_limit, g.subject, g.location, g.visibility, g.latitude, g.longitude
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
		var latitude, longitude sql.NullFloat64
		err := rows.Scan(&group.ID, &group.Name, &group.Description, &group.HasMemberLimit, &memberLimit, &group.Subject, &group.Location, &group.Visibility, &latitude, &longitude)
		if err != nil {
			return nil, err
		}

		if memberLimit.Valid {
			group.MemberLimit = new(int)
			*group.MemberLimit = int(memberLimit.Int64)
		} else {
			group.MemberLimit = nil
		}

		if latitude.Valid {
			val := latitude.Float64
			group.Latitude = &val
		}

		if longitude.Valid {
			val := longitude.Float64
			group.Longitude = &val
		}

		groups = append(groups, group)
	}

	return groups, nil
}

func (s *GroupRepository) GetGroupsWithLocation(ctx context.Context) ([]Group, error) {
	query := `
		SELECT id, name, description, has_member_limit, member_limit, subject, location, visibility, latitude, longitude
		FROM groups
		WHERE visibility = 'public' AND latitude IS NOT NULL AND longitude IS NOT NULL
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var groups []Group
	for rows.Next() {
		var group Group
		var memberLimit sql.NullInt64
		var latitude, longitude sql.NullFloat64

		err := rows.Scan(
			&group.ID,
			&group.Name,
			&group.Description,
			&group.HasMemberLimit,
			&memberLimit,
			&group.Subject,
			&group.Location,
			&group.Visibility,
			&latitude,
			&longitude,
		)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		if memberLimit.Valid {
			group.MemberLimit = new(int)
			*group.MemberLimit = int(memberLimit.Int64)
		} else {
			group.MemberLimit = nil
		}

		if latitude.Valid {
			val := latitude.Float64
			group.Latitude = &val
		} else {
			group.Latitude = nil
		}

		if longitude.Valid {
			val := longitude.Float64
			group.Longitude = &val
		} else {
			group.Longitude = nil
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

type GroupWithDistance struct {
	Group
	Distance float64 `json:"distance"`
}

func (s *GroupRepository) GetGroupsWithDistance(ctx context.Context, latitude float64, longitude float64) ([]GroupWithDistance, error) {

	groups, err := s.GetGroupsWithLocation(ctx)
	if err != nil {
		return nil, err
	}

	var groupsWithDistance []GroupWithDistance

	for _, group := range groups {
		distance := haversine(*group.Latitude, *group.Longitude, latitude, longitude)
		groupsWithDistance = append(groupsWithDistance, GroupWithDistance{Group: group, Distance: distance})
	}

	return groupsWithDistance, nil
}

// Haversine formula to calculate the distance between two points on the Earth's surface
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth radius in km
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	lat1 = lat1 * math.Pi / 180
	lat2 = lat2 * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1)*math.Cos(lat2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
