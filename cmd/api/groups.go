package main

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/RakibulBh/studygroup-backend/internal/store"
	"github.com/go-chi/chi/v5"
)

// Create
type CreateGroupRequest struct {
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	HasMemberLimit bool    `json:"has_member_limit"`
	MemberLimit    int     `json:"member_limit"`
	Subject        string  `json:"subject"`
	Location       string  `json:"location"`
	Visibility     string  `json:"visibility"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
}

func (app *application) CreateGroup(w http.ResponseWriter, r *http.Request) {

	// Get user id from context
	user := r.Context().Value(userCtx).(store.User)

	// Parse the request
	var payload CreateGroupRequest
	err := app.readJSON(r, &payload)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Verify empty data
	if payload.Name == "" || payload.Description == "" || payload.Subject == "" || payload.Location == "" {
		app.badRequestResponse(w, r, errors.New("invalid request"))
		return
	}

	// Verify values
	if payload.Visibility != "public" && payload.Visibility != "private" {
		app.badRequestResponse(w, r, errors.New("invalid visibility value"))
		return
	}

	// Verify lengths
	if len(payload.Name) > 100 || len(payload.Description) > 500 || len(payload.Subject) > 100 || len(payload.Location) > 100 {
		app.badRequestResponse(w, r, errors.New("invalid character limit"))
		return
	}

	// Handle member limit - set to nil if 0 or HasMemberLimit is false
	var memberLimit *int
	if payload.HasMemberLimit && payload.MemberLimit > 0 {
		memberLimit = new(int)
		*memberLimit = payload.MemberLimit
	}

	ctx := r.Context()

	// Check if both latitude and longitude are 0, store as null if so
	var lat, long *float64
	if payload.Latitude != 0 || payload.Longitude != 0 {
		lat = new(float64)
		long = new(float64)
		*lat = payload.Latitude
		*long = payload.Longitude
	}

	group := &store.Group{
		Name:           payload.Name,
		Description:    payload.Description,
		HasMemberLimit: payload.HasMemberLimit,
		MemberLimit:    memberLimit,
		Subject:        payload.Subject,
		Latitude:       lat,
		Longitude:      long,
		Location:       payload.Location,
		Visibility:     payload.Visibility,
	}

	// Create group
	id, err := app.store.GroupRepository.CreateGroup(ctx, group)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	group.ID = id

	// Make admin
	err = app.store.GroupMembershipManagement.MakeAdmin(ctx, group.ID, user.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, "Group created successfully", nil)
}

// query
type GroupWithMembers struct {
	store.Group
	Members []store.User `json:"members"`
}

func (app *application) GetUserGroups(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userCtx).(store.User)
	groups, err := app.store.GroupRepository.GetUserGroups(r.Context(), user.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	var groupWithMembers []GroupWithMembers

	// For each group get the members
	for _, group := range groups {
		members, err := app.store.GroupMembership.GetGroupMembers(r.Context(), group.ID)
		if err != nil {
			app.internalServerErrorResponse(w, r, err)
			return
		}
		groupWithMembers = append(groupWithMembers, GroupWithMembers{
			Group:   group,
			Members: members,
		})
	}

	app.writeJSON(w, http.StatusOK, "Groups fetched successfully", groupWithMembers)
}

type GroupWithMetadata struct {
	store.Group
	JoinRequested bool         `json:"join_requested"`
	Members       []store.User `json:"members"`
}

func (app *application) SearchGroup(w http.ResponseWriter, r *http.Request) {
	// Get search query from path
	searchQuery := chi.URLParam(r, "search_query")

	ctx := r.Context()
	user := r.Context().Value(userCtx).(store.User)

	// Search for group
	groups, err := app.store.GroupRepository.SearchGroup(ctx, searchQuery)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	var groupsWithMetadata []GroupWithMetadata

	// Check if user is a member of any of the group and if they have requested
	for _, group := range groups {
		isMember, err := app.store.GroupMembership.IsMember(ctx, group.ID, user.ID)
		if err != nil {
			app.internalServerErrorResponse(w, r, err)
			return
		}

		// If user is not a member, check if they have requested to join
		if !isMember {
			hasJoinRequested, err := app.store.GroupJoinRequests.IsJoinRequested(ctx, group.ID, user.ID)
			if err != nil {
				app.internalServerErrorResponse(w, r, err)
				return
			}
			members, err := app.store.GroupMembership.GetGroupMembers(ctx, group.ID)
			if err != nil {
				app.internalServerErrorResponse(w, r, err)
				return
			}
			groupsWithMetadata = append(groupsWithMetadata, GroupWithMetadata{
				Group:         group,
				JoinRequested: hasJoinRequested,
				Members:       members,
			})
		}
	}

	app.writeJSON(w, http.StatusOK, "Group searched successfully", groupsWithMetadata)
}

// Get a group by id
func (app *application) GetGroup(w http.ResponseWriter, r *http.Request) {
	// Get group id from path
	groupID := chi.URLParam(r, "id")

	// Convert groupID to int
	groupIDInt, err := strconv.Atoi(groupID)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	// Get group from store
	group, err := app.store.GroupRepository.GetGroupByID(ctx, groupIDInt)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, "Group fetched successfully", group)
}

// Get nearby groups
type NearbyGroupsRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type NearbyGroupWithMembers struct {
	store.GroupWithDistance
	Members []store.User `json:"members"`
}

func (app *application) GetNearbyGroups(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var payload NearbyGroupsRequest
	err := app.readJSON(r, &payload)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Convert latitude and longitude to float64
	latitudeFloat := payload.Latitude
	longitudeFloat := payload.Longitude

	ctx := r.Context()
	user := r.Context().Value(userCtx).(store.User)

	// Get all groups with their distances
	groups, err := app.store.GroupRepository.GetGroupsWithDistance(ctx, latitudeFloat, longitudeFloat)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	// sort by distance
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Distance < groups[j].Distance
	})

	if len(groups) > 5 {
		groups = groups[:5]
	}

	var nearbyGroupsWithMembers []NearbyGroupWithMembers

	// For each group add a members too
	for _, group := range groups {
		// Check if user is a member of the group, if so then do not show that group
		isMember, err := app.store.GroupMembership.IsMember(ctx, group.ID, user.ID)
		if err != nil {
			app.internalServerErrorResponse(w, r, err)
			return
		}
		if isMember {
			continue
		}

		// Get members of the group
		members, err := app.store.GroupMembership.GetGroupMembers(ctx, group.ID)
		if err != nil {
			app.internalServerErrorResponse(w, r, err)
			return
		}

		// Check the distance
		if group.Distance > 5 {
			continue
		}

		nearbyGroupsWithMembers = append(nearbyGroupsWithMembers, NearbyGroupWithMembers{
			GroupWithDistance: group,
			Members:           members,
		})
	}

	app.writeJSON(w, http.StatusOK, "Nearby groups fetched successfully", nearbyGroupsWithMembers)
}

// Membership

func (app *application) GetGroupMembers(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "id")

	groupIDInt, err := strconv.Atoi(groupID)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	group, err := app.store.GroupRepository.GetGroupByID(ctx, groupIDInt)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	members, err := app.store.GroupMembership.GetGroupMembers(ctx, group.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	user := r.Context().Value(userCtx).(store.User)

	// Check if the user is allowed to see this group
	isMember, err := app.store.GroupMembership.IsMember(ctx, group.ID, user.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
	if group.Visibility == "private" && !isMember {
		app.writeJSON(w, http.StatusForbidden, "not allowed", nil)
		return
	}

	app.writeJSON(w, http.StatusOK, "Group members fetched successfully", members)
}

// Joining and leaving groups
func (app *application) JoinGroup(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userCtx).(store.User)

	groupID := chi.URLParam(r, "id")

	groupIDInt, err := strconv.Atoi(groupID)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	err = app.store.GroupJoinRequests.JoinRequest(ctx, groupIDInt, user.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, "Group join request sent successfully", nil)
}

type GroupJoinRequest struct {
	store.GroupJoinRequest
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	University string `json:"university"`
	Subject    string `json:"subject"`
}

func (app *application) GetGroupJoinRequests(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "id")

	ctx := r.Context()

	groupIDInt, err := strconv.Atoi(groupID)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := r.Context().Value(userCtx).(store.User)

	// Check if the user is the admin of the group
	isAdmin, err := app.store.GroupMembership.IsAdmin(ctx, groupIDInt, user.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
	if !isAdmin {
		app.writeJSON(w, http.StatusForbidden, "not allowed", nil)
		return
	}

	joinRequests, err := app.store.GroupJoinRequests.GetJoinRequests(ctx, groupIDInt)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	var groupJoinRequests []GroupJoinRequest
	for _, joinRequest := range joinRequests {
		user, err := app.store.User.GetUserByID(ctx, joinRequest.UserID)
		if err != nil {
			app.internalServerErrorResponse(w, r, err)
			return
		}
		groupJoinRequests = append(groupJoinRequests, GroupJoinRequest{
			GroupJoinRequest: joinRequest,
			FirstName:        user.FirstName,
			LastName:         user.LastName,
			Email:            user.Email,
			University:       user.University,
			Subject:          user.Subject,
		})
	}

	app.writeJSON(w, http.StatusOK, "Join requests fetched successfully", groupJoinRequests)
}

// Approve join request
type ApproveJoinRequestRequest struct {
	UserID  int  `json:"user_id"`
	Approve bool `json:"approve"`
}

func (app *application) ApproveJoinRequest(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "id")

	var payload ApproveJoinRequestRequest
	err := app.readJSON(r, &payload)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	groupIDInt, err := strconv.Atoi(groupID)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	user := r.Context().Value(userCtx).(store.User)

	// Check if the user is the admin of the group, is he allowed to approve/reject?
	isAdmin, err := app.store.GroupMembership.IsAdmin(ctx, groupIDInt, user.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
	if !isAdmin {
		app.writeJSON(w, http.StatusForbidden, "not allowed", nil)
		return
	}

	if payload.Approve {
		err = app.store.GroupJoinRequests.ApproveJoinRequest(ctx, groupIDInt, payload.UserID)
		if err != nil {
			app.internalServerErrorResponse(w, r, err)
			return
		}
	} else {
		err = app.store.GroupJoinRequests.RejectJoinRequest(ctx, groupIDInt, payload.UserID)
		if err != nil {
			app.internalServerErrorResponse(w, r, err)
			return
		}
	}

	app.writeJSON(w, http.StatusOK, "Join request approved successfully", nil)
}

func (app *application) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "id")

	groupIDInt, err := strconv.Atoi(groupID)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	user := r.Context().Value(userCtx).(store.User)

	err = app.store.GroupMembershipManagement.LeaveGroup(ctx, groupIDInt, user.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, "Group left successfully", nil)
}

// Get the user's joined groups
func (app *application) GetJoinedGroups(w http.ResponseWriter, r *http.Request) {

	// Get user id from context
	user := r.Context().Value(userCtx).(store.User)

	// Get group id from path
	ctx := r.Context()

	// Get joined groups
	groups, err := app.store.GroupRepository.GetJoinedGroups(ctx, user.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, "Joined groups fetched successfully", groups)
}

func (app *application) IsAdmin(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "id")

	groupIDInt, err := strconv.Atoi(groupID)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	user := r.Context().Value(userCtx).(store.User)

	isAdmin, err := app.store.GroupMembership.IsAdmin(ctx, groupIDInt, user.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, "Is admin", isAdmin)
}

// Invite user to group
type InviteUserToGroupRequest struct {
	Email string `json:"email"`
}

func (app *application) InviteUserToGroup(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "id")

	var payload InviteUserToGroupRequest
	err := app.readJSON(r, &payload)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	groupIDInt, err := strconv.Atoi(groupID)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	// Get user id from context
	user := r.Context().Value(userCtx).(store.User)

	// Check if user is admin
	isAdmin, err := app.store.GroupMembership.IsAdmin(ctx, groupIDInt, user.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
	if !isAdmin {
		app.writeJSON(w, http.StatusForbidden, "not allowed", nil)
		return
	}

	// Check if user exists
	invitedUser, err := app.store.User.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Check if user is already a member
	isMember, err := app.store.GroupMembership.IsMember(ctx, groupIDInt, invitedUser.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
	if isMember {
		app.badRequestResponse(w, r, errors.New("user is already a member"))
		return
	}

	// Invite user
	err = app.store.GroupInvitations.InviteUserToGroup(ctx, groupIDInt, invitedUser.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, "User invited successfully", nil)
}

type UserInvitationsResponse struct {
	GroupID   int       `json:"group_id"`
	GroupName string    `json:"group_name"`
	InvitedAt time.Time `json:"invited_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// get user invitations
func (app *application) GetUserInvitations(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userCtx).(store.User)

	ctx := r.Context()

	invitations, err := app.store.GroupInvitations.GetInvitations(ctx, user.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	// Find group name with ID
	var userInvitations []UserInvitationsResponse
	for _, invitation := range invitations {
		group, err := app.store.GroupRepository.GetGroupByID(ctx, invitation.GroupID)
		if err != nil {
			app.internalServerErrorResponse(w, r, err)
			return
		}
		userInvitations = append(userInvitations, UserInvitationsResponse{
			GroupID:   invitation.GroupID,
			GroupName: group.Name,
			InvitedAt: invitation.InvitedAt,
			ExpiresAt: invitation.ExpiresAt,
		})
	}

	app.writeJSON(w, http.StatusOK, "User invitations fetched successfully", userInvitations)
}

// accept invitation
func (app *application) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "id")

	groupIDInt, err := strconv.Atoi(groupID)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()
	user := r.Context().Value(userCtx).(store.User)

	err = app.store.GroupInvitations.AcceptInvitation(ctx, user.ID, groupIDInt)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, "Invitation accepted successfully", nil)
}

// resolve invitation
type ResolveInvitationRequest struct {
	GroupID int  `json:"group_id"`
	Accept  bool `json:"accept"`
}

func (app *application) ResolveInvitation(w http.ResponseWriter, r *http.Request) {
	var payload ResolveInvitationRequest
	err := app.readJSON(r, &payload)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()
	user := r.Context().Value(userCtx).(store.User)

	// Check if user is already a member
	isMember, err := app.store.GroupMembership.IsMember(ctx, payload.GroupID, user.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
	if isMember {
		app.badRequestResponse(w, r, errors.New("user is already a member"))
		return
	}

	if payload.Accept {
		err = app.store.GroupInvitations.AcceptInvitation(ctx, user.ID, payload.GroupID)
	} else {
		err = app.store.GroupInvitations.RejectInvitation(ctx, user.ID, payload.GroupID)
	}
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	action := "accepted"
	if !payload.Accept {
		action = "rejected"
	}

	app.writeJSON(w, http.StatusOK, fmt.Sprintf("Invitation %s successfully", action), nil)
}

// delete group
func (app *application) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "id")

	groupIDInt, err := strconv.Atoi(groupID)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()
	user := r.Context().Value(userCtx).(store.User)

	// Check if user is admin
	isAdmin, err := app.store.GroupMembership.IsAdmin(ctx, groupIDInt, user.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
	if !isAdmin {
		app.writeJSON(w, http.StatusForbidden, "not allowed", nil)
		return
	}

	err = app.store.GroupRepository.DeleteGroup(ctx, groupIDInt)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, "Group deleted successfully", nil)
}

// kick user from group
type KickUserFromGroupRequest struct {
	UserID int `json:"user_id"`
}

func (app *application) KickUserFromGroup(w http.ResponseWriter, r *http.Request) {
	var payload KickUserFromGroupRequest
	err := app.readJSON(r, &payload)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	groupID := chi.URLParam(r, "id")

	groupIDInt, err := strconv.Atoi(groupID)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()
	user := r.Context().Value(userCtx).(store.User)

	// You can't kick yourself out
	if payload.UserID == user.ID {
		app.badRequestResponse(w, r, errors.New("you can't kick yourself out"))
		return
	}

	// Check if user is admin
	isAdmin, err := app.store.GroupMembership.IsAdmin(ctx, groupIDInt, user.ID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}
	if !isAdmin {
		app.writeJSON(w, http.StatusForbidden, "not allowed", nil)
		return
	}

	err = app.store.GroupMembershipManagement.KickUserFromGroup(ctx, groupIDInt, payload.UserID)
	if err != nil {
		app.internalServerErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, "User kicked from group successfully", nil)
}
