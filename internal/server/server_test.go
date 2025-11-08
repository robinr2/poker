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

// TestServerFindPlayerSeat verifies FindPlayerSeat finds player across all 4 tables
func TestServerFindPlayerSeat(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)

	token1 := "player-1"
	token2 := "player-2"
	token3 := "player-3"

	// Assign token1 to table 0, seat 0
	table0 := server.tables[0]
	seat1, _ := table0.AssignSeat(&token1)

	// Assign token2 to table 1, seat 1
	table1 := server.tables[1]
	seat2, _ := table1.AssignSeat(&token2)

	// Assign token3 to table 3, seat 2
	table3 := server.tables[3]
	seat3, _ := table3.AssignSeat(&token3)

	// Find token1 - should return seat from table 0
	foundSeat := server.FindPlayerSeat(&token1)
	if foundSeat == nil {
		t.Fatal("expected to find token1, got nil")
	}

	if foundSeat.Index != seat1.Index {
		t.Errorf("expected seat index %d, got %d", seat1.Index, foundSeat.Index)
	}

	if foundSeat.Token == nil || *foundSeat.Token != token1 {
		t.Errorf("expected token '%s', got %v", token1, foundSeat.Token)
	}

	// Find token2 - should return seat from table 1
	foundSeat = server.FindPlayerSeat(&token2)
	if foundSeat == nil {
		t.Fatal("expected to find token2, got nil")
	}

	if foundSeat.Index != seat2.Index {
		t.Errorf("expected seat index %d for token2, got %d", seat2.Index, foundSeat.Index)
	}

	if foundSeat.Token == nil || *foundSeat.Token != token2 {
		t.Errorf("expected token '%s', got %v", token2, foundSeat.Token)
	}

	// Find token3 - should return seat from table 3
	foundSeat = server.FindPlayerSeat(&token3)
	if foundSeat == nil {
		t.Fatal("expected to find token3, got nil")
	}

	if foundSeat.Index != seat3.Index {
		t.Errorf("expected seat index %d for token3, got %d", seat3.Index, foundSeat.Index)
	}

	if foundSeat.Token == nil || *foundSeat.Token != token3 {
		t.Errorf("expected token '%s', got %v", token3, foundSeat.Token)
	}
}

// TestServerFindPlayerSeatNotFound verifies returns nil when player not seated
func TestServerFindPlayerSeatNotFound(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)

	tokenNotSeated := "player-not-seated"

	// Try to find player who is not seated
	foundSeat := server.FindPlayerSeat(&tokenNotSeated)
	if foundSeat != nil {
		t.Errorf("expected nil for unseated player, got %v", foundSeat)
	}

	// Seat one player, then search for different token
	token1 := "player-1"
	server.tables[0].AssignSeat(&token1)

	foundSeat = server.FindPlayerSeat(&tokenNotSeated)
	if foundSeat != nil {
		t.Errorf("expected nil for different token, got %v", foundSeat)
	}
}
