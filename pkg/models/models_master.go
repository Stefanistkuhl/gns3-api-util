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
	Role   string `json:"role"`
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
	Authenticated bool   `json:"authenticated"`
	User          string `json:"user"`
	Scopes        string `json:"scopes"`
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
