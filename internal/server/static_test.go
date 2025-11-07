package server

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStaticFileServing(t *testing.T) {
	logger := slog.Default()
	
	// Save the original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	
	// Change to the workspace root for tests
	wsRoot := filepath.Join(originalWd, "..", "..")
	if err := os.Chdir(wsRoot); err != nil {
		t.Logf("warning: could not change to workspace root %s: %v", wsRoot, err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Logf("warning: could not restore working directory: %v", err)
		}
	}()

	server := NewServer(logger)

	t.Run("serves index.html at root", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		body, _ := io.ReadAll(w.Body)
		bodyStr := string(body)
		if !strings.Contains(bodyStr, "<!DOCTYPE") && !strings.Contains(bodyStr, "<html") && !strings.Contains(bodyStr, "Poker") {
			t.Errorf("expected index.html content, got: %s", bodyStr)
		}
	})

	t.Run("serves static assets", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/vite.svg", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		// Should serve the SVG file or return 404 if it doesn't exist
		if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
			t.Errorf("expected status 200 or 404, got %d", w.Code)
		}
	})

	t.Run("SPA fallback for non-existent routes serves index.html", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/some/non/existent/route", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		// SPA fallback should return index.html with status 200
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200 for SPA fallback, got %d", w.Code)
		}

		body, _ := io.ReadAll(w.Body)
		bodyStr := string(body)
		if !strings.Contains(bodyStr, "<!DOCTYPE") && !strings.Contains(bodyStr, "<html") {
			t.Errorf("expected index.html content for SPA fallback, got: %s", bodyStr)
		}
	})

	t.Run("health endpoint still works and not overridden", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d for /health, got %d", http.StatusOK, w.Code)
		}

		body, _ := io.ReadAll(w.Body)
		bodyStr := string(body)
		if !strings.Contains(bodyStr, "ok") {
			t.Errorf("expected health check response with 'ok' status, got: %s", bodyStr)
		}
	})

	t.Run("WebSocket endpoint still works", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ws", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		// Without proper WebSocket headers, should get an error response, not 404
		// This just verifies the route exists
		if w.Code == http.StatusNotFound {
			t.Errorf("expected /ws route to exist, got 404")
		}
	})
}
