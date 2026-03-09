package commonhandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandleHealthz(t *testing.T) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", "/healthz", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	HandleHealthz(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expectedContentType := "application/json"
	if ct := rr.Header().Get("Content-Type"); ct != expectedContentType {
		t.Errorf("handler returned wrong content type: got %v want %v",
			ct, expectedContentType)
	}

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "OK" {
		t.Errorf("handler returned unexpected body: got %v want %v",
			response["status"], "OK")
	}
}

func TestHandleHealthzMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req, err := http.NewRequestWithContext(context.Background(), method, "/healthz", http.NoBody)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			HandleHealthz(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code for %s: got %v want %v",
					method, status, http.StatusOK)
			}

			expectedContentType := "application/json"
			if ct := rr.Header().Get("Content-Type"); ct != expectedContentType {
				t.Errorf("handler returned wrong content type for %s: got %v want %v",
					method, ct, expectedContentType)
			}
		})
	}
}

func TestHandleHealthzResponseStructure(t *testing.T) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", "/healthz", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	HandleHealthz(rr, req)

	var response map[string]any
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if len(response) != 1 {
		t.Errorf("Response should have exactly one field, got %d", len(response))
	}

	if _, ok := response["status"]; !ok {
		t.Error("Response should have 'status' field")
	}

	status, ok := response["status"].(string)
	if !ok {
		t.Error("Status should be a string")
	}

	if status != "OK" {
		t.Errorf("Status should be 'OK', got '%s'", status)
	}
}

func TestHandleHealthzHeaders(t *testing.T) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", "/healthz", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	HandleHealthz(rr, req)

	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", ct)
	}

	if rr.Body.Len() == 0 {
		t.Error("Response body should not be empty")
	}

	if !json.Valid(rr.Body.Bytes()) {
		t.Error("Response should be valid JSON")
	}
}

func TestHandleHealthzConcurrent(t *testing.T) {
	const numGoroutines = 10
	const numRequests = 100

	done := make(chan bool, numGoroutines)

	for range numGoroutines {
		go func() {
			defer func() { done <- true }()

			for range numRequests {
				req, err := http.NewRequestWithContext(context.Background(), "GET", "/healthz", http.NoBody)
				if err != nil {
					t.Errorf("Failed to create request: %v", err)
					return
				}

				rr := httptest.NewRecorder()
				HandleHealthz(rr, req)

				if status := rr.Code; status != http.StatusOK {
					t.Errorf("Expected status 200, got %d", status)
					return
				}

				var response map[string]string
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
					return
				}

				if response["status"] != "OK" {
					t.Errorf("Expected status OK, got %s", response["status"])
					return
				}
			}
		}()
	}

	for range numGoroutines {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out waiting for goroutines")
		}
	}
}

func TestHandleHealthzWithQueryParams(t *testing.T) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", "/healthz?verbose=true&format=json", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	HandleHealthz(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "OK" {
		t.Errorf("handler returned unexpected body: got %v want %v",
			response["status"], "OK")
	}
}

func TestHandleHealthzWithHeaders(t *testing.T) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", "/healthz", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Custom-Header", "test-value")

	rr := httptest.NewRecorder()
	HandleHealthz(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "OK" {
		t.Errorf("handler returned unexpected body: got %v want %v",
			response["status"], "OK")
	}
}
