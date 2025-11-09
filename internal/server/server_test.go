package server

import (
	"context"
	"encoding/json"
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

// TestBroadcastActionRequest verifies action_request is sent to all clients at table
func TestBroadcastActionRequest(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Start the hub
	go hub.Run()

	// Create two mock clients with send channels
	client1 := &Client{
		hub:   hub,
		send:  make(chan []byte, 256),
		Token: "token1",
	}
	client2 := &Client{
		hub:   hub,
		send:  make(chan []byte, 256),
		Token: "token2",
	}

	// Register clients in hub
	hub.register <- client1
	hub.register <- client2
	time.Sleep(50 * time.Millisecond)

	// Seat both clients at table-1
	table := server.tables[0]
	table.Seats[0].Token = &client1.Token
	table.Seats[0].Status = "active"
	table.Seats[1].Token = &client2.Token
	table.Seats[1].Status = "active"

	// Create mock session manager and add sessions for both clients
	sm := server.sessionManager
	session1, _ := sm.CreateSession("Player1")
	session2, _ := sm.CreateSession("Player2")
	client1.Token = session1.Token
	client2.Token = session2.Token
	sm.UpdateSession(session1.Token, &table.ID, &table.Seats[0].Index)
	sm.UpdateSession(session2.Token, &table.ID, &table.Seats[1].Index)

	// Update client tokens in hub
	hub.mu.Lock()
	hub.clients[client1] = true
	hub.clients[client2] = true
	hub.mu.Unlock()

	// Call BroadcastActionRequest
	err := server.BroadcastActionRequest(table.ID, 0, []string{"call", "fold"}, 20, 10, 100)
	if err != nil {
		t.Fatalf("BroadcastActionRequest failed: %v", err)
	}

	// Give time for broadcast
	time.Sleep(50 * time.Millisecond)

	// Verify both clients received the message
	select {
	case msg := <-client1.send:
		var wsMsg WebSocketMessage
		err := json.Unmarshal(msg, &wsMsg)
		if err != nil {
			t.Errorf("client1: failed to unmarshal message: %v", err)
		}
		if wsMsg.Type != "action_request" {
			t.Errorf("client1: expected message type 'action_request', got %q", wsMsg.Type)
		}
	default:
		t.Error("client1: did not receive action_request message")
	}

	select {
	case msg := <-client2.send:
		var wsMsg WebSocketMessage
		err := json.Unmarshal(msg, &wsMsg)
		if err != nil {
			t.Errorf("client2: failed to unmarshal message: %v", err)
		}
		if wsMsg.Type != "action_request" {
			t.Errorf("client2: expected message type 'action_request', got %q", wsMsg.Type)
		}
	default:
		t.Error("client2: did not receive action_request message")
	}
}

// TestBroadcastActionResult verifies action_result is sent to all clients at table
func TestBroadcastActionResult(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Start the hub
	go hub.Run()

	// Create two mock clients with send channels
	client1 := &Client{
		hub:   hub,
		send:  make(chan []byte, 256),
		Token: "token1",
	}
	client2 := &Client{
		hub:   hub,
		send:  make(chan []byte, 256),
		Token: "token2",
	}

	// Register clients in hub
	hub.register <- client1
	hub.register <- client2
	time.Sleep(50 * time.Millisecond)

	// Seat both clients at table-1
	table := server.tables[0]
	table.Seats[0].Token = &client1.Token
	table.Seats[0].Status = "active"
	table.Seats[1].Token = &client2.Token
	table.Seats[1].Status = "active"

	// Create mock session manager and add sessions for both clients
	sm := server.sessionManager
	session1, _ := sm.CreateSession("Player1")
	session2, _ := sm.CreateSession("Player2")
	client1.Token = session1.Token
	client2.Token = session2.Token
	sm.UpdateSession(session1.Token, &table.ID, &table.Seats[0].Index)
	sm.UpdateSession(session2.Token, &table.ID, &table.Seats[1].Index)

	// Update client tokens in hub
	hub.mu.Lock()
	hub.clients[client1] = true
	hub.clients[client2] = true
	hub.mu.Unlock()

	// Call BroadcastActionResult
	nextActor := 1
	err := server.BroadcastActionResult(table.ID, 0, "call", 20, 80, 150, &nextActor, false, nil)
	if err != nil {
		t.Fatalf("BroadcastActionResult failed: %v", err)
	}

	// Give time for broadcast
	time.Sleep(50 * time.Millisecond)

	// Verify both clients received the message
	select {
	case msg := <-client1.send:
		var wsMsg WebSocketMessage
		err := json.Unmarshal(msg, &wsMsg)
		if err != nil {
			t.Errorf("client1: failed to unmarshal message: %v", err)
		}
		if wsMsg.Type != "action_result" {
			t.Errorf("client1: expected message type 'action_result', got %q", wsMsg.Type)
		}
		var payload ActionResultPayload
		err = json.Unmarshal(wsMsg.Payload, &payload)
		if err != nil {
			t.Errorf("client1: failed to unmarshal payload: %v", err)
		}
		if payload.SeatIndex != 0 {
			t.Errorf("client1: expected seatIndex 0, got %d", payload.SeatIndex)
		}
		if payload.Action != "call" {
			t.Errorf("client1: expected action 'call', got %q", payload.Action)
		}
	default:
		t.Error("client1: did not receive action_result message")
	}

	select {
	case msg := <-client2.send:
		var wsMsg WebSocketMessage
		err := json.Unmarshal(msg, &wsMsg)
		if err != nil {
			t.Errorf("client2: failed to unmarshal message: %v", err)
		}
		if wsMsg.Type != "action_result" {
			t.Errorf("client2: expected message type 'action_result', got %q", wsMsg.Type)
		}
	default:
		t.Error("client2: did not receive action_result message")
	}
}
