package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/state"
	"github.com/0xveya/gns3util/pkg/web/auth"
	"github.com/0xveya/gns3util/pkg/web/helpers"
	"github.com/0xveya/gns3util/pkg/web/middleware"
	"github.com/go-chi/chi/v5"
)

type Master struct {
	IDMgr  *auth.IdentityManager
	Store  *state.StateManager
	TLSDir string
	Logger *slog.Logger
}

// HandleCreateToken creates a new authentication token
//
//	@Summary		Create authentication token
//	@Description	Mints a new JWT token for user authentication
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.CreateTokenRequest	true	"Token creation request"
//	@Success		200		{object}	map[string]string
//	@Failure		400		{object}	helpers.APIErrorResponse
//	@Failure		403		{object}	helpers.APIErrorResponse
//	@Failure		500		{object}	helpers.APIErrorResponse
//	@Router			/api/v1/auth/token [post]
func (m *Master) HandleCreateToken(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteAPIError(w, "Invalid request body", helpers.ErrCodeInvalidRequest, fmt.Sprintf("failed to decode request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.UserID == "" {
		helpers.WriteAPIError(w, "user_id required", helpers.ErrCodeInvalidInput, "user_id is required to create token", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userPerms, err := m.Store.GetUserPermissions(ctx, req.UserID)
	if err != nil {
		helpers.WriteAPIError(w, "Failed to fetch user permissions", helpers.ErrCodeInternal, fmt.Sprintf("failed to fetch user permissions: %v. Run 'ctl create user' first.", err), http.StatusForbidden)
		return
	}

	// Use the first assigned role as the primary role in the JWT.
	// Actual permission checks are always performed against the etcd state, so
	// embedding all roles is not required.
	primaryRole := ""
	if len(userPerms.RoleNames) > 0 {
		primaryRole = userPerms.RoleNames[0]
	}

	claims := auth.NewClaims(req.UserID, primaryRole, []string{}, 365*24*time.Hour)
	token, err := m.IDMgr.Mint(claims)
	if err != nil {
		helpers.WriteAPIError(w, "Failed to mint token", helpers.ErrCodeInternal, fmt.Sprintf("failed to mint token: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"token": token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

// HandleGrantAccess grants a scope to a user
//
//	@Summary		Grant access scope to user
//	@Description	Grants a permission scope to a specific user
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id	query		string	true	"User ID"
//	@Param			scope	query		string	true	"Scope to grant"
//	@Success		200		{string}	string	"OK"
//	@Failure		400		{object}	helpers.APIErrorResponse
//	@Failure		500		{object}	helpers.APIErrorResponse
//	@Router			/api/v1/auth/grant [post]
func (m *Master) HandleGrantAccess(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	scope := r.URL.Query().Get("scope")

	if userID == "" || scope == "" {
		helpers.WriteAPIError(w, "user_id and scope required", helpers.ErrCodeInvalidInput, "both user_id and scope query parameters are required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	existingPerms, err := m.Store.GetUserPermissions(ctx, userID)
	if err != nil && !strings.Contains(err.Error(), "user not found") {
		helpers.WriteAPIError(w, "Failed to fetch existing permissions", helpers.ErrCodeInternal, fmt.Sprintf("failed to fetch existing permissions: %v", err), http.StatusInternalServerError)
		return
	}

	roleNames := existingPerms.GetRoleNames()
	if !slices.Contains(roleNames, scope) {
		roleNames = append(roleNames, scope)
	}

	if err := m.Store.PutUserPermissions(ctx, userID, roleNames, existingPerms.GetDenyScopes()); err != nil {
		helpers.WriteAPIError(w, "Failed to grant access", helpers.ErrCodeInternal, fmt.Sprintf("failed to grant access: %v", err), http.StatusInternalServerError)
		return
	}

	_, writeErr := w.Write([]byte("OK"))
	if writeErr != nil {
		helpers.WriteAPIError(w, "Failed to write response", helpers.ErrCodeInternal, fmt.Sprintf("failed to write response: %v", writeErr), http.StatusInternalServerError)
		return
	}
}

// HandleRevokeAccess revokes all scopes from a user
//
//	@Summary		Revoke user access
//	@Description	Revokes all permission scopes from a user
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id	query		string	true	"User ID"
//	@Success		200		{string}	string	"OK"
//	@Failure		400		{object}	helpers.APIErrorResponse
//	@Failure		500		{object}	helpers.APIErrorResponse
//	@Router			/api/v1/auth/revoke [post]
func (m *Master) HandleRevokeAccess(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")

	if userID == "" {
		helpers.WriteAPIError(w, "user_id required", helpers.ErrCodeInvalidInput, "user_id query parameter is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err := m.Store.MasterClient.Delete(ctx, "/auth/scopes/"+userID)
	if err != nil {
		helpers.WriteAPIError(w, "Failed to revoke access", helpers.ErrCodeInternal, fmt.Sprintf("failed to revoke access: %v", err), http.StatusInternalServerError)
		return
	}

	_, writeErr := w.Write([]byte("OK"))
	if writeErr != nil {
		helpers.WriteAPIError(w, "Failed to write response", helpers.ErrCodeInternal, fmt.Sprintf("failed to write response: %v", writeErr), http.StatusInternalServerError)
		return
	}
}

// HandleGenerateUserToken mints a new JWT for a named user (requires admin bearer auth)
//
//	@Summary		Generate token for user
//	@Description	Mints a new JWT for the given user. Requires an authenticated admin session — no cluster_access.toml needed.
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id	path		string	true	"User ID"
//	@Success		200		{object}	map[string]string
//	@Failure		400		{object}	helpers.APIErrorResponse
//	@Failure		403		{object}	helpers.APIErrorResponse
//	@Failure		500		{object}	helpers.APIErrorResponse
//	@Router			/api/v1/auth/users/{user_id}/token [post]
func (m *Master) HandleGenerateUserToken(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	if userID == "" {
		helpers.WriteAPIError(w, "user_id required", helpers.ErrCodeInvalidInput, "user_id path parameter is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userPerms, err := m.Store.GetUserPermissions(ctx, userID)
	if err != nil {
		helpers.WriteAPIError(w, "Failed to fetch user permissions", helpers.ErrCodeInternal,
			fmt.Sprintf("failed to fetch user permissions: %v. Run 'ctl users create' first.", err), http.StatusForbidden)
		return
	}

	primaryRole := ""
	if len(userPerms.RoleNames) > 0 {
		primaryRole = userPerms.RoleNames[0]
	}

	claims := auth.NewClaims(userID, primaryRole, []string{}, 365*24*time.Hour)
	token, err := m.IDMgr.Mint(claims)
	if err != nil {
		helpers.WriteAPIError(w, "Failed to mint token", helpers.ErrCodeInternal, fmt.Sprintf("failed to mint token: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"token": token}); err != nil {
		m.Logger.Error("Failed to write generate token response", "err", err)
	}
}

// HandleRevokeToken adds a token's JTI to the revocation list
//
//	@Summary		Revoke a JWT token
//	@Description	Adds a token JTI to the etcd revocation list so it is rejected on all future requests
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		models.RevokeTokenRequest	true	"Token revocation request"
//	@Success		200		{string}	string						"OK"
//	@Failure		400		{object}	helpers.APIErrorResponse
//	@Failure		500		{object}	helpers.APIErrorResponse
//	@Router			/api/v1/auth/tokens/revoke [post]
func (m *Master) HandleRevokeToken(w http.ResponseWriter, r *http.Request) {
	var req models.RevokeTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteAPIError(w, "Invalid request body", helpers.ErrCodeInvalidRequest, fmt.Sprintf("failed to decode request body: %v", err), http.StatusBadRequest)
		return
	}
	if req.JTI == "" {
		helpers.WriteAPIError(w, "jti required", helpers.ErrCodeInvalidInput, "jti field is required in request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := m.Store.RevokeToken(ctx, req.JTI); err != nil {
		helpers.WriteAPIError(w, "Failed to revoke token", helpers.ErrCodeInternal, fmt.Sprintf("failed to revoke token: %v", err), http.StatusInternalServerError)
		return
	}

	if _, err := w.Write([]byte("OK")); err != nil {
		m.Logger.Error("Failed to write revoke token response", "err", err)
	}
}

// HandleAuthStatus returns current authentication status
//
//	@Summary		Get authentication status
//	@Description	Returns current user's authentication status and scopes
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	models.AuthStatusResponse
//	@Failure		500	{object}	helpers.APIErrorResponse
//	@Router			/api/v1/auth/status [get]
func (m *Master) HandleAuthStatus(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r)
	if !ok {
		m.Logger.Error("Failed to get claims from JWT", "err", "claims not found in context")
		helpers.WriteAPIError(w, "failed to get claims from jwt", helpers.ErrCodeGetClaims, "failed to get claims from jwt even though this is past middleware and shouldn't happen", http.StatusInternalServerError)
		return
	}
	userID := claims.UserID
	perms, err := m.Store.GetUserPermissions(r.Context(), userID)
	if err != nil {
		m.Logger.Error("Failed to get user permissions", "err", err, "user_id", userID)
		helpers.WriteAPIError(w, "failed to get user permissions", helpers.ErrCodeInternal, fmt.Sprintf("failed to get user permissions: %v", err), http.StatusInternalServerError)
		return
	}
	effectiveScopes, err := m.Store.GetUserScopes(r.Context(), userID)
	if err != nil {
		m.Logger.Error("Failed to get effective user scopes", "err", err, "user_id", userID)
		helpers.WriteAPIError(w, "failed to get user scopes", helpers.ErrCodeInternal, fmt.Sprintf("failed to get user scopes: %v", err), http.StatusInternalServerError)
		return
	}
	res := models.AuthStatusResponse{
		Authenticated: true,
		User:          userID,
		Roles:         perms.RoleNames,
		Permissions:   scopesToInfo(effectiveScopes),
		DenyScopes:    scopesToInfo(perms.DenyScopes),
	}
	if writeResErr := helpers.WriteJSON(w, res); writeResErr != nil {
		m.Logger.Error("Failed to write auth status response", "err", writeResErr)
		helpers.WriteAPIError(w, "Failed to write auth status response", helpers.ErrCodeInternal, fmt.Sprintf("failed to write auth status response: %v", writeResErr), http.StatusInternalServerError)
		return
	}
}
