package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/state/pb"
	"github.com/0xveya/gns3util/pkg/web/helpers"
	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AssignRole appends a role to a user's existing role assignments
//
//	@Summary		Assign role to user
//	@Description	Adds a role to a user's existing role list in etcd (idempotent)
//	@Tags			auth, rbac
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id	path		string						true	"User ID"
//	@Param			request	body		models.AssignRoleRequest	true	"Role to assign"
//	@Success		200		{object}	models.UserInfo
//	@Failure		400		{object}	helpers.APIErrorResponse
//	@Failure		404		{object}	helpers.APIErrorResponse
//	@Failure		500		{object}	helpers.APIErrorResponse
//	@Router			/api/v1/auth/users/{user_id}/roles [post]
func (m *Master) AssignRole(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	if userID == "" {
		helpers.WriteAPIError(w, "user_id required", helpers.ErrCodeInvalidInput, "user_id path parameter is required", http.StatusBadRequest)
		return
	}

	var req models.AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteAPIError(w, "Invalid request body", helpers.ErrCodeInvalidRequest, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Role == "" {
		helpers.WriteAPIError(w, "role required", helpers.ErrCodeInvalidInput, "role field is required in request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if _, err := m.Store.GetRole(ctx, req.Role); err != nil {
		if strings.Contains(err.Error(), "not found") {
			helpers.WriteAPIError(w, "Role not found", helpers.ErrCodeNotFound, fmt.Sprintf("role %q not found", req.Role), http.StatusNotFound)
			return
		}
		helpers.WriteAPIError(w, "Failed to validate role", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}

	perms, err := m.Store.GetUserPermissions(ctx, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			helpers.WriteAPIError(w, "User not found", helpers.ErrCodeNotFound, fmt.Sprintf("user %q not found", userID), http.StatusNotFound)
			return
		}
		helpers.WriteAPIError(w, "Failed to get user", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, existing := range perms.RoleNames {
		if existing == req.Role {
			info, infoErr := m.userInfoFromPermissions(ctx, perms)
			if infoErr != nil {
				helpers.WriteAPIError(w, "Failed to resolve effective permissions", helpers.ErrCodeInternal, infoErr.Error(), http.StatusInternalServerError)
				return
			}
			if writeErr := helpers.WriteJSON(w, info); writeErr != nil {
				m.Logger.Error("Failed to write assign role response", "err", writeErr)
			}
			return
		}
	}

	perms.RoleNames = append(perms.RoleNames, req.Role)
	if putErr := m.Store.PutUserPermissions(ctx, userID, perms.RoleNames, perms.DenyScopes); putErr != nil {
		helpers.WriteAPIError(w, "Failed to assign role", helpers.ErrCodeInternal, putErr.Error(), http.StatusInternalServerError)
		return
	}

	info, err := m.userInfoFromPermissions(ctx, &pb.UserPermissions{
		UserId:     userID,
		RoleNames:  perms.RoleNames,
		UpdatedAt:  timestamppb.Now(),
		DenyScopes: perms.DenyScopes,
	})
	if err != nil {
		helpers.WriteAPIError(w, "Failed to resolve effective permissions", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := helpers.WriteJSON(w, info); err != nil {
		m.Logger.Error("Failed to write assign role response", "err", err)
	}
}

// ListUsers returns all users stored in etcd
//
//	@Summary		List all users
//	@Description	Returns all users and their assigned roles from etcd
//	@Tags			auth, rbac
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	models.ListUsersResponse
//	@Failure		500	{object}	helpers.APIErrorResponse
//	@Router			/api/v1/auth/users [get]
func (m *Master) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	users, err := m.Store.ListUsers(ctx)
	if err != nil {
		helpers.WriteAPIError(w, "Failed to list users", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}

	infos := make([]models.UserInfo, 0, len(users))
	for _, u := range users {
		info, infoErr := m.userInfoFromPermissions(ctx, u)
		if infoErr != nil {
			helpers.WriteAPIError(w, "Failed to resolve effective permissions", helpers.ErrCodeInternal, infoErr.Error(), http.StatusInternalServerError)
			return
		}
		infos = append(infos, info)
	}

	if err := helpers.WriteJSON(w, models.ListUsersResponse{Users: infos, Count: len(infos)}); err != nil {
		m.Logger.Error("Failed to write list users response", "err", err)
	}
}

// GetUser returns a single user's roles and effective scopes
//
//	@Summary		Get user
//	@Description	Returns a user's assigned roles and effective permission scopes from etcd
//	@Tags			auth, rbac
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id	path		string	true	"User ID"
//	@Success		200		{object}	models.UserInfo
//	@Failure		404		{object}	helpers.APIErrorResponse
//	@Failure		500		{object}	helpers.APIErrorResponse
//	@Router			/api/v1/auth/users/{user_id} [get]
func (m *Master) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	if userID == "" {
		helpers.WriteAPIError(w, "user_id required", helpers.ErrCodeInvalidInput, "user_id path parameter is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	perms, err := m.Store.GetUserPermissions(ctx, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			helpers.WriteAPIError(w, "User not found", helpers.ErrCodeNotFound, fmt.Sprintf("user %q not found", userID), http.StatusNotFound)
			return
		}
		helpers.WriteAPIError(w, "Failed to get user", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}

	info, err := m.userInfoFromPermissions(ctx, perms)
	if err != nil {
		helpers.WriteAPIError(w, "Failed to resolve effective permissions", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := helpers.WriteJSON(w, info); err != nil {
		m.Logger.Error("Failed to write get user response", "err", err)
	}
}

// DeleteUser removes a user's permission record from etcd
//
//	@Summary		Delete user
//	@Description	Removes a user and all their role assignments from etcd
//	@Tags			auth, rbac
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id	path		string	true	"User ID"
//	@Success		200		{string}	string	"OK"
//	@Failure		400		{object}	helpers.APIErrorResponse
//	@Failure		500		{object}	helpers.APIErrorResponse
//	@Router			/api/v1/auth/users/{user_id} [delete]
func (m *Master) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	if userID == "" {
		helpers.WriteAPIError(w, "user_id required", helpers.ErrCodeInvalidInput, "user_id path parameter is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := m.Store.DeleteUser(ctx, userID); err != nil {
		helpers.WriteAPIError(w, "Failed to delete user", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := w.Write([]byte("OK")); err != nil {
		m.Logger.Error("Failed to write delete user response", "err", err)
	}
}

// CreateUser creates a user in etcd
//
//	@Summary		Create user
//	@Description	Creates a user in etcd with no assigned roles.
//	@Tags			auth, rbac
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user	body		models.CreateUserRequest	true	"User Creation Payload"
//	@Success		200		{object}	models.UserInfo				"OK"
//	@Failure		400		{object}	helpers.APIErrorResponse
//	@Failure		500		{object}	helpers.APIErrorResponse
//	@Router			/api/v1/auth/users [post]
func (m *Master) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteAPIError(w, "Invalid request body", helpers.ErrCodeInvalidRequest, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		helpers.WriteAPIError(w, "name required", helpers.ErrCodeInvalidInput, "name field is required in request body", http.StatusBadRequest)
		return
	}
	user, err := m.Store.ListUsers(r.Context())
	if err != nil {
		helpers.WriteAPIError(w, "Failed to validate user name", helpers.ErrCodeInternal, fmt.Sprintf("error fetching users for validation: %v", err), http.StatusInternalServerError)
		return
	}
	for i := range user {
		if user[i].UserId == req.Name {
			helpers.WriteAPIError(w, "User already exists", helpers.ErrCodeInvalidInput, fmt.Sprintf("user with name %q already exists", req.Name), http.StatusBadRequest)
			return
		}
	}

	roles, err := m.Store.ListRoles(r.Context())
	if err != nil {
		helpers.WriteAPIError(w, "Failed to validate roles", helpers.ErrCodeInternal, fmt.Sprintf("error fetching roles for validation: %v", err), http.StatusInternalServerError)
		return
	}
	for i := range req.Roles {
		found := false
		for _, r := range roles {
			if r.Name == req.Roles[i] {
				found = true
				break
			}
		}
		if !found {
			helpers.WriteAPIError(w, "Invalid role", helpers.ErrCodeInvalidInput, fmt.Sprintf("role %q does not exist", req.Roles[i]), http.StatusBadRequest)
			return
		}
	}

	denyScopes, err := scopeInfosToProto(req.DenyScopes)
	if err != nil {
		helpers.WriteAPIError(w, "Invalid deny scopes", helpers.ErrCodeInvalidInput, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if createErr := m.Store.CreateUser(ctx, req.Name, req.Roles, denyScopes); createErr != nil {
		helpers.WriteAPIError(w, "Failed to create user", helpers.ErrCodeInternal, createErr.Error(), http.StatusInternalServerError)
		return
	}

	userInfo, err := m.userInfoFromPermissions(ctx, &pb.UserPermissions{
		UserId:     req.Name,
		RoleNames:  req.Roles,
		UpdatedAt:  timestamppb.Now(),
		DenyScopes: denyScopes,
	})
	if err != nil {
		helpers.WriteAPIError(w, "Failed to resolve effective permissions", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}
	writeErr := helpers.WriteJSON(w, userInfo)
	if writeErr != nil {
		m.Logger.Error("Failed to write response", "err", writeErr, "user_id", req.Name)
		helpers.WriteAPIError(w, "failed to write response", helpers.ErrCodeInternal, writeErr.Error(), http.StatusInternalServerError)
		return
	}
}
