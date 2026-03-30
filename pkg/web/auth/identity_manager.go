package auth

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	RoleAdmin  = "admin"
	RoleWorker = "worker"
	RolePublic = "public"
)

type Claims struct {
	UserID string `json:"user_id"`
	// Role is the primary role assigned to this user (e.g. "admin", "worker").
	Role string `json:"role"`
	// Scopes is the list of explicit permission scopes (e.g. "read:vms").
	Scopes []string `json:"scopes"`
	jwt.RegisteredClaims
}

func NewClaims(userID, role string, scopes []string, ttl time.Duration) *Claims {
	now := time.Now()
	return &Claims{
		UserID: userID,
		Role:   role,
		Scopes: scopes,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    "gns3util-cluster",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        fmt.Sprintf("%s-%d", userID, now.UnixNano()),
		},
	}
}

type IdentityManager struct {
	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey
}

func NewIdentityManager(privKey ed25519.PrivateKey) (*IdentityManager, error) {
	pub, ok := privKey.Public().(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("invalid key type: expected ed25519.PublicKey, got %T", privKey.Public())
	}
	return &IdentityManager{
		publicKey:  pub,
		privateKey: privKey,
	}, nil
}

func NewIdentityManagerFromPubKey(pubKeyStr string) (*IdentityManager, error) {
	pubKey, err := DecodePublicKey(pubKeyStr)
	if err != nil {
		return nil, err
	}
	return &IdentityManager{
		publicKey:  pubKey,
		privateKey: nil,
	}, nil
}

func (m *IdentityManager) Mint(claims *Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return token.SignedString(m.privateKey)
}

func (m *IdentityManager) Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return m.publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (m *IdentityManager) PublicKey() ed25519.PublicKey {
	return m.publicKey
}

func EncodePrivateKey(key ed25519.PrivateKey) string {
	return base64.StdEncoding.EncodeToString(key)
}

func EncodePublicKey(key ed25519.PublicKey) string {
	return base64.StdEncoding.EncodeToString(key)
}

func DecodePrivateKey(s string) (ed25519.PrivateKey, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}
	if len(data) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: expected %d, got %d", ed25519.PrivateKeySize, len(data))
	}
	priv := make(ed25519.PrivateKey, ed25519.PrivateKeySize)
	copy(priv, data)
	return priv, nil
}

func DecodePublicKey(s string) (ed25519.PublicKey, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}
	if len(data) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: expected %d, got %d", ed25519.PublicKeySize, len(data))
	}
	pub := make(ed25519.PublicKey, ed25519.PublicKeySize)
	copy(pub, data)
	return pub, nil
}

func GenerateKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(rand.Reader)
}
