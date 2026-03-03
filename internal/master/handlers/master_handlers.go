package handlers

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/0xveya/gns3util/pkg/web/auth"
	"github.com/0xveya/gns3util/pkg/web/certs"
)

func (m *Master) HandleJoinCluster(w http.ResponseWriter, r *http.Request) {
	expectedToken := os.Getenv("JOIN_TOKEN")
	authHeader := r.Header.Get("Authorization")
	if authHeader != "Bearer "+expectedToken {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req JoinClusterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	caCertPath := filepath.Join(m.TLSDir, "node.crt")
	caKeyPath := filepath.Join(m.TLSDir, "node.key")

	caCertPEM, err := os.ReadFile(caCertPath)
	if err != nil {
		http.Error(w, "Failed to read Master CA cert", http.StatusInternalServerError)
		return
	}
	caKeyPEM, err := os.ReadFile(caKeyPath)
	if err != nil {
		http.Error(w, "Failed to read Master CA key", http.StatusInternalServerError)
		return
	}

	caBlock, _ := pem.Decode(caCertPEM)
	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		http.Error(w, "Failed to parse Master CA cert", http.StatusInternalServerError)
		return
	}

	keyBlock, _ := pem.Decode(caKeyPEM)
	parsedKey, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		http.Error(w, "Failed to parse Master CA key", http.StatusInternalServerError)
		return
	}
	caPrivKey := parsedKey.(ed25519.PrivateKey)

	signedCertPEM, err := certs.SignCSR(req.CSRPEM, caCert, caPrivKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to sign CSR: %v", err), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	memberResp, err := m.Store.MasterClient.MemberAddAsLearner(ctx, req.PeerURLs)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to add member to etcd: %v", err), http.StatusInternalServerError)
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

	resp := JoinClusterResponse{
		MemberID: memberResp.Member.ID,
		Cluster:  cluster,
		CertPEM:  signedCertPEM,
		CACert:   caCertPEM,
	}

	w.Header().Set("Content-Type", "application/json")
	if encodeErr := json.NewEncoder(w).Encode(resp); encodeErr != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", encodeErr), http.StatusInternalServerError)
		return
	}
}

func (m *Master) HandleCreateToken(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	role := r.URL.Query().Get("role")

	if userID == "" {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}
	if role == "" {
		role = "worker"
	}

	scopes := []string{"files:read", "files:write"}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := m.Store.PutUserPermissions(ctx, userID, scopes); err != nil {
		http.Error(w, "Failed to store permissions", http.StatusInternalServerError)
		return
	}

	claims := auth.NewClaims(userID, role, scopes, 24*time.Hour)
	token, err := m.IDMgr.Mint(claims)
	if err != nil {
		http.Error(w, "Failed to mint token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, writeErr := fmt.Fprintf(w, `{"token":"%s"}`, token)
	if writeErr != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", writeErr), http.StatusInternalServerError)
		return
	}
}

func (m *Master) HandleGrantAccess(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	scope := r.URL.Query().Get("scope")

	if userID == "" || scope == "" {
		http.Error(w, "user_id and scope required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := m.Store.PutUserPermissions(ctx, userID, []string{scope}); err != nil {
		http.Error(w, "Failed to grant access", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("OK"))
}

func (m *Master) HandleRevokeAccess(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")

	if userID == "" {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err := m.Store.MasterClient.Delete(ctx, "/auth/scopes/"+userID)
	if err != nil {
		http.Error(w, "Failed to revoke access", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("OK"))
}
