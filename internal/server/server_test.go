package server

import (
	"context"
	"log/slog"
	"net/http"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)

	if server == nil {
		t.Fatal("expected server to be initialized, got nil")
	}

	if server.router == nil {
		t.Fatal("expected router to be initialized, got nil")
	}

	if server.logger == nil {
		t.Fatal("expected logger to be initialized, got nil")
	}

	if server.upgrader == nil {
		t.Fatal("expected websocket upgrader to be initialized, got nil")
	}

	if server.hub == nil {
		t.Fatal("expected hub to be initialized, got nil")
	}
}

func TestServerStart(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)

	// Channel to signal that server started
	done := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		err := server.Start("127.0.0.1:0") // Use port 0 for automatic port assignment
		if err != nil && err != http.ErrServerClosed {
			done <- err
		}
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Verify the server is listening by using the Shutdown method
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		t.Errorf("error shutting down server: %v", err)
	}

	// Wait for the server to finish
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("server returned error: %v", err)
		}
	case <-time.After(1 * time.Second):
		// Server stopped cleanly
	}
}
