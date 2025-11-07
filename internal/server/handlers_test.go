package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheckHandler(t *testing.T) {
	logger := slog.Default()
	handler := HealthCheckHandler(logger)

	// Create a test HTTP request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Call the handler
	handler(w, req)

	// Check response status code
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Check response content type
	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	// Verify response body
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}

	// Check that status field exists and is "ok"
	status, ok := response["status"]
	if !ok {
		t.Fatal("expected 'status' field in response")
	}

	if status != "ok" {
		t.Errorf("expected status 'ok', got %v", status)
	}
}
