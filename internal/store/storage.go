package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Errors
var (
	ErrNotFound = errors.New("not found")
	ErrNoRows   = errors.New("user not found")
	ErrConflict = errors.New("conflict")
	ErrInternal = errors.New("internal server error")
	ErrInvalid  = errors.New("invalid input")
)

type Storage struct {
	Auth interface {
		HashPassword(password string) (string, error)
		Register(ctx context.Context, request RegisterRequest) error
		VerifyPassword(password string, hash string) (bool, error)
		GenerateJWT(userID int, expiresAt time.Time, secret string) (string, error)
		VerifyToken(tokenString string, secret string) (*jwt.Token, error)
		StoreRefreshToken(ctx context.Context, userID int, token string, expiresAt time.Time) error
		RefreshToken(ctx context.Context, userID int, secret string, refreshExp time.Duration, accessExp time.Duration) (string, string, error)
		DeleteRefreshTokens(ctx context.Context, userID int) error
	}
	User interface {
		GetUserByID(ctx context.Context, id int) (User, error)
		GetUserByEmail(ctx context.Context, email string) (UserData, error)
	}
	GroupRepository interface {
		CreateGroup(ctx context.Context, group *Group) (int, error)
		GetGroupByID(ctx context.Context, id int) (Group, error)
		GetUserGroups(ctx context.Context, userID int) ([]Group, error)
		SearchGroup(ctx context.Context, searchQuery string) ([]Group, error)
		GetJoinedGroups(ctx context.Context, userID int) ([]Group, error)
		GetGroupsWithDistance(ctx context.Context, latitude float64, longitude float64) ([]GroupWithDistance, error)
		DeleteGroup(ctx context.Context, groupID int) error
	}
	GroupJoinRequests interface {
		JoinRequest(ctx context.Context, groupID int, userID int) error
		GetJoinRequests(ctx context.Context, groupID int) ([]GroupJoinRequest, error)
		IsJoinRequested(ctx context.Context, groupID int, userID int) (bool, error)
		ApproveJoinRequest(ctx context.Context, groupID int, userID int) error
		RejectJoinRequest(ctx context.Context, groupID int, userID int) error
	}
	GroupInvitations interface {
		InviteUserToGroup(ctx context.Context, groupID int, userID int) error
		AcceptInvitation(ctx context.Context, userID int, groupID int) error
		GetInvitations(ctx context.Context, userID int) ([]GroupInvitation, error)
		RejectInvitation(ctx context.Context, userID int, groupID int) error
	}
	GroupMembership interface {
		IsMember(ctx context.Context, groupID int, userID int) (bool, error)
		IsAdmin(ctx context.Context, groupID int, userID int) (bool, error)
		GetMemberCount(ctx context.Context, groupID int) (int, error)
		GetGroupMembers(ctx context.Context, groupID int) ([]User, error)
	}
	GroupMembershipManagement interface {
		LeaveGroup(ctx context.Context, groupID int, userID int) error
		MakeAdmin(ctx context.Context, groupID int, userID int) error
		KickUserFromGroup(ctx context.Context, groupID int, userID int) error
	}
	Session interface {
		CreateStudySession(ctx context.Context, session *StudySession) (int, error)
		DeleteStudySession(ctx context.Context, sessionID int) error
		GetStudySessionByID(ctx context.Context, sessionID int) (StudySession, error)
		GetUserStudySessions(ctx context.Context, userID int) ([]StudySession, error)
		GetGroupStudySessions(ctx context.Context, groupID int) ([]StudySession, error)
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Auth:                      &AuthStore{db: db},
		User:                      &UserStore{db: db},
		GroupRepository:           &GroupRepository{db: db},
		GroupJoinRequests:         &GroupJoinRequestsStore{db: db},
		GroupInvitations:          &GroupInvitationsStore{db: db},
		GroupMembership:           &GroupMembershipStore{db: db},
		GroupMembershipManagement: &GroupMembershipManagementStore{db: db},
		Session:                   &SessionStore{db: db},
	}
}
