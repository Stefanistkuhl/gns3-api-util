package auth

import (
	"crypto/ed25519"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/0xveya/gns3util/pkg/web/scopes"
)

func TestNewClaims(t *testing.T) {
	userID := "user123"
	role := RoleWorker
	userScopes := []string{"read:vms", "write:backups"}
	ttl := time.Hour

	claims := NewClaims(userID, role, userScopes, ttl)

	if claims.UserID != userID {
		t.Errorf("UserID = %v, want %v", claims.UserID, userID)
	}
	if claims.Role != role {
		t.Errorf("Role = %v, want %v", claims.Role, role)
	}
	if len(claims.Scopes) != len(userScopes) {
		t.Errorf("Scopes length = %v, want %v", len(claims.Scopes), len(userScopes))
	}
	for i, scope := range claims.Scopes {
		if scope != userScopes[i] {
			t.Errorf("Scopes[%d] = %v, want %v", i, scope, userScopes[i])
		}
	}
	if claims.Subject != userID {
		t.Errorf("Subject = %v, want %v", claims.Subject, userID)
	}
	if claims.Issuer != "gns3util-cluster" {
		t.Errorf("Issuer = %v, want %v", claims.Issuer, "gns3util-cluster")
	}
	if claims.IssuedAt == nil {
		t.Error("IssuedAt should not be nil")
	}
	if claims.ExpiresAt == nil {
		t.Error("ExpiresAt should not be nil")
	}
	if claims.ID == "" {
		t.Error("ID should not be empty")
	}
	if !strings.HasPrefix(claims.ID, userID) {
		t.Errorf("ID should start with userID, got %v", claims.ID)
	}
}

func TestClaimsHasScope(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		scopes   []string
		action   scopes.Action
		resource scopes.Resource
		expected bool
	}{
		{"admin always has scope", RoleAdmin, []string{}, scopes.Read, scopes.VMs, true},
		{"worker with exact scope", RoleWorker, []string{"read:vms"}, scopes.Read, scopes.VMs, true},
		{"worker with wildcard resource", RoleWorker, []string{"*:vms"}, scopes.Read, scopes.VMs, true},
		{"worker with superuser", RoleWorker, []string{"*:*"}, scopes.Read, scopes.VMs, true},
		{"worker without scope", RoleWorker, []string{"write:backups"}, scopes.Read, scopes.VMs, false},
		{"worker with different action", RoleWorker, []string{"write:vms"}, scopes.Read, scopes.VMs, false},
		{"worker with different resource", RoleWorker, []string{"read:backups"}, scopes.Read, scopes.VMs, false},
		{"public without scope", RolePublic, []string{}, scopes.Read, scopes.VMs, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &Claims{
				Role:   tt.role,
				Scopes: tt.scopes,
			}
			got := claims.HasScope(tt.action, tt.resource)
			if got != tt.expected {
				t.Errorf("HasScope(%v, %v) = %v, want %v", tt.action, tt.resource, got, tt.expected)
			}
		})
	}
}

func TestEncodeDecodePrivateKey(t *testing.T) {
	_, privKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	encoded := EncodePrivateKey(privKey)
	if encoded == "" {
		t.Error("EncodePrivateKey() returned empty string")
	}

	decoded, err := DecodePrivateKey(encoded)
	if err != nil {
		t.Fatalf("DecodePrivateKey() error = %v", err)
	}

	if !privKey.Equal(decoded) {
		t.Error("Decoded private key doesn't match original")
	}
}

func TestEncodeDecodePublicKey(t *testing.T) {
	pubKey, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	encoded := EncodePublicKey(pubKey)
	if encoded == "" {
		t.Error("EncodePublicKey() returned empty string")
	}

	decoded, err := DecodePublicKey(encoded)
	if err != nil {
		t.Fatalf("DecodePublicKey() error = %v", err)
	}

	if !pubKey.Equal(decoded) {
		t.Error("Decoded public key doesn't match original")
	}
}

func TestDecodePrivateKeyErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"invalid base64", "invalid-base64!", true},
		{"wrong size", base64.StdEncoding.EncodeToString([]byte{1, 2, 3}), true},
		{"empty string", "", true},
		{"valid but wrong length", base64.StdEncoding.EncodeToString(make([]byte, 100)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodePrivateKey(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodePrivateKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDecodePublicKeyErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"invalid base64", "invalid-base64!", true},
		{"wrong size", base64.StdEncoding.EncodeToString([]byte{1, 2, 3}), true},
		{"empty string", "", true},
		{"valid but wrong length", base64.StdEncoding.EncodeToString(make([]byte, 100)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodePublicKey(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodePublicKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateKeyPair(t *testing.T) {
	pubKey, privKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	if len(pubKey) != ed25519.PublicKeySize {
		t.Errorf("Public key size = %v, want %v", len(pubKey), ed25519.PublicKeySize)
	}
	if len(privKey) != ed25519.PrivateKeySize {
		t.Errorf("Private key size = %v, want %v", len(privKey), ed25519.PrivateKeySize)
	}

	derivedPub, ok := privKey.Public().(ed25519.PublicKey)
	if !ok {
		t.Fatal("Failed to derive public key")
	}
	if !derivedPub.Equal(pubKey) {
		t.Error("Public key derived from private key doesn't match")
	}
}

func TestNewIdentityManager(t *testing.T) {
	_, privKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	mgr, err := NewIdentityManager(privKey)
	if err != nil {
		t.Fatalf("NewIdentityManager() error = %v", err)
	}

	if mgr.privateKey == nil {
		t.Error("privateKey should not be nil")
	}
	if mgr.publicKey == nil {
		t.Error("publicKey should not be nil")
	}

	pubKey := mgr.PublicKey()
	if !pubKey.Equal(mgr.publicKey) {
		t.Error("PublicKey() doesn't return the stored public key")
	}
}

func TestNewIdentityManagerFromPubKey(t *testing.T) {
	pubKey, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	encodedPub := EncodePublicKey(pubKey)
	mgr, err := NewIdentityManagerFromPubKey(encodedPub)
	if err != nil {
		t.Fatalf("NewIdentityManagerFromPubKey() error = %v", err)
	}

	if mgr.publicKey == nil {
		t.Error("publicKey should not be nil")
	}
	if mgr.privateKey != nil {
		t.Error("privateKey should be nil for public key only manager")
	}
}

func TestIdentityManagerMintValidate(t *testing.T) {
	_, privKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	mgr, err := NewIdentityManager(privKey)
	if err != nil {
		t.Fatalf("NewIdentityManager() error = %v", err)
	}

	claims := NewClaims("user123", RoleWorker, []string{"read:vms"}, time.Hour)
	token, err := mgr.Mint(claims)
	if err != nil {
		t.Fatalf("Mint() error = %v", err)
	}

	if token == "" {
		t.Error("Mint() returned empty token")
	}

	validatedClaims, err := mgr.Validate(token)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if validatedClaims.UserID != claims.UserID {
		t.Errorf("Validated UserID = %v, want %v", validatedClaims.UserID, claims.UserID)
	}
	if validatedClaims.Role != claims.Role {
		t.Errorf("Validated Role = %v, want %v", validatedClaims.Role, claims.Role)
	}
}

func TestIdentityManagerValidateErrors(t *testing.T) {
	_, privKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	mgr, err := NewIdentityManager(privKey)
	if err != nil {
		t.Fatalf("NewIdentityManager() error = %v", err)
	}

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{"empty token", "", true},
		{"invalid token", "invalid.jwt.token", true},
		{"malformed token", "not.a.jwt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mgr.Validate(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIdentityManagerMintErrors(t *testing.T) {
	pubKey, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	mgr, err := NewIdentityManagerFromPubKey(EncodePublicKey(pubKey))
	if err != nil {
		t.Fatalf("NewIdentityManagerFromPubKey() error = %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic when minting with public-key-only manager: %v", r)
		}
	}()

	claims := NewClaims("user123", RoleWorker, []string{"read:vms"}, time.Hour)
	_, err = mgr.Mint(claims)

	if err == nil {
		t.Error("Mint() should panic with public key only manager")
	}
}
