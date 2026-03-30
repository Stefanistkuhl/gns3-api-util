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
)

// ListRoles returns all roles stored in etcd
//
//	@Summary		List all roles
//	@Description	Returns all roles and their permission scopes from etcd
//	@Tags			auth, rbac
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	models.ListRolesResponse
//	@Failure		500	{object}	helpers.APIErrorResponse
//	@Router			/api/v1/auth/roles [get]
func (m *Master) ListRoles(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	roles, err := m.Store.ListRoles(ctx)
	if err != nil {
		helpers.WriteAPIError(w, "Failed to list roles", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}

	infos := make([]models.RoleInfo, 0, len(roles))
	for _, role := range roles {
		infos = append(infos, roleToInfo(role))
	}

	if err := helpers.WriteJSON(w, models.ListRolesResponse{Roles: infos, Count: len(infos)}); err != nil {
		m.Logger.Error("Failed to write list roles response", "err", err)
	}
}

// GetRole returns a single role's details
//
//	@Summary		Get role
//	@Description	Returns a role's description and permission scopes from etcd
//	@Tags			auth, rbac
//	@Produce		json
//	@Security		BearerAuth
//	@Param			role_name	path		string	true	"Role name"
//	@Success		200			{object}	models.RoleInfo
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/auth/roles/{role_name} [get]
func (m *Master) GetRole(w http.ResponseWriter, r *http.Request) {
	roleName := chi.URLParam(r, "role_name")
	if roleName == "" {
		helpers.WriteAPIError(w, "role_name required", helpers.ErrCodeInvalidInput, "role_name path parameter is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	role, err := m.Store.GetRole(ctx, roleName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			helpers.WriteAPIError(w, "Role not found", helpers.ErrCodeNotFound, fmt.Sprintf("role %q not found", roleName), http.StatusNotFound)
			return
		}
		helpers.WriteAPIError(w, "Failed to get role", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := helpers.WriteJSON(w, roleToInfo(role)); err != nil {
		m.Logger.Error("Failed to write get role response", "err", err)
	}
}

// CreateRole creates a new role with given scopes
//
//	@Summary		Create role
//	@Description	Creates a new role in etcd with specified permission scopes
//	@Tags			auth, rbac
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		models.CreateRoleRequest	true	"Role definition"
//	@Success		201		{object}	models.RoleInfo
//	@Failure		400		{object}	helpers.APIErrorResponse
//	@Failure		500		{object}	helpers.APIErrorResponse
//	@Router			/api/v1/auth/roles [post]
func (m *Master) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteAPIError(w, "Invalid request body", helpers.ErrCodeInvalidRequest, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		helpers.WriteAPIError(w, "name required", helpers.ErrCodeInvalidInput, "role name is required", http.StatusBadRequest)
		return
	}

	scopes, err := scopeInfosToProto(req.Scopes)
	if err != nil {
		helpers.WriteAPIError(w, "Invalid scope", helpers.ErrCodeInvalidInput, err.Error(), http.StatusBadRequest)
		return
	}

	role := &pb.Role{
		Name:        req.Name,
		Description: req.Description,
		Scopes:      scopes,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := m.Store.CreateRole(ctx, role); err != nil {
		helpers.WriteAPIError(w, "Failed to create role", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := helpers.WriteJSON(w, roleToInfo(role)); err != nil {
		m.Logger.Error("Failed to write create role response", "err", err)
	}
}

// UpdateRole replaces a role's scopes
//
//	@Summary		Update role
//	@Description	Replaces a role's description and permission scopes in etcd
//	@Tags			auth, rbac
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			role_name	path		string						true	"Role name"
//	@Param			request		body		models.UpdateRoleRequest	true	"Updated role data"
//	@Success		200			{object}	models.RoleInfo
//	@Failure		400			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/auth/roles/{role_name} [put]
func (m *Master) UpdateRole(w http.ResponseWriter, r *http.Request) {
	roleName := chi.URLParam(r, "role_name")
	if roleName == "" {
		helpers.WriteAPIError(w, "role_name required", helpers.ErrCodeInvalidInput, "role_name path parameter is required", http.StatusBadRequest)
		return
	}

	var req models.UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteAPIError(w, "Invalid request body", helpers.ErrCodeInvalidRequest, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	existing, err := m.Store.GetRole(ctx, roleName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			helpers.WriteAPIError(w, "Role not found", helpers.ErrCodeNotFound, fmt.Sprintf("role %q not found", roleName), http.StatusNotFound)
			return
		}
		helpers.WriteAPIError(w, "Failed to get role", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}

	scopes, err := scopeInfosToProto(req.Scopes)
	if err != nil {
		helpers.WriteAPIError(w, "Invalid scope", helpers.ErrCodeInvalidInput, err.Error(), http.StatusBadRequest)
		return
	}

	existing.Description = req.Description
	existing.Scopes = scopes

	if err := m.Store.UpdateRole(ctx, existing); err != nil {
		helpers.WriteAPIError(w, "Failed to update role", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := helpers.WriteJSON(w, roleToInfo(existing)); err != nil {
		m.Logger.Error("Failed to write update role response", "err", err)
	}
}

// DeleteRole removes a role from etcd
//
//	@Summary		Delete role
//	@Description	Removes a role and all its scopes from etcd
//	@Tags			auth, rbac
//	@Produce		json
//	@Security		BearerAuth
//	@Param			role_name	path		string	true	"Role name"
//	@Success		200			{string}	string	"OK"
//	@Failure		400			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/auth/roles/{role_name} [delete]
func (m *Master) DeleteRole(w http.ResponseWriter, r *http.Request) {
	roleName := chi.URLParam(r, "role_name")
	if roleName == "" {
		helpers.WriteAPIError(w, "role_name required", helpers.ErrCodeInvalidInput, "role_name path parameter is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := m.Store.DeleteRole(ctx, roleName); err != nil {
		helpers.WriteAPIError(w, "Failed to delete role", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := w.Write([]byte("OK")); err != nil {
		m.Logger.Error("Failed to write delete role response", "err", err)
	}
}
