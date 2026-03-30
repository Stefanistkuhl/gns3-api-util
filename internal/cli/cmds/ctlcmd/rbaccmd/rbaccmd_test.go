package rbaccmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/globals"
	"github.com/0xveya/gns3util/pkg/models"
)

type mockClient struct {
	listRolesResp  *models.ListRolesResponse
	listRolesErr   error
	getRoleResp    *models.RoleInfo
	getRoleErr     error
	createRoleResp *models.RoleInfo
	createRoleErr  error
	updateRoleResp *models.RoleInfo
	updateRoleErr  error
	deleteRoleErr  error

	listUsersResp     *models.ListUsersResponse
	listUsersErr      error
	getUserResp       *models.UserInfo
	getUserErr        error
	deleteUserErr     error
	createUserResp    *models.UserInfo
	createUserErr     error
	assignRoleResp    *models.UserInfo
	assignRoleErr     error
	generateTokenResp string
	generateTokenErr  error
	revokeTokenErr    error

	lastCreateRoleReq  models.CreateRoleRequest
	lastUpdateRoleReq  models.UpdateRoleRequest
	lastDeletedRole    string
	lastDeletedUser    string
	lastAssignRoleReq  models.AssignRoleRequest
	lastRevokeTokenReq models.RevokeTokenRequest
}

func (m *mockClient) ListRoles(_ context.Context) (*models.ListRolesResponse, error) {
	return m.listRolesResp, m.listRolesErr
}

func (m *mockClient) GetRole(_ context.Context, roleName string) (*models.RoleInfo, error) {
	return m.getRoleResp, m.getRoleErr
}

func (m *mockClient) CreateRole(_ context.Context, req models.CreateRoleRequest) (*models.RoleInfo, error) {
	m.lastCreateRoleReq = req
	return m.createRoleResp, m.createRoleErr
}

func (m *mockClient) UpdateRole(_ context.Context, roleName string, req models.UpdateRoleRequest) (*models.RoleInfo, error) {
	m.lastUpdateRoleReq = req
	return m.updateRoleResp, m.updateRoleErr
}

func (m *mockClient) DeleteRole(_ context.Context, roleName string) error {
	m.lastDeletedRole = roleName
	return m.deleteRoleErr
}

func (m *mockClient) ListUsers(_ context.Context) (*models.ListUsersResponse, error) {
	return m.listUsersResp, m.listUsersErr
}

func (m *mockClient) GetUser(_ context.Context, userID string) (*models.UserInfo, error) {
	return m.getUserResp, m.getUserErr
}

func (m *mockClient) DeleteUser(_ context.Context, userID string) error {
	m.lastDeletedUser = userID
	return m.deleteUserErr
}

func (m *mockClient) CreateUser(_ context.Context, req models.CreateUserRequest) (*models.UserInfo, error) {
	return m.createUserResp, m.createUserErr
}

func (m *mockClient) AssignRole(_ context.Context, _ string, req models.AssignRoleRequest) (*models.UserInfo, error) {
	m.lastAssignRoleReq = req
	return m.assignRoleResp, m.assignRoleErr
}

func (m *mockClient) GenerateUserToken(_ context.Context, _ string) (string, error) {
	return m.generateTokenResp, m.generateTokenErr
}

func (m *mockClient) RevokeToken(_ context.Context, req models.RevokeTokenRequest) error {
	m.lastRevokeTokenReq = req
	return m.revokeTokenErr
}

func fixedFactory(c rbacClient) clientFactory {
	return func(_ *config.GlobalOptions) (rbacClient, error) {
		return c, nil
	}
}

func errFactory(err error) clientFactory {
	return func(_ *config.GlobalOptions) (rbacClient, error) {
		return nil, err
	}
}

func executeCmd(t *testing.T, factory clientFactory, args []string) (string, error) {
	t.Helper()

	ctx := config.WithGlobalOptions(
		context.Background(),
		&config.GlobalOptions{
			OutputFormat: 0,
		},
	)

	cmd := newRolesCmdWithFactory(factory)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetContext(ctx)
	cmd.SetArgs(args)

	_, err := cmd.ExecuteC()
	return buf.String(), err
}

func executeUsersCmd(t *testing.T, factory clientFactory, args []string) (string, error) {
	t.Helper()

	ctx := config.WithGlobalOptions(
		context.Background(),
		&config.GlobalOptions{OutputFormat: 0},
	)

	cmd := newUsersCmdWithFactory(factory)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetContext(ctx)
	cmd.SetArgs(args)

	_, err := cmd.ExecuteC()
	return buf.String(), err
}

func TestScopesString(t *testing.T) {
	tests := []struct {
		name   string
		scopes []models.ScopeInfo
		want   string
	}{
		{
			name:   "empty",
			scopes: nil,
			want:   "",
		},
		{
			name: "single scope",
			scopes: []models.ScopeInfo{
				{Action: "read", Resource: "files"},
			},
			want: "read:files",
		},
		{
			name: "multiple scopes",
			scopes: []models.ScopeInfo{
				{Action: "read", Resource: "files"},
				{Action: "write", Resource: "backups"},
			},
			want: "read:files, write:backups",
		},
		// superuser / admin aliases
		{
			name:   "superuser action empty resource",
			scopes: []models.ScopeInfo{{Action: "superuser", Resource: ""}},
			want:   "superuser",
		},
		{
			name:   "ACTION_ADMIN RESOURCE_UNSPECIFIED (proto round-trip)",
			scopes: []models.ScopeInfo{{Action: "ACTION_ADMIN", Resource: "RESOURCE_UNSPECIFIED"}},
			want:   "superuser",
		},
		{
			name:   "admin bare (no resource)",
			scopes: []models.ScopeInfo{{Action: "admin", Resource: ""}},
			want:   "superuser",
		},
		{
			name: "superuser mixed with normal scope",
			scopes: []models.ScopeInfo{
				{Action: "ACTION_ADMIN", Resource: "RESOURCE_UNSPECIFIED"},
				{Action: "read", Resource: "files"},
			},
			want: "superuser, read:files",
		},
		{
			// admin:system is a valid non-global admin scope — must NOT become "superuser"
			name:   "admin with specific resource is not superuser",
			scopes: []models.ScopeInfo{{Action: "admin", Resource: "system"}},
			want:   "admin:system",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scopesString(tt.scopes)
			if got != tt.want {
				t.Errorf("scopesString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseScopeStr(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    []models.ScopeInfo
		wantErr bool
	}{
		{
			name: "empty string",
			raw:  "",
			want: nil,
		},
		{
			name: "single scope",
			raw:  "read:files",
			want: []models.ScopeInfo{{Action: "read", Resource: "files"}},
		},
		{
			name: "multiple scopes",
			raw:  "read:files,write:backups",
			want: []models.ScopeInfo{
				{Action: "read", Resource: "files"},
				{Action: "write", Resource: "backups"},
			},
		},
		{
			name: "scopes with spaces",
			raw:  " read:files , write:backups ",
			want: []models.ScopeInfo{
				{Action: "read", Resource: "files"},
				{Action: "write", Resource: "backups"},
			},
		},
		{
			name:    "missing colon",
			raw:     "readfiles",
			wantErr: true,
		},
		{
			name:    "missing colon in one of many",
			raw:     "read:files,writebackups",
			wantErr: true,
		},
		{
			name: "empty segments ignored",
			raw:  "read:files,,write:backups",
			want: []models.ScopeInfo{
				{Action: "read", Resource: "files"},
				{Action: "write", Resource: "backups"},
			},
		},
		// superuser / admin aliases
		{
			name: "bare superuser",
			raw:  "superuser",
			want: []models.ScopeInfo{{Action: "superuser", Resource: ""}},
		},
		{
			name: "bare SUPERUSER uppercase",
			raw:  "SUPERUSER",
			want: []models.ScopeInfo{{Action: "superuser", Resource: ""}},
		},
		{
			name: "bare admin",
			raw:  "admin",
			want: []models.ScopeInfo{{Action: "superuser", Resource: ""}},
		},
		{
			name: "superuser mixed with normal",
			raw:  "superuser,read:files",
			want: []models.ScopeInfo{
				{Action: "superuser", Resource: ""},
				{Action: "read", Resource: "files"},
			},
		},
		{
			// admin:system still parses normally — only bare "admin" is the alias
			name: "admin with resource is not alias",
			raw:  "admin:system",
			want: []models.ScopeInfo{{Action: "admin", Resource: "system"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseScopeStr(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseScopeStr(%q) error = %v, wantErr %v", tt.raw, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Fatalf("parseScopeStr(%q) = %v, want %v", tt.raw, got, tt.want)
				}
				for i, scope := range got {
					if scope != tt.want[i] {
						t.Errorf("scope[%d] = %+v, want %+v", i, scope, tt.want[i])
					}
				}
			}
		})
	}
}

func TestRoleRowHeaders(t *testing.T) {
	r := &roleRow{}
	headers := r.GetHeaders()
	expected := []string{"NAME", "DESCRIPTION", "SCOPES", "UPDATED AT"}
	if len(headers) != len(expected) {
		t.Fatalf("GetHeaders() len = %d, want %d", len(headers), len(expected))
	}
	for i, h := range headers {
		if h != expected[i] {
			t.Errorf("header[%d] = %q, want %q", i, h, expected[i])
		}
	}
}

func TestRoleRowGetRow(t *testing.T) {
	r := &roleRow{
		Name:        "admin",
		Description: "All access",
		Scopes:      "read:vms, write:vms",
		UpdatedAt:   "2024-01-01",
	}
	row := r.GetRow()
	expected := []string{"admin", "All access", "read:vms, write:vms", "2024-01-01"}
	for i, v := range row {
		if v != expected[i] {
			t.Errorf("GetRow()[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

func TestUserRowHeaders(t *testing.T) {
	u := &userRow{}
	headers := u.GetHeaders()
	expected := []string{"USER ID", "ROLES", "PERMISSIONS", "UPDATED AT"}
	if len(headers) != len(expected) {
		t.Fatalf("GetHeaders() len = %d, want %d", len(headers), len(expected))
	}
	for i, h := range headers {
		if h != expected[i] {
			t.Errorf("header[%d] = %q, want %q", i, h, expected[i])
		}
	}
}

func TestUserRowGetRow(t *testing.T) {
	u := &userRow{
		UserID: "alice",
		Roles:  "admin, worker",
		Permissions: []models.ScopeInfo{
			{Action: "admin", Resource: "*"},
			{Action: "read", Resource: "jobs"},
		},
		UpdatedAt: "2024-01-01",
	}
	row := u.GetRow()
	expected := []string{"alice", "admin, worker", "admin:*, read:jobs", "2024-01-01"}
	for i, v := range row {
		if v != expected[i] {
			t.Errorf("GetRow()[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

func TestNewRolesCmd_Structure(t *testing.T) {
	cmd := NewRolesCmd()
	if cmd == nil {
		t.Fatal("NewRolesCmd() returned nil")
		return
	}
	if cmd.Use != "roles" {
		t.Errorf("Use = %q, want %q", cmd.Use, "roles")
	}

	subNames := map[string]bool{}
	for _, sub := range cmd.Commands() {
		subNames[sub.Name()] = true
	}
	for _, expected := range []string{"list", "get", "create", "update", "delete"} {
		if !subNames[expected] {
			t.Errorf("expected sub-command %q not found", expected)
		}
	}
}

func TestNewUsersCmd_Structure(t *testing.T) {
	cmd := NewUsersCmd()
	if cmd == nil {
		t.Fatal("NewUsersCmd() returned nil")
		return
	}
	if cmd.Use != "users" {
		t.Errorf("Use = %q, want %q", cmd.Use, "users")
	}

	subNames := map[string]bool{}
	for _, sub := range cmd.Commands() {
		subNames[sub.Name()] = true
	}
	for _, expected := range []string{"list", "get", "delete", "assign-role", "gen-token", "revoke-token"} {
		if !subNames[expected] {
			t.Errorf("expected sub-command %q not found", expected)
		}
	}
}

func TestListRoles_Empty(t *testing.T) {
	mc := &mockClient{
		listRolesResp: &models.ListRolesResponse{Roles: nil},
	}
	out, err := executeCmd(t, fixedFactory(mc), []string{"list"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "No roles found.") {
		t.Errorf("expected 'No roles found.' in output, got: %q", out)
	}
}

func TestListRoles_WithData(t *testing.T) {
	mc := &mockClient{
		listRolesResp: &models.ListRolesResponse{
			Roles: []models.RoleInfo{
				{
					Name:        "viewer",
					Description: "Read-only",
					Scopes:      []models.ScopeInfo{{Action: "read", Resource: "files"}},
					UpdatedAt:   "2024-01-01",
				},
			},
		},
	}

	out, err := executeCmd(t, fixedFactory(mc), []string{"list"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "viewer") {
		t.Errorf("expected role name 'viewer' in output, got: %q", out)
	}
}

func TestListRoles_APIError(t *testing.T) {
	mc := &mockClient{listRolesErr: errors.New("connection refused")}
	_, err := executeCmd(t, fixedFactory(mc), []string{"list"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("error = %v, want to contain 'connection refused'", err)
	}
}

func TestListRoles_FactoryError(t *testing.T) {
	_, err := executeCmd(t, errFactory(errors.New("no cluster")), []string{"list"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetRole_Success(t *testing.T) {
	mc := &mockClient{
		getRoleResp: &models.RoleInfo{
			Name:        "admin",
			Description: "Full access",
			Scopes:      []models.ScopeInfo{{Action: "admin", Resource: "system"}},
			UpdatedAt:   "2024-01-01",
		},
	}

	out, err := executeCmd(t, fixedFactory(mc), []string{"get", "admin"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "admin") {
		t.Errorf("expected 'admin' in output, got: %q", out)
	}
}

func TestGetRole_NotFound(t *testing.T) {
	mc := &mockClient{getRoleErr: errors.New("role not found")}
	_, err := executeCmd(t, fixedFactory(mc), []string{"get", "nonexistent"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateRole_Success(t *testing.T) {
	mc := &mockClient{
		createRoleResp: &models.RoleInfo{
			Name:        "viewer",
			Description: "Read-only access",
			Scopes:      []models.ScopeInfo{{Action: "read", Resource: "files"}},
			UpdatedAt:   "2024-01-01",
		},
	}

	out, err := executeCmd(t, fixedFactory(mc), []string{
		"create", "viewer",
		"--description", "Read-only access",
		"--scopes", "read:files",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "viewer") {
		t.Errorf("expected 'viewer' in output, got: %q", out)
	}
	if mc.lastCreateRoleReq.Name != "viewer" {
		t.Errorf("CreateRole called with name %q, want 'viewer'", mc.lastCreateRoleReq.Name)
	}
	if mc.lastCreateRoleReq.Description != "Read-only access" {
		t.Errorf("CreateRole called with description %q, want 'Read-only access'", mc.lastCreateRoleReq.Description)
	}
	if len(mc.lastCreateRoleReq.Scopes) != 1 {
		t.Fatalf("CreateRole called with %d scopes, want 1", len(mc.lastCreateRoleReq.Scopes))
	}
	if mc.lastCreateRoleReq.Scopes[0].Action != "read" || mc.lastCreateRoleReq.Scopes[0].Resource != "files" {
		t.Errorf("unexpected scope: %+v", mc.lastCreateRoleReq.Scopes[0])
	}
}

func TestCreateRole_APIError(t *testing.T) {
	mc := &mockClient{createRoleErr: errors.New("already exists")}
	_, err := executeCmd(t, fixedFactory(mc), []string{"create", "viewer"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateRole_InvalidScopes(t *testing.T) {
	mc := &mockClient{}
	_, err := executeCmd(t, fixedFactory(mc), []string{
		"create", "viewer",
		"--scopes", "invalidscope",
	})
	if err == nil {
		t.Fatal("expected error for invalid scope, got nil")
	}
}

func TestUpdateRole_Success(t *testing.T) {
	mc := &mockClient{
		updateRoleResp: &models.RoleInfo{
			Name:        "viewer",
			Description: "Updated description",
			Scopes:      []models.ScopeInfo{{Action: "read", Resource: "files"}},
			UpdatedAt:   "2024-06-01",
		},
	}

	out, err := executeCmd(t, fixedFactory(mc), []string{
		"update", "viewer",
		"--description", "Updated description",
		"--scopes", "read:files",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "viewer") {
		t.Errorf("expected 'viewer' in output, got: %q", out)
	}
}

func TestUpdateRole_APIError(t *testing.T) {
	mc := &mockClient{updateRoleErr: errors.New("not found")}
	_, err := executeCmd(t, fixedFactory(mc), []string{"update", "viewer"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeleteRole_Force(t *testing.T) {
	mc := &mockClient{}

	out, err := executeCmd(t, fixedFactory(mc), []string{"delete", "viewer", "--force"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mc.lastDeletedRole != "viewer" {
		t.Errorf("DeleteRole called with %q, want 'viewer'", mc.lastDeletedRole)
	}
	if !strings.Contains(out, `"viewer" deleted`) {
		t.Errorf("expected deletion message in output, got: %q", out)
	}
}

func TestDeleteRole_APIError(t *testing.T) {
	mc := &mockClient{deleteRoleErr: errors.New("not found")}
	_, err := executeCmd(t, fixedFactory(mc), []string{"delete", "viewer", "--force"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListUsers_Empty(t *testing.T) {
	mc := &mockClient{
		listUsersResp: &models.ListUsersResponse{Users: nil},
	}
	out, err := executeUsersCmd(t, fixedFactory(mc), []string{"list"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "No users found.") {
		t.Errorf("expected 'No users found.' in output, got: %q", out)
	}
}

func TestListUsers_WithData(t *testing.T) {
	mc := &mockClient{
		listUsersResp: &models.ListUsersResponse{
			Users: []models.UserInfo{
				{UserID: "alice", RoleNames: []string{"admin"}, UpdatedAt: "2024-01-01"},
				{UserID: "bob", RoleNames: []string{"worker"}, UpdatedAt: "2024-01-02"},
			},
		},
	}

	out, err := executeUsersCmd(t, fixedFactory(mc), []string{"list"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "alice") {
		t.Errorf("expected 'alice' in output, got: %q", out)
	}
	if !strings.Contains(out, "bob") {
		t.Errorf("expected 'bob' in output, got: %q", out)
	}
}

func TestListUsers_APIError(t *testing.T) {
	mc := &mockClient{listUsersErr: errors.New("etcd unavailable")}
	_, err := executeUsersCmd(t, fixedFactory(mc), []string{"list"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetUser_Success(t *testing.T) {
	mc := &mockClient{
		getUserResp: &models.UserInfo{
			UserID:    "alice",
			RoleNames: []string{"admin"},
			UpdatedAt: "2024-01-01",
		},
	}

	out, err := executeUsersCmd(t, fixedFactory(mc), []string{"get", "alice"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "alice") {
		t.Errorf("expected 'alice' in output, got: %q", out)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	mc := &mockClient{getUserErr: errors.New("user not found")}
	_, err := executeUsersCmd(t, fixedFactory(mc), []string{"get", "ghost"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeleteUser_Force(t *testing.T) {
	mc := &mockClient{}

	out, err := executeUsersCmd(t, fixedFactory(mc), []string{"delete", "alice", "--force"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mc.lastDeletedUser != "alice" {
		t.Errorf("DeleteUser called with %q, want 'alice'", mc.lastDeletedUser)
	}
	if !strings.Contains(out, `"alice" deleted`) {
		t.Errorf("expected deletion message in output, got: %q", out)
	}
}

func TestDeleteUser_APIError(t *testing.T) {
	mc := &mockClient{deleteUserErr: errors.New("not found")}
	_, err := executeUsersCmd(t, fixedFactory(mc), []string{"delete", "alice", "--force"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListRoles_JSONOutput(t *testing.T) {
	mc := &mockClient{
		listRolesResp: &models.ListRolesResponse{
			Roles: []models.RoleInfo{
				{
					Name:        "viewer",
					Description: "Read-only",
					Scopes:      []models.ScopeInfo{{Action: "read", Resource: "files"}},
					UpdatedAt:   "2024-01-01",
				},
			},
		},
	}

	ctx := config.WithGlobalOptions(
		context.Background(),
		&config.GlobalOptions{OutputFormat: globals.OutputCollapsed},
	)

	cmd := newRolesCmdWithFactory(fixedFactory(mc))
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"list"})
	if _, err := cmd.ExecuteC(); err != nil {
		t.Fatalf("ExecuteC() error: %v", err)
	}

	var out []any
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &out); err != nil {
		t.Errorf("output is not valid JSON: %v\noutput: %q", err, buf.String())
	}
}

func TestListUsers_JSONOutput(t *testing.T) {
	mc := &mockClient{
		listUsersResp: &models.ListUsersResponse{
			Users: []models.UserInfo{
				{UserID: "alice", RoleNames: []string{"admin"}, UpdatedAt: "2024-01-01"},
			},
		},
	}

	ctx := config.WithGlobalOptions(
		context.Background(),
		&config.GlobalOptions{OutputFormat: globals.OutputCollapsed},
	)

	cmd := newUsersCmdWithFactory(fixedFactory(mc))
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"list"})
	if _, err := cmd.ExecuteC(); err != nil {
		t.Fatalf("ExecuteC() error: %v", err)
	}

	var out []any
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &out); err != nil {
		t.Errorf("output is not valid JSON: %v\noutput: %q", err, buf.String())
	}
}

func TestCreateRoleCmd_Flags(t *testing.T) {
	cmd := newRolesCmdWithFactory(newMasterClient)

	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil || createCmd.Use == "roles" {
		t.Fatal("could not find 'create' sub-command")
	}

	if f := createCmd.Flags().Lookup("description"); f == nil {
		t.Error("flag --description not registered on create")
	}
	if f := createCmd.Flags().Lookup("scopes"); f == nil {
		t.Error("flag --scopes not registered on create")
	}
}

func TestUpdateRoleCmd_Flags(t *testing.T) {
	cmd := newRolesCmdWithFactory(newMasterClient)

	updateCmd, _, err := cmd.Find([]string{"update"})
	if err != nil || updateCmd.Use == "roles" {
		t.Fatal("could not find 'update' sub-command")
	}

	if f := updateCmd.Flags().Lookup("description"); f == nil {
		t.Error("flag --description not registered on update")
	}
	if f := updateCmd.Flags().Lookup("scopes"); f == nil {
		t.Error("flag --scopes not registered on update")
	}
}

func TestDeleteRoleCmd_Flags(t *testing.T) {
	cmd := newRolesCmdWithFactory(newMasterClient)

	deleteCmd, _, err := cmd.Find([]string{"delete"})
	if err != nil || deleteCmd.Use == "roles" {
		t.Fatal("could not find 'delete' sub-command")
	}

	if f := deleteCmd.Flags().Lookup("force"); f == nil {
		t.Error("flag --force not registered on delete")
	}
}

func TestDeleteUserCmd_Flags(t *testing.T) {
	cmd := newUsersCmdWithFactory(newMasterClient)

	deleteCmd, _, err := cmd.Find([]string{"delete"})
	if err != nil || deleteCmd.Use == "users" {
		t.Fatal("could not find 'delete' sub-command")
	}

	if f := deleteCmd.Flags().Lookup("force"); f == nil {
		t.Error("flag --force not registered on delete")
	}
}

func TestListUsers_MultipleRolesDisplay(t *testing.T) {
	mc := &mockClient{
		listUsersResp: &models.ListUsersResponse{
			Users: []models.UserInfo{
				{
					UserID:    "alice",
					RoleNames: []string{"admin", "worker"},
					UpdatedAt: "2024-01-01",
				},
			},
		},
	}

	out, err := executeUsersCmd(t, fixedFactory(mc), []string{"list"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "admin") || !strings.Contains(out, "worker") {
		t.Errorf("expected both roles in output, got: %q", out)
	}
}

func TestCreateRole_MultipleScopes(t *testing.T) {
	mc := &mockClient{
		createRoleResp: &models.RoleInfo{
			Name: "power-user",
			Scopes: []models.ScopeInfo{
				{Action: "read", Resource: "files"},
				{Action: "write", Resource: "backups"},
				{Action: "delete", Resource: "configs"},
			},
		},
	}

	_, err := executeCmd(t, fixedFactory(mc), []string{
		"create", "power-user",
		"--scopes", "read:files,write:backups,delete:configs",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mc.lastCreateRoleReq.Scopes) != 3 {
		t.Errorf("expected 3 scopes, got %d", len(mc.lastCreateRoleReq.Scopes))
	}
}

func TestRolesCmd_Help(t *testing.T) {
	cmd := NewRolesCmd()
	if err := cmd.Help(); err != nil {
		t.Errorf("Help() returned error: %v", err)
	}
}

func TestUsersCmd_Help(t *testing.T) {
	cmd := NewUsersCmd()
	if err := cmd.Help(); err != nil {
		t.Errorf("Help() returned error: %v", err)
	}
}

func TestRolesGet_FactoryError(t *testing.T) {
	_, err := executeCmd(t, errFactory(fmt.Errorf("no cluster configured")), []string{"get", "admin"})
	if err == nil {
		t.Fatal("expected error from factory, got nil")
	}
}

func TestUsersGet_FactoryError(t *testing.T) {
	_, err := executeUsersCmd(t, errFactory(fmt.Errorf("no cluster configured")), []string{"get", "alice"})
	if err == nil {
		t.Fatal("expected error from factory, got nil")
	}
}

func TestAssignRole_Success(t *testing.T) {
	mc := &mockClient{
		assignRoleResp: &models.UserInfo{
			UserID:    "alice",
			RoleNames: []string{"admin", "viewer"},
			UpdatedAt: "2024-01-01",
		},
	}

	out, err := executeUsersCmd(t, fixedFactory(mc), []string{"assign-role", "alice", "viewer"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mc.lastAssignRoleReq.Role != "viewer" {
		t.Errorf("AssignRole called with role %q, want 'viewer'", mc.lastAssignRoleReq.Role)
	}
	if !strings.Contains(out, "alice") {
		t.Errorf("expected 'alice' in output, got: %q", out)
	}
}

func TestAssignRole_APIError(t *testing.T) {
	mc := &mockClient{assignRoleErr: errors.New("role not found")}
	_, err := executeUsersCmd(t, fixedFactory(mc), []string{"assign-role", "alice", "nonexistent"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "role not found") {
		t.Errorf("error = %v, want to contain 'role not found'", err)
	}
}

func TestAssignRole_FactoryError(t *testing.T) {
	_, err := executeUsersCmd(t, errFactory(fmt.Errorf("no cluster")), []string{"assign-role", "alice", "viewer"})
	if err == nil {
		t.Fatal("expected factory error, got nil")
	}
}

// testJWT builds a minimal but structurally valid JWT string whose payload
// contains only the "jti" claim. The signature segment is a placeholder —
// extractJTI does not verify signatures.
func testJWT(jti string) string {
	header := "eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9" // {"alg":"EdDSA","typ":"JWT"}
	payload := base64.RawURLEncoding.EncodeToString(
		[]byte(`{"jti":"` + jti + `"}`),
	)
	return header + "." + payload + ".fakesignature"
}

func TestGenToken_Success(t *testing.T) {
	token := testJWT("alice-gen-token")
	mc := &mockClient{generateTokenResp: token}

	out, err := executeUsersCmd(t, fixedFactory(mc), []string{"gen-token", "alice"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, token) {
		t.Errorf("expected token in output, got: %q", out)
	}
	if !strings.Contains(out, "alice") {
		t.Errorf("expected user id in output, got: %q", out)
	}
}

func TestGenToken_APIError(t *testing.T) {
	mc := &mockClient{generateTokenErr: errors.New("user not found")}
	_, err := executeUsersCmd(t, fixedFactory(mc), []string{"gen-token", "ghost"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "user not found") {
		t.Errorf("error = %v, want to contain 'user not found'", err)
	}
}

func TestGenToken_FactoryError(t *testing.T) {
	_, err := executeUsersCmd(t, errFactory(fmt.Errorf("no cluster")), []string{"gen-token", "alice"})
	if err == nil {
		t.Fatal("expected factory error, got nil")
	}
}

func TestRevokeToken_Success(t *testing.T) {
	mc := &mockClient{}
	token := testJWT("alice-1712345678000000000")

	out, err := executeUsersCmd(t, fixedFactory(mc), []string{"revoke-token", token})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mc.lastRevokeTokenReq.JTI != "alice-1712345678000000000" {
		t.Errorf("RevokeToken called with JTI %q, want 'alice-1712345678000000000'", mc.lastRevokeTokenReq.JTI)
	}
	if !strings.Contains(out, "revoked") {
		t.Errorf("expected 'revoked' in output, got: %q", out)
	}
}

func TestRevokeToken_APIError(t *testing.T) {
	mc := &mockClient{revokeTokenErr: errors.New("etcd write failed")}
	_, err := executeUsersCmd(t, fixedFactory(mc), []string{"revoke-token", testJWT("alice-123")})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "etcd write failed") {
		t.Errorf("error = %v, want to contain 'etcd write failed'", err)
	}
}

func TestRevokeToken_FactoryError(t *testing.T) {
	_, err := executeUsersCmd(t, errFactory(fmt.Errorf("no cluster")), []string{"revoke-token", testJWT("alice-123")})
	if err == nil {
		t.Fatal("expected factory error, got nil")
	}
}

func TestRevokeToken_InvalidToken(t *testing.T) {
	mc := &mockClient{}
	_, err := executeUsersCmd(t, fixedFactory(mc), []string{"revoke-token", "notavalidjwt"})
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
	if !strings.Contains(err.Error(), "invalid token") {
		t.Errorf("error = %v, want to contain 'invalid token'", err)
	}
}

func TestExtractJTI(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantJTI string
		wantErr string
	}{
		{
			name:    "generated jwt with jti",
			token:   testJWT("erm-1775241419739174946"),
			wantJTI: "erm-1775241419739174946",
		},
		{
			name:    "minimal test jwt",
			token:   testJWT("bob-9999"),
			wantJTI: "bob-9999",
		},
		{
			name:    "not enough parts",
			token:   "onlyone",
			wantErr: "expected 3 dot-separated parts",
		},
		{
			name:    "bad base64 payload",
			token:   "header.!!!.sig",
			wantErr: "failed to base64-decode payload",
		},
		{
			name:    "missing jti claim",
			token:   "h." + base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"alice"}`)) + ".s",
			wantErr: "token does not contain a jti claim",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractJTI(tt.token)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantJTI {
				t.Errorf("extractJTI() = %q, want %q", got, tt.wantJTI)
			}
		})
	}
}
