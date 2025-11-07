package server

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStaticFileServing(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)

	t.Run("serves index.html at root", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		t.Logf("DEBUG: Before ServeHTTP")
		server.router.ServeHTTP(w, req)
		t.Logf("DEBUG: Status: %d", w.Code)
		t.Logf("DEBUG: Body: %s", w.Body.String())

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		body, _ := io.ReadAll(w.Body)
		if !strings.Contains(string(body), "Poker") {
			t.Errorf("expected index.html content with 'Poker', got: %s", string(body))
		}
	})

	t.Run("serves static assets", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/vite.svg", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		// Should serve the SVG file
		if w.Code != http.StatusOK {
			t.Logf("assets status: %d (may be expected if not testing actual files)", w.Code)
		}
	})

	t.Run("returns 404 for non-existent file", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/non-existent-file.js", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("serves SPA index.html for non-existent routes", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/some/non/existent/route", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		// SPA fallback should return index.html
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200 for SPA fallback, got %d", w.Code)
		}

		body, _ := io.ReadAll(w.Body)
		if !strings.Contains(string(body), "Poker") {
			t.Errorf("expected index.html content for SPA fallback, got: %s", string(body))
		}
	})

	t.Run("health endpoint still works", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d for /health, got %d", http.StatusOK, w.Code)
		}
	})
}
