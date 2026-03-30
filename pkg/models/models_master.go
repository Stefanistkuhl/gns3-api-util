package models

type CertUploadRequest struct {
	CertPEM string `json:"cert_pem"`
	KeyPEM  string `json:"key_pem"`
}

type JoinClusterRequest struct {
	Name     string   `json:"name"`
	PeerURLs []string `json:"peer_urls"`
	CSRPEM   []byte   `json:"csr_pem"`
}

type JoinClusterResponse struct {
	MemberID uint64 `json:"member_id"`
	Cluster  string `json:"cluster"`
	CertPEM  []byte `json:"cert_pem"`
	CACert   []byte `json:"ca_cert"`
}

type CreateTokenRequest struct {
	UserID string `json:"user_id"`
}

type JoinFilestoreRequest struct {
	CSRPEM  string `json:"csr"`
	ID      string `json:"id"`
	IP      string `json:"ip_address"`
	APIPort uint32 `json:"api_port"`
}

type JoinFilestoreResponse struct {
	NodeCert string `json:"node_cert"`
	CACert   string `json:"ca_cert"`
}

type AuthStatusResponse struct {
	Authenticated bool        `json:"authenticated"`
	User          string      `json:"user"`
	Roles         []string    `json:"roles"`
	Permissions   []ScopeInfo `json:"permissions,omitempty"`
	DenyScopes    []ScopeInfo `json:"deny_scopes,omitempty"`
}

type GetNodesResponse struct {
	Nodes []NodeInfo `json:"nodes"`
}

type NodeInfo struct {
	ID      string   `json:"id"`
	IP      string   `json:"ip"`
	APIPort uint32   `json:"api_port"`
	Type    NodeType `json:"type"`
}

type UserInfo struct {
	UserID      string      `json:"user_id"`
	RoleNames   []string    `json:"role_names"`
	Permissions []ScopeInfo `json:"permissions,omitempty"`
	DenyScopes  []ScopeInfo `json:"deny_scopes,omitempty"`
	UpdatedAt   string      `json:"updated_at,omitempty"`
}

type ListUsersResponse struct {
	Users []UserInfo `json:"users"`
	Count int        `json:"count"`
}

type ScopeInfo struct {
	Action   string `json:"action"`
	Resource string `json:"resource"`
}

type RoleInfo struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Scopes      []ScopeInfo `json:"scopes"`
	CreatedAt   string      `json:"created_at,omitempty"`
	UpdatedAt   string      `json:"updated_at,omitempty"`
}

type ListRolesResponse struct {
	Roles []RoleInfo `json:"roles"`
	Count int        `json:"count"`
}

type CreateRoleRequest struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Scopes      []ScopeInfo `json:"scopes"`
}

type UpdateRoleRequest struct {
	Description string      `json:"description"`
	Scopes      []ScopeInfo `json:"scopes"`
}

type CreateUserRequest struct {
	Name       string      `json:"name"`
	Roles      []string    `json:"roles"`
	DenyScopes []ScopeInfo `json:"deny_scopes,omitempty"`
}

type AssignRoleRequest struct {
	Role string `json:"role"`
}

type RevokeTokenRequest struct {
	JTI string `json:"jti"`
}
