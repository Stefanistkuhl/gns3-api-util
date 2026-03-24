package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/0xveya/gns3util/pkg/web/auth"
)

func TestGetClaims(t *testing.T) {
	claims := &auth.Claims{
		UserID: "user123",
		Role:   "worker",
		Scopes: []string{"read:vms"},
	}

	ctx := context.WithValue(context.Background(), ClaimsKey, claims)
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/", http.NoBody)
	req = req.WithContext(ctx)

	got, ok := GetClaims(req)
	if !ok {
		t.Error("GetClaims() should return true when claims exist")
	}
	if got != claims {
		t.Errorf("GetClaims() = %v, want %v", got, claims)
	}

	req2 := httptest.NewRequestWithContext(context.Background(), "GET", "/", http.NoBody)
	got2, ok2 := GetClaims(req2)
	if ok2 {
		t.Error("GetClaims() should return false when no claims exist")
	}
	if got2 != nil {
		t.Errorf("GetClaims() = %v, want nil", got2)
	}
}

func TestGetClaimsWrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), ClaimsKey, "not a claim")
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/", http.NoBody)
	req = req.WithContext(ctx)

	got, ok := GetClaims(req)
	if ok {
		t.Error("GetClaims() should return false when wrong type in context")
	}
	if got != nil {
		t.Errorf("GetClaims() = %v, want nil", got)
	}
}

func TestClaimsKey(t *testing.T) {
	if _, ok := any(ClaimsKey).(ctxKey); !ok {
		t.Errorf("ClaimsKey should be of type ctxKey, got %T", ClaimsKey)
	}

	if ClaimsKey == "" {
		t.Error("ClaimsKey should not be empty")
	}

	if string(ClaimsKey) != "user_claims" {
		t.Errorf("ClaimsKey = %v, want %v", ClaimsKey, "user_claims")
	}
}

func TestContextValueOperations(t *testing.T) {
	originalClaims := &auth.Claims{
		UserID: "user456",
		Role:   "admin",
		Scopes: []string{"*:*"},
	}

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/", http.NoBody)

	_, ok := GetClaims(req)
	if ok {
		t.Error("Request should initially have no claims")
	}

	ctx := context.WithValue(req.Context(), ClaimsKey, originalClaims)
	req = req.WithContext(ctx)

	retrievedClaims, ok := GetClaims(req)
	if !ok {
		t.Error("Request should now have claims")
	}
	if retrievedClaims != originalClaims {
		t.Errorf("Retrieved claims = %v, want %v", retrievedClaims, originalClaims)
	}

	if retrievedClaims.UserID != "user456" {
		t.Errorf("UserID = %v, want %v", retrievedClaims.UserID, "user456")
	}
	if retrievedClaims.Role != "admin" {
		t.Errorf("Role = %v, want %v", retrievedClaims.Role, "admin")
	}
	if len(retrievedClaims.Scopes) != 1 || retrievedClaims.Scopes[0] != "*:*" {
		t.Errorf("Scopes = %v, want %v", retrievedClaims.Scopes, []string{"*:*"})
	}
}

func TestGetClaimsConcurrent(t *testing.T) {
	const numGoroutines = 10
	const numOperations = 50

	claims := &auth.Claims{
		UserID: "concurrent-user",
		Role:   "worker",
		Scopes: []string{"read:vms"},
	}

	done := make(chan bool, numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			defer func() { done <- true }()

			for range numOperations {
				ctx := context.WithValue(context.Background(), ClaimsKey, claims)
				req := httptest.NewRequestWithContext(context.Background(), "GET", "/", http.NoBody)
				req = req.WithContext(ctx)

				got, ok := GetClaims(req)
				if !ok {
					t.Errorf("Goroutine %d: GetClaims() should return true", id)
					return
				}
				if got != claims {
					t.Errorf("Goroutine %d: GetClaims() should return the same claims", id)
					return
				}

				req2 := httptest.NewRequestWithContext(context.Background(), "GET", "/", http.NoBody)
				got2, ok2 := GetClaims(req2)
				if ok2 {
					t.Errorf("Goroutine %d: GetClaims() should return false for request without claims", id)
					return
				}
				if got2 != nil {
					t.Errorf("Goroutine %d: GetClaims() should return nil for request without claims", id)
					return
				}
			}
		}(i)
	}

	for range numGoroutines {
		select {
		case <-done:
		case <-time.After(1 * time.Second):
			t.Fatal("Test timed out waiting for goroutines")
		}
	}
}

func TestGetClaimsNilContext(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/", http.NoBody)

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic with nil context: %v", r)
		}
	}()

	req = req.WithContext(context.TODO())
	_, ok := GetClaims(req)
	if ok {
		t.Error("GetClaims() should return false with empty context")
	}
}

func TestClaimsKeyUniqueness(t *testing.T) {
	otherKey := ctxKey("other_key")
	otherValue := "some_other_value"

	ctx := context.Background()
	ctx = context.WithValue(ctx, ClaimsKey, &auth.Claims{UserID: "user1"})
	ctx = context.WithValue(ctx, otherKey, otherValue)

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/", http.NoBody)
	req = req.WithContext(ctx)

	claims, ok := GetClaims(req)
	if !ok {
		t.Error("GetClaims() should return true")
	}
	if claims.UserID != "user1" {
		t.Errorf("UserID = %v, want %v", claims.UserID, "user1")
	}

	if ctx.Value(otherKey) != otherValue {
		t.Errorf("Other context value = %v, want %v", ctx.Value(otherKey), otherValue)
	}
}

func TestGetClaimsModifiedContext(t *testing.T) {
	originalClaims := &auth.Claims{
		UserID: "original-user",
		Role:   "worker",
		Scopes: []string{"read:vms"},
	}

	ctx := context.WithValue(context.Background(), ClaimsKey, originalClaims)
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/", http.NoBody)
	req = req.WithContext(ctx)

	claims, ok := GetClaims(req)
	if !ok {
		t.Error("GetClaims() should return true")
	}
	if claims.UserID != "original-user" {
		t.Errorf("UserID = %v, want %v", claims.UserID, "original-user")
	}

	claims.UserID = "modified-user"

	claims2, ok2 := GetClaims(req)
	if !ok2 {
		t.Error("GetClaims() should still return true")
	}
	if claims2.UserID != "modified-user" {
		t.Errorf("Modified UserID = %v, want %v", claims2.UserID, "modified-user")
	}
}

func TestGetClaimsWithDifferentRoles(t *testing.T) {
	roles := []string{"admin", "worker", "public"}

	for _, role := range roles {
		t.Run(role, func(t *testing.T) {
			claims := &auth.Claims{
				UserID: "testuser",
				Role:   role,
				Scopes: []string{},
			}

			ctx := context.WithValue(context.Background(), ClaimsKey, claims)
			req := httptest.NewRequestWithContext(context.Background(), "GET", "/", http.NoBody)
			req = req.WithContext(ctx)

			retrievedClaims, ok := GetClaims(req)
			if !ok {
				t.Error("GetClaims() should return true")
			}
			if retrievedClaims.Role != role {
				t.Errorf("Role = %q, want %q", retrievedClaims.Role, role)
			}
		})
	}
}

func TestGetClaimsWithMultipleScopes(t *testing.T) {
	scopes := []string{"read:vms", "write:backups", "delete:configs", "admin:system"}
	claims := &auth.Claims{
		UserID: "testuser",
		Role:   "admin",
		Scopes: scopes,
	}

	ctx := context.WithValue(context.Background(), ClaimsKey, claims)
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/", http.NoBody)
	req = req.WithContext(ctx)

	retrievedClaims, ok := GetClaims(req)
	if !ok {
		t.Error("GetClaims() should return true")
	}

	if len(retrievedClaims.Scopes) != len(scopes) {
		t.Errorf("Scopes length = %d, want %d", len(retrievedClaims.Scopes), len(scopes))
	}

	for i, scope := range retrievedClaims.Scopes {
		if scope != scopes[i] {
			t.Errorf("Scope[%d] = %q, want %q", i, scope, scopes[i])
		}
	}
}
