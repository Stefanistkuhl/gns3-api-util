package auth

import (
	"crypto/ed25519"
	"fmt"
	"time"

	"github.com/0xveya/gns3util/pkg/web/scopes"
	"github.com/golang-jwt/jwt/v5"
)

const (
	RoleAdmin  = "admin"
	RoleWorker = "worker"
	RolePublic = "public"
)

type Claims struct {
	UserID string   `json:"user_id"`
	Role   string   `json:"role"`
	Scopes []string `json:"scopes"`
	jwt.RegisteredClaims
}

func NewClaims(userID, role string, userScopes []string, ttl time.Duration) *Claims {
	now := time.Now()
	return &Claims{
		UserID: userID,
		Role:   role,
		Scopes: userScopes,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    "gns3util-cluster",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        fmt.Sprintf("%s-%d", userID, now.UnixNano()),
		},
	}
}

func (c *Claims) HasScope(action scopes.Action, res scopes.Resource) bool {
	if c.Role == RoleAdmin {
		return true
	}
	target := scopes.New(action, res)
	wildcard := scopes.All(res)

	for _, s := range c.Scopes {
		if s == target || s == wildcard || s == "*:*" {
			return true
		}
	}
	return false
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
			return nil, jwt.ErrSignatureInvalid
		}
		return m.publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
