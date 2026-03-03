package handlers

import (
	"github.com/0xveya/gns3util/pkg/state"
	"github.com/0xveya/gns3util/pkg/web/auth"
)

type Master struct {
	IDMgr  *auth.IdentityManager
	Store  *state.StateManager
	TLSDir string
}

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
