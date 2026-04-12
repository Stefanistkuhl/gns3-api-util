package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	sharedpb "github.com/0xveya/gns3util/internal/shared/pb"
	"github.com/0xveya/gns3util/pkg/web/auth"
)

type stubRBACChecker struct {
	allowed  bool
	err      error
	called   bool
	action   sharedpb.Action
	resource sharedpb.Resource
}

func (s *stubRBACChecker) CheckPermission(
	_ context.Context,
	_ string,
	_ string,
	_ []string,
	action sharedpb.Action,
	resource sharedpb.Resource,
) (bool, error) {
	s.called = true
	s.action = action
	s.resource = resource
	return s.allowed, s.err
}

func TestRequirePermissionAllowsJobsExecute(t *testing.T) {
	ctx := context.Background()
	checker := &stubRBACChecker{allowed: true}
	calledNext := false
	handler := RequirePermission(checker, sharedpb.Action_ACTION_EXECUTE, sharedpb.Resource_RESOURCE_JOBS)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calledNext = true
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/jobs/test/run", http.NoBody)
	req = req.WithContext(context.WithValue(req.Context(), ClaimsKey, &auth.Claims{
		UserID: "alice",
		Role:   "worker",
	}))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}
	if !calledNext {
		t.Fatal("next handler was not called")
	}
	if !checker.called {
		t.Fatal("checker was not called")
	}
	if checker.action != sharedpb.Action_ACTION_EXECUTE || checker.resource != sharedpb.Resource_RESOURCE_JOBS {
		t.Fatalf("checker called with %v/%v, want %v/%v", checker.action, checker.resource, sharedpb.Action_ACTION_EXECUTE, sharedpb.Resource_RESOURCE_JOBS)
	}
}

func TestRequirePermissionDeniesJobsExecute(t *testing.T) {
	ctx := context.Background()
	checker := &stubRBACChecker{allowed: false}
	handler := RequirePermission(checker, sharedpb.Action_ACTION_EXECUTE, sharedpb.Resource_RESOURCE_JOBS)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/jobs/test/run", http.NoBody)
	req = req.WithContext(context.WithValue(req.Context(), ClaimsKey, &auth.Claims{
		UserID: "alice",
		Role:   "worker",
	}))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}
