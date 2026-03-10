package handlers

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/state"
	"github.com/0xveya/gns3util/pkg/web/auth"
	"github.com/0xveya/gns3util/pkg/web/certs"
	"github.com/0xveya/gns3util/pkg/web/helpers"
	"github.com/0xveya/gns3util/pkg/web/middleware"
)

type Master struct {
	IDMgr  *auth.IdentityManager
	Store  *state.StateManager
	TLSDir string
	Logger *slog.Logger
}

func (m *Master) HandleJoinCluster(w http.ResponseWriter, r *http.Request) {
	expectedToken := os.Getenv("JOIN_TOKEN")
	authHeader := r.Header.Get("Authorization")
	if authHeader != "Bearer "+expectedToken {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.JoinClusterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	cleanPath := filepath.Clean(m.TLSDir)
	caCertPath := filepath.Join(cleanPath, "node.crt")
	caKeyPath := filepath.Join(cleanPath, "node.key")

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
	caPrivKey, ok := parsedKey.(ed25519.PrivateKey)
	if !ok {
		http.Error(w, "Failed to parse Master CA key", http.StatusInternalServerError)
		return
	}

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

	resp := models.JoinClusterResponse{
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
	var req models.CreateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" {
		http.Error(w, "user_id required\n", http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		req.Role = "worker"
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	scopesStr, err := m.Store.GetUserScopes(ctx, req.UserID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch user permissions: %v. Run 'ctl create user' first.\n", err), http.StatusForbidden)
		return
	}

	var scopes []string
	if scopesStr != "" {
		scopes = strings.Split(scopesStr, ",")
	}

	claims := auth.NewClaims(req.UserID, req.Role, scopes, 365*24*time.Hour)
	token, err := m.IDMgr.Mint(claims)
	if err != nil {
		http.Error(w, "Failed to mint token\n", http.StatusInternalServerError)
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

func (m *Master) HandleGrantAccess(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	scope := r.URL.Query().Get("scope")

	if userID == "" || scope == "" {
		http.Error(w, "user_id and scope required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	existingScopesStr, err := m.Store.GetUserScopes(ctx, userID)
	if err != nil && !strings.Contains(err.Error(), "user not found") {
		http.Error(w, fmt.Sprintf("Failed to fetch existing permissions: %v", err), http.StatusInternalServerError)
		return
	}

	var scopes []string
	if existingScopesStr != "" {
		scopes = strings.Split(existingScopesStr, ",")
	}

	hasScope := slices.Contains(scopes, scope)

	if !hasScope {
		scopes = append(scopes, scope)
	}

	if err := m.Store.PutUserPermissions(ctx, userID, scopes); err != nil {
		http.Error(w, "Failed to grant access", http.StatusInternalServerError)
		return
	}

	_, writeErr := w.Write([]byte("OK"))
	if writeErr != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", writeErr), http.StatusInternalServerError)
		return
	}
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

	_, writeErr := w.Write([]byte("OK"))
	if writeErr != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", writeErr), http.StatusInternalServerError)
		return
	}
}

func (m *Master) HandleJoinFilestore(w http.ResponseWriter, r *http.Request) {
	expectedToken := os.Getenv("JOIN_TOKEN")
	authHeader := r.Header.Get("Authorization")
	if authHeader != "Bearer "+expectedToken {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.JoinFilestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	cleanPath := filepath.Clean(m.TLSDir)
	caCertPath := filepath.Join(cleanPath, "node.crt")
	caKeyPath := filepath.Join(cleanPath, "node.key")

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
	caPrivKey, ok := parsedKey.(ed25519.PrivateKey)
	if !ok {
		http.Error(w, "Failed to parse Master CA key", http.StatusInternalServerError)
		return
	}

	signedCertPEM, err := certs.SignCSR([]byte(req.CSRPEM), caCert, caPrivKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to sign CSR: %v", err), http.StatusInternalServerError)
		return
	}

	resp := models.JoinFilestoreResponse{
		NodeCert: string(signedCertPEM),
		CACert:   string(caCertPEM),
	}

	w.Header().Set("Content-Type", "application/json")
	if encodeErr := json.NewEncoder(w).Encode(resp); encodeErr != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", encodeErr), http.StatusInternalServerError)
		return
	}
}

func (m *Master) HandleAuthStatus(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r)
	if !ok {
		m.Logger.Error("Failed to get claims from JWT", "err", "claims not found in context")
		helpers.WriteAPIError(w, "failed to get claims from jwt", helpers.ErrCodeGetClaims, "failed to get claims from jwt even though this is past middleware and shouldn't happen", http.StatusInternalServerError)
		return
	}
	userID := claims.UserID
	scopes, err := m.Store.GetUserScopes(r.Context(), userID)
	if err != nil {
		m.Logger.Error("Failed to get user scopes", "err", err, "user_id", userID)
		helpers.WriteAPIError(w, "failed to get user scopes", helpers.ErrCodeGetClaims, fmt.Sprintf("failed to get user scopes: %v", err), http.StatusInternalServerError)
	}
	res := models.AuthStatusResponse{
		Authenticated: true,
		User:          userID,
		Scopes:        scopes,
	}
	if writeResErr := helpers.WriteJSON(w, res); writeResErr != nil {
		m.Logger.Error("Failed to write auth status response", "err", writeResErr)
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}
