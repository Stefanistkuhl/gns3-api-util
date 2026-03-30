package handlers

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/state/pb"
	"github.com/0xveya/gns3util/pkg/web/certs"
	"github.com/0xveya/gns3util/pkg/web/helpers"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GetNodes returns current authentication status
//
//	@Summary		Return nodes in cluster
//	@Description	Returns all nodes in the cluster with their status and basic info
//	@Tags			discovery
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	models.GetNodesResponse
//	@Failure		500	{object}	helpers.APIErrorResponse
//	@Router			/api/v1/auth/status [get]
func (m *Master) GetNodes(w http.ResponseWriter, r *http.Request) {
	nodes, getNodeErr := m.Store.GetNodes(r.Context())
	if getNodeErr != nil {
		m.Logger.Error("Failed to get nodes from state manager", "error", getNodeErr)
		helpers.WriteAPIError(w, "Failed to get nodes", helpers.ErrCodeInternal, fmt.Sprintf("failed to get nodes: %v", getNodeErr), http.StatusInternalServerError)
		return
	}

	var nodeInfos []models.NodeInfo
	for _, node := range nodes {
		nodeInfos = append(nodeInfos, models.NodeInfo{
			ID:      node.Id,
			IP:      node.IpAddress,
			APIPort: uint32(node.ApiPort),
			Type:    models.NodeType(node.Type),
		})
	}

	res := models.GetNodesResponse{
		Nodes: nodeInfos,
	}
	if writeResErr := helpers.WriteJSON(w, res); writeResErr != nil {
		m.Logger.Error("Failed to write get nodes response", "err", writeResErr)
		helpers.WriteAPIError(w, "Failed to write get nodes response", helpers.ErrCodeInternal, fmt.Sprintf("failed to write get nodes response: %v", writeResErr), http.StatusInternalServerError)
		return
	}
}

// HandleJoinFilestore handles filestore node join requests
//
//	@Summary		Join filestore cluster
//	@Description	Signs CSR for filestore node
//	@Tags			cluster
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string						true	"Bearer JOIN_TOKEN"
//	@Param			request			body		models.JoinFilestoreRequest	true	"Join request with CSR"
//	@Success		200				{object}	models.JoinFilestoreResponse
//	@Failure		400				{object}	helpers.APIErrorResponse
//	@Failure		401				{object}	helpers.APIErrorResponse
//	@Failure		500				{object}	helpers.APIErrorResponse
//	@Router			/api/v1/cluster/join/filestore [post]
func (m *Master) HandleJoinFilestore(w http.ResponseWriter, r *http.Request) {
	expectedToken := os.Getenv("JOIN_TOKEN")
	authHeader := r.Header.Get("Authorization")
	if authHeader != "Bearer "+expectedToken {
		helpers.WriteAPIError(w, "Unauthorized", helpers.ErrCodeUnauthorized, "invalid or missing join token", http.StatusUnauthorized)
		return
	}

	var req models.JoinFilestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteAPIError(w, "Invalid request body", helpers.ErrCodeInvalidRequest, fmt.Sprintf("failed to decode request body: %v", err), http.StatusBadRequest)
		return
	}

	cleanPath := filepath.Clean(m.TLSDir)
	caCertPath := filepath.Join(cleanPath, "node.crt")
	caKeyPath := filepath.Join(cleanPath, "node.key")

	caCertPEM, err := os.ReadFile(caCertPath)
	if err != nil {
		helpers.WriteAPIError(w, "Failed to read Master CA cert", helpers.ErrCodeFileNotFound, fmt.Sprintf("failed to read master ca cert: %v", err), http.StatusInternalServerError)
		return
	}
	caKeyPEM, err := os.ReadFile(caKeyPath)
	if err != nil {
		helpers.WriteAPIError(w, "Failed to read Master CA key", helpers.ErrCodeFileNotFound, fmt.Sprintf("failed to read master ca key: %v", err), http.StatusInternalServerError)
		return
	}

	caBlock, _ := pem.Decode(caCertPEM)
	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		helpers.WriteAPIError(w, "Failed to parse Master CA cert", helpers.ErrCodeInvalidInput, fmt.Sprintf("failed to parse master ca cert: %v", err), http.StatusInternalServerError)
		return
	}

	keyBlock, _ := pem.Decode(caKeyPEM)
	parsedKey, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		helpers.WriteAPIError(w, "Failed to parse Master CA key", helpers.ErrCodeInvalidInput, fmt.Sprintf("failed to parse master ca key: %v", err), http.StatusInternalServerError)
		return
	}
	signedCertPEM, err := certs.SignCSR([]byte(req.CSRPEM), caCert, parsedKey)
	if err != nil {
		helpers.WriteAPIError(w, "Failed to sign CSR", helpers.ErrCodeInternal, fmt.Sprintf("failed to sign csr: %v", err), http.StatusInternalServerError)
		return
	}

	resp := models.JoinFilestoreResponse{
		NodeCert: string(signedCertPEM),
		CACert:   string(caCertPEM),
	}
	nodeIP := req.IP
	if nodeIP == "" {
		nodeIP = strings.Split(r.RemoteAddr, ":")[0]
	}
	newNode := &pb.Node{
		Id:        req.ID,
		IpAddress: nodeIP,
		Status:    pb.NodeStatus_NODE_STATUS_ONLINE,
		Type:      pb.NodeType_NODE_TYPE_STORAGE,
		LastSeen:  timestamppb.Now(),
		ApiPort:   req.APIPort,
		Network:   &pb.NodeNetwork{}, // TODO: add this when networking is improved
	}

	nodeScopes := []string{"node:status:update", "filestore:access"}

	success, err := m.Store.RegisterNodeTxn(r.Context(), newNode, nodeScopes)
	if err != nil {
		helpers.WriteAPIError(w, "Internal Error", helpers.ErrCodeInternal, "etcd txn failed", http.StatusInternalServerError)
		return
	}

	if !success {
		helpers.WriteAPIError(w, "Conflict", helpers.ErrCodeInvalidRequest, "Node ID already in use", http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if encodeErr := json.NewEncoder(w).Encode(resp); encodeErr != nil {
		helpers.WriteAPIError(w, "Failed to encode response", helpers.ErrCodeInternal, fmt.Sprintf("failed to encode response: %v", encodeErr), http.StatusInternalServerError)
		return
	}
}

// HandleJoinCluster handles node join requests
//
//	@Summary		Join cluster as a new node
//	@Description	Signs CSR and adds node to etcd cluster
//	@Tags			cluster
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string						true	"Bearer JOIN_TOKEN"
//	@Param			request			body		models.JoinClusterRequest	true	"Join request with CSR"
//	@Success		200				{object}	models.JoinClusterResponse
//	@Failure		400				{object}	helpers.APIErrorResponse
//	@Failure		401				{object}	helpers.APIErrorResponse
//	@Failure		500				{object}	helpers.APIErrorResponse
//	@Router			/api/v1/cluster/join [post]
func (m *Master) HandleJoinCluster(w http.ResponseWriter, r *http.Request) {
	// change this to use mtls proerly bc this used to be for a learner
	expectedToken := os.Getenv("JOIN_TOKEN")
	authHeader := r.Header.Get("Authorization")
	if authHeader != "Bearer "+expectedToken {
		helpers.WriteAPIError(w, "Unauthorized", helpers.ErrCodeUnauthorized, "invalid or missing join token", http.StatusUnauthorized)
		return
	}

	var req models.JoinClusterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteAPIError(w, "Invalid request body", helpers.ErrCodeInvalidRequest, fmt.Sprintf("failed to decode request body: %v", err), http.StatusBadRequest)
		return
	}

	cleanPath := filepath.Clean(m.TLSDir)
	caCertPath := filepath.Join(cleanPath, "node.crt")
	caKeyPath := filepath.Join(cleanPath, "node.key")

	caCertPEM, err := os.ReadFile(caCertPath)
	if err != nil {
		helpers.WriteAPIError(w, "Failed to read Master CA cert", helpers.ErrCodeFileNotFound, fmt.Sprintf("failed to read master ca cert: %v", err), http.StatusInternalServerError)
		return
	}
	caKeyPEM, err := os.ReadFile(caKeyPath)
	if err != nil {
		helpers.WriteAPIError(w, "Failed to read Master CA key", helpers.ErrCodeFileNotFound, fmt.Sprintf("failed to read master ca key: %v", err), http.StatusInternalServerError)
		return
	}

	caBlock, _ := pem.Decode(caCertPEM)
	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		helpers.WriteAPIError(w, "Failed to parse Master CA cert", helpers.ErrCodeInvalidInput, fmt.Sprintf("failed to parse master ca cert: %v", err), http.StatusInternalServerError)
		return
	}

	keyBlock, _ := pem.Decode(caKeyPEM)
	parsedKey, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		helpers.WriteAPIError(w, "Failed to parse Master CA key", helpers.ErrCodeInvalidInput, fmt.Sprintf("failed to parse master ca key: %v", err), http.StatusInternalServerError)
		return
	}

	signedCertPEM, err := certs.SignCSR(req.CSRPEM, caCert, parsedKey)
	if err != nil {
		helpers.WriteAPIError(w, "Failed to sign CSR", helpers.ErrCodeInternal, fmt.Sprintf("failed to sign csr: %v", err), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	memberResp, err := m.Store.MasterClient.MemberAdd(ctx, req.PeerURLs)
	if err != nil {
		helpers.WriteAPIError(w, "Failed to add member to cluster", helpers.ErrCodeInternal, fmt.Sprintf("failed to add member to cluster: %v", err), http.StatusInternalServerError)
		return
	}

	cluster := ""
	for _, mem := range memberResp.Members {
		if cluster != "" {
			cluster += ","
		}
		name := mem.Name
		if name == "" && mem.ID == memberResp.Member.ID {
			name = req.Name
		}
		cluster += fmt.Sprintf("%s=%s", name, mem.PeerURLs[0])
	}

	resp := models.JoinClusterResponse{
		MemberID: memberResp.Member.ID,
		Cluster:  cluster,
		CertPEM:  signedCertPEM,
		CACert:   caCertPEM,
	}

	w.Header().Set("Content-Type", "application/json")
	if encodeErr := json.NewEncoder(w).Encode(resp); encodeErr != nil {
		helpers.WriteAPIError(w, "Failed to encode response", helpers.ErrCodeInternal, fmt.Sprintf("failed to encode response: %v", encodeErr), http.StatusInternalServerError)
		return
	}
}
