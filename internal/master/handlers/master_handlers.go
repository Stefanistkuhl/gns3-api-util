package handlers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/0xveya/gns3util/pkg/web/auth"
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
	}

	w.Header().Set("Content-Type", "application/json")
	encodeErr := json.NewEncoder(w).Encode(resp)
	if encodeErr != nil {
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

func (m *Master) HandleUploadCert(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	rawContentType := r.Header.Get("Content-Type")
	contentType := strings.Split(rawContentType, ";")[0]
	var certPEM, keyPEM []byte

	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "application/json":
		var req CertUploadRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}
		certPEM = []byte(req.CertPEM)
		keyPEM = []byte(req.KeyPEM)

	case "multipart/form-data":
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse multipart: %v", err), http.StatusBadRequest)
			return
		}

		certFile, _, err := r.FormFile("cert")
		if err != nil {
			http.Error(w, "cert file required", http.StatusBadRequest)
			return
		}
		defer certFile.Close()
		certPEM, err = io.ReadAll(certFile)
		if err != nil {
			http.Error(w, "Failed to read cert", http.StatusBadRequest)
			return
		}

		keyFile, _, err := r.FormFile("key")
		if err != nil {
			http.Error(w, "key file required", http.StatusBadRequest)
			return
		}
		defer keyFile.Close()
		keyPEM, err = io.ReadAll(keyFile)
		if err != nil {
			http.Error(w, "Failed to read key", http.StatusBadRequest)
			return
		}

	default:
		http.Error(w, "Content-Type must be application/json or multipart/form-data", http.StatusBadRequest)
		return
	}

	if _, err := tls.X509KeyPair(certPEM, keyPEM); err != nil {
		http.Error(w, fmt.Sprintf("Invalid cert/key pair: %v", err), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := m.Store.PutCertificate(ctx, certPEM, keyPEM); err != nil {
		http.Error(w, "Failed to store certificate", http.StatusInternalServerError)
		return
	}

	log.Printf("Certificate stored in etcd")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func (m *Master) HandleGetCert(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	certPEM, _, err := m.Store.GetCertificate(ctx)
	if err != nil {
		http.Error(w, "Certificate not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Write(certPEM)
}
