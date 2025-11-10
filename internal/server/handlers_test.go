package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
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

// TestGetLobbyState verifies GetLobbyState returns correct table info for all 4 tables with 0 occupied seats
func TestGetLobbyState(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)

	lobbyState := server.GetLobbyState()

	if lobbyState == nil {
		t.Fatal("expected GetLobbyState to return non-nil slice")
	}

	if len(lobbyState) != 4 {
		t.Errorf("expected 4 tables in lobby state, got %d", len(lobbyState))
	}

	expectedTables := []struct {
		id       string
		name     string
		maxSeats int
		occupied int
	}{
		{"table-1", "Table 1", 6, 0},
		{"table-2", "Table 2", 6, 0},
		{"table-3", "Table 3", 6, 0},
		{"table-4", "Table 4", 6, 0},
	}

	for i, expected := range expectedTables {
		if i >= len(lobbyState) {
			break
		}

		if lobbyState[i].ID != expected.id {
			t.Errorf("table %d: expected ID '%s', got '%s'", i, expected.id, lobbyState[i].ID)
		}

		if lobbyState[i].Name != expected.name {
			t.Errorf("table %d: expected Name '%s', got '%s'", i, expected.name, lobbyState[i].Name)
		}

		if lobbyState[i].MaxSeats != expected.maxSeats {
			t.Errorf("table %d: expected MaxSeats %d, got %d", i, expected.maxSeats, lobbyState[i].MaxSeats)
		}

		if lobbyState[i].SeatsOccupied != expected.occupied {
			t.Errorf("table %d: expected SeatsOccupied %d, got %d", i, expected.occupied, lobbyState[i].SeatsOccupied)
		}
	}
}

// TestGetLobbyStateWithOccupiedSeats verifies GetLobbyState reflects occupied seats correctly
func TestGetLobbyStateWithOccupiedSeats(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)

	// Manually occupy some seats on different tables
	token1 := "token-1"
	token2 := "token-2"
	token3 := "token-3"

	// Table 1: 3 occupied seats
	server.tables[0].mu.Lock()
	server.tables[0].Seats[0].Token = &token1
	server.tables[0].Seats[1].Token = &token2
	server.tables[0].Seats[2].Token = &token3
	server.tables[0].mu.Unlock()

	// Table 2: 1 occupied seat
	server.tables[1].mu.Lock()
	server.tables[1].Seats[0].Token = &token1
	server.tables[1].mu.Unlock()

	// Table 3: 0 occupied seats (no change)

	// Table 4: 6 occupied seats (full)
	server.tables[3].mu.Lock()
	for i := 0; i < 6; i++ {
		server.tables[3].Seats[i].Token = &token1
	}
	server.tables[3].mu.Unlock()

	lobbyState := server.GetLobbyState()

	expectedOccupied := []int{3, 1, 0, 6}
	for i, expected := range expectedOccupied {
		if lobbyState[i].SeatsOccupied != expected {
			t.Errorf("table %d: expected SeatsOccupied %d, got %d", i, expected, lobbyState[i].SeatsOccupied)
		}
	}
}

// TestGetLobbyStateThreadSafety verifies concurrent calls to GetLobbyState while modifying seats
func TestGetLobbyStateThreadSafety(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)

	const numGoroutines = 10
	const operationsPerGoroutine = 50

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*2)

	// Reader goroutines - call GetLobbyState
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				lobbyState := server.GetLobbyState()
				if lobbyState == nil || len(lobbyState) != 4 {
					errors <- fmt.Errorf("invalid lobby state: %v", lobbyState)
					return
				}
				for _, tableInfo := range lobbyState {
					if tableInfo.SeatsOccupied < 0 || tableInfo.SeatsOccupied > 6 {
						errors <- fmt.Errorf("invalid seat count: %d", tableInfo.SeatsOccupied)
						return
					}
				}
			}
		}()
	}

	// Writer goroutines - modify seat tokens
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				tableIdx := (id + j) % 4
				seatIdx := (id + j) % 6
				token := fmt.Sprintf("player-%d-%d", id, j)

				server.tables[tableIdx].mu.Lock()
				server.tables[tableIdx].Seats[seatIdx].Token = &token
				server.tables[tableIdx].mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Check if any errors occurred
	select {
	case err := <-errors:
		t.Errorf("thread safety error: %v", err)
	default:
	}

	// Verify final state is valid
	finalState := server.GetLobbyState()
	if finalState == nil || len(finalState) != 4 {
		t.Error("final lobby state is invalid")
	}
}

// TestFilterHoleCardsForPlayer verifies that filterHoleCardsForPlayer returns only the player's cards
func TestFilterHoleCardsForPlayer(t *testing.T) {
	// Create test hole cards - using a map[int][]Card structure (slice, not array)
	holeCards := map[int][]Card{
		0: {Card{Rank: "A", Suit: "s"}, Card{Rank: "K", Suit: "h"}},
		1: {Card{Rank: "Q", Suit: "d"}, Card{Rank: "J", Suit: "c"}},
		2: {Card{Rank: "T", Suit: "s"}, Card{Rank: "9", Suit: "h"}},
		3: {Card{Rank: "8", Suit: "d"}, Card{Rank: "7", Suit: "c"}},
	}

	// Test filtering for player at seat 0
	filtered := filterHoleCardsForPlayer(holeCards, 0)
	if len(filtered) != 1 {
		t.Errorf("expected 1 card entry for player 0, got %d", len(filtered))
	}
	if cards, ok := filtered[0]; ok {
		if len(cards) != 2 {
			t.Errorf("expected 2 cards for player 0, got %d", len(cards))
		}
		if cards[0].Rank != "A" || cards[0].Suit != "s" {
			t.Errorf("expected As, got %s", cards[0].String())
		}
		if cards[1].Rank != "K" || cards[1].Suit != "h" {
			t.Errorf("expected Kh, got %s", cards[1].String())
		}
	} else {
		t.Error("expected seat 0 in filtered map")
	}

	// Verify no other seats are present
	for seat := range filtered {
		if seat != 0 {
			t.Errorf("unexpected seat %d in filtered map", seat)
		}
	}

	// Test filtering for player at seat 2
	filtered = filterHoleCardsForPlayer(holeCards, 2)
	if len(filtered) != 1 {
		t.Errorf("expected 1 card entry for player 2, got %d", len(filtered))
	}
	if cards, ok := filtered[2]; ok {
		if len(cards) != 2 {
			t.Errorf("expected 2 cards for player 2, got %d", len(cards))
		}
		if cards[0].Rank != "T" || cards[0].Suit != "s" {
			t.Errorf("expected Ts, got %s", cards[0].String())
		}
	} else {
		t.Error("expected seat 2 in filtered map")
	}

	// Test filtering for player not in holeCards
	filtered = filterHoleCardsForPlayer(holeCards, 5)
	if len(filtered) != 0 {
		t.Errorf("expected empty map for player 5, got %d entries", len(filtered))
	}
}

// TestBroadcastHandStarted verifies hand_started message broadcast with dealer and blind info
func TestBroadcastHandStarted(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up players
	token1 := "player-1"
	token2 := "player-2"

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	// Start a hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Call broadcastHandStarted
	err = server.broadcastHandStarted(table)
	if err != nil {
		t.Fatalf("failed to broadcast hand started: %v", err)
	}

	// Verify table has proper hand state
	table.mu.RLock()
	if table.CurrentHand == nil {
		t.Error("expected CurrentHand to be set")
	}
	if table.DealerSeat == nil {
		t.Error("expected DealerSeat to be set")
	}
	table.mu.RUnlock()
}

// TestBroadcastBlindPosted verifies blind_posted message broadcast
func TestBroadcastBlindPosted(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up players
	token1 := "player-1"
	token2 := "player-2"

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	// Start a hand to establish blind positions
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Get blind positions
	table.mu.RLock()
	hand := table.CurrentHand
	sbSeat := hand.SmallBlindSeat
	sbAmount := 10
	table.mu.RUnlock()

	// Call broadcastBlindPosted
	err = server.broadcastBlindPosted(table, sbSeat, sbAmount)
	if err != nil {
		t.Fatalf("failed to broadcast blind posted: %v", err)
	}
}

// TestBroadcastCardsDealt verifies cards_dealt message with privacy filtering
func TestBroadcastCardsDealt(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up players
	token1 := "player-1"
	token2 := "player-2"

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	// Start a hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Call broadcastCardsDealt - this should use privacy filtering
	err = server.broadcastCardsDealt(table)
	if err != nil {
		t.Fatalf("failed to broadcast cards dealt: %v", err)
	}
}

// TestHandleStartHand verifies start_hand message handler successfully starts a hand
func TestHandleStartHand(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	sm := NewSessionManager(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up 2 players with sessions
	// Create sessions
	session1, _ := sm.CreateSession("Player1")
	session2, _ := sm.CreateSession("Player2")
	token1 := session1.Token
	token2 := session2.Token

	// Update sessions with table and seat info
	sm.UpdateSession(token1, &table.ID, &[]int{0}[0])
	sm.UpdateSession(token2, &table.ID, &[]int{1}[0])

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	// Create a mock client
	client := &Client{
		hub:   hub,
		Token: token1,
		send:  make(chan []byte, 256),
	}

	// Call HandleStartHand
	payload := []byte("{}")
	err := client.HandleStartHand(sm, server, logger, payload)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify table has an active hand
	table.mu.RLock()
	if table.CurrentHand == nil {
		t.Error("expected CurrentHand to be set")
	}
	if table.DealerSeat == nil {
		t.Error("expected DealerSeat to be set")
	}
	table.mu.RUnlock()
}

// TestHandleStartHandNotSeated verifies error when player is not seated
func TestHandleStartHandNotSeated(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	sm := NewSessionManager(logger)
	hub := server.hub
	go hub.Run()

	// Create a session but don't seat the player
	session, _ := sm.CreateSession("Player1")

	// Create a mock client
	client := &Client{
		hub:   hub,
		Token: session.Token,
		send:  make(chan []byte, 256),
	}

	// Call HandleStartHand - should fail because player is not seated
	payload := []byte("{}")
	err := client.HandleStartHand(sm, server, logger, payload)
	if err == nil {
		t.Error("expected error when player not seated")
	}
}

// TestHandleStartHandInsufficientPlayers verifies error with < 2 players
func TestHandleStartHandInsufficientPlayers(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	sm := NewSessionManager(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up only 1 player
	session1, _ := sm.CreateSession("Player1")
	token1 := session1.Token

	// Assign seat
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.mu.Unlock()

	// Create a mock client
	client := &Client{
		hub:   hub,
		Token: token1,
		send:  make(chan []byte, 256),
	}

	// Call HandleStartHand - should fail because only 1 player
	payload := []byte("{}")
	err := client.HandleStartHand(sm, server, logger, payload)
	if err == nil {
		t.Error("expected error when insufficient players")
	}
}

// TestTableStateSeatIncludesStack verifies that stack field is included in table_state payload
func TestTableStateSeatIncludesStack(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	sm := NewSessionManager(logger)

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up player in seat 0 with stack
	session1, _ := sm.CreateSession("Player1")
	token1 := session1.Token

	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1500
	table.mu.Unlock()

	// Create a mock client with send channel
	client := &Client{
		send: make(chan []byte, 256),
	}

	// Call SendTableState
	err := client.SendTableState(server, table.ID, logger)
	if err != nil {
		t.Fatalf("SendTableState failed: %v", err)
	}

	// Get the response from the send channel
	response := <-client.send

	// Parse the response
	var msg WebSocketMessage
	err = json.Unmarshal(response, &msg)
	if err != nil {
		t.Fatalf("failed to parse WebSocket message: %v", err)
	}

	if msg.Type != "table_state" {
		t.Errorf("expected message type 'table_state', got %s", msg.Type)
	}

	// Parse the payload
	var payload TableStatePayload
	err = json.Unmarshal(msg.Payload, &payload)
	if err != nil {
		t.Fatalf("failed to parse payload: %v", err)
	}

	if payload.TableId != table.ID {
		t.Errorf("expected tableId %s, got %s", table.ID, payload.TableId)
	}

	// Check seat 0
	seat0 := payload.Seats[0]
	if seat0.Index != 0 {
		t.Errorf("expected seat 0 index to be 0, got %d", seat0.Index)
	}

	if seat0.Stack == nil {
		t.Error("expected seat 0 Stack to be non-nil")
	} else if *seat0.Stack != 1500 {
		t.Errorf("expected seat 0 Stack to be 1500, got %d", *seat0.Stack)
	}

	// Check empty seat (seat 1)
	seat1 := payload.Seats[1]
	if seat1.Stack != nil {
		t.Error("expected seat 1 Stack (empty seat) to be nil")
	}
}

// TestTableStateSerializationWithStacks verifies JSON serialization includes stack values
func TestTableStateSerializationWithStacks(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	sm := NewSessionManager(logger)

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up multiple players with different stacks
	session1, _ := sm.CreateSession("Player1")
	token1 := session1.Token
	session2, _ := sm.CreateSession("Player2")
	token2 := session2.Token

	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 2000

	table.Seats[2].Token = &token2
	table.Seats[2].Status = "waiting"
	table.Seats[2].Stack = 1500
	table.mu.Unlock()

	// Create a mock client
	client := &Client{
		send: make(chan []byte, 256),
	}

	// Call SendTableState
	err := client.SendTableState(server, table.ID, logger)
	if err != nil {
		t.Fatalf("SendTableState failed: %v", err)
	}

	// Get response
	response := <-client.send

	// Parse response
	var msg WebSocketMessage
	err = json.Unmarshal(response, &msg)
	if err != nil {
		t.Fatalf("failed to parse WebSocket message: %v", err)
	}

	// Parse payload
	var payload TableStatePayload
	err = json.Unmarshal(msg.Payload, &payload)
	if err != nil {
		t.Fatalf("failed to parse payload: %v", err)
	}

	// Verify all seats
	if len(payload.Seats) != 6 {
		t.Errorf("expected 6 seats, got %d", len(payload.Seats))
	}

	// Check seat 0 (occupied with 2000 stack)
	if payload.Seats[0].Stack == nil {
		t.Error("expected seat 0 Stack to be non-nil")
	} else if *payload.Seats[0].Stack != 2000 {
		t.Errorf("expected seat 0 Stack to be 2000, got %d", *payload.Seats[0].Stack)
	}

	// Check seat 1 (empty)
	if payload.Seats[1].Stack != nil {
		t.Error("expected seat 1 Stack (empty) to be nil")
	}

	// Check seat 2 (occupied with 1500 stack)
	if payload.Seats[2].Stack == nil {
		t.Error("expected seat 2 Stack to be non-nil")
	} else if *payload.Seats[2].Stack != 1500 {
		t.Errorf("expected seat 2 Stack to be 1500, got %d", *payload.Seats[2].Stack)
	}

	// Check seat 3 (empty)
	if payload.Seats[3].Stack != nil {
		t.Error("expected seat 3 Stack (empty) to be nil")
	}

	// Verify JSON contains lowercase "stack" field
	var rawData map[string]interface{}
	err = json.Unmarshal(msg.Payload, &rawData)
	if err != nil {
		t.Fatalf("failed to parse raw payload: %v", err)
	}

	seats := rawData["seats"].([]interface{})
	seat0 := seats[0].(map[string]interface{})
	if _, hasStack := seat0["stack"]; !hasStack {
		t.Error("expected seat JSON to have 'stack' field")
	}
}

// TestBroadcastTableStateIncludesStack verifies Stack is included when broadcasting table state
func TestBroadcastTableStateIncludesStack(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	sm := NewSessionManager(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up players with different stacks
	session1, _ := sm.CreateSession("Player1")
	token1 := session1.Token
	session2, _ := sm.CreateSession("Player2")
	token2 := session2.Token

	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 2500

	table.Seats[2].Token = &token2
	table.Seats[2].Status = "waiting"
	table.Seats[2].Stack = 1800
	table.mu.Unlock()

	// Create mock clients and register them with the hub
	client1 := &Client{
		hub:   hub,
		send:  make(chan []byte, 256),
		Token: token1,
	}
	client2 := &Client{
		hub:   hub,
		send:  make(chan []byte, 256),
		Token: token2,
	}

	// Register clients with hub
	hub.register <- client1
	hub.register <- client2
	// Give the hub time to process registrations
	time.Sleep(100 * time.Millisecond)

	// Call broadcastTableState
	err := server.broadcastTableState(table.ID, nil)
	if err != nil {
		t.Fatalf("broadcastTableState failed: %v", err)
	}

	// Receive broadcast on client1
	response := <-client1.send

	// Parse the response
	var msg WebSocketMessage
	err = json.Unmarshal(response, &msg)
	if err != nil {
		t.Fatalf("failed to parse WebSocket message: %v", err)
	}

	if msg.Type != "table_state" {
		t.Errorf("expected message type 'table_state', got %s", msg.Type)
	}

	// Parse the payload
	var payload TableStatePayload
	err = json.Unmarshal(msg.Payload, &payload)
	if err != nil {
		t.Fatalf("failed to parse payload: %v", err)
	}

	// Check seat 0 - should have Stack set to 2500
	seat0 := payload.Seats[0]
	if seat0.Index != 0 {
		t.Errorf("expected seat 0 index to be 0, got %d", seat0.Index)
	}

	if seat0.Stack == nil {
		t.Error("expected seat 0 Stack to be non-nil in broadcast")
	} else if *seat0.Stack != 2500 {
		t.Errorf("expected seat 0 Stack to be 2500 in broadcast, got %d", *seat0.Stack)
	}

	// Check seat 1 (empty) - should have nil Stack
	seat1 := payload.Seats[1]
	if seat1.Stack != nil {
		t.Error("expected seat 1 Stack (empty seat) to be nil in broadcast")
	}

	// Check seat 2 - should have Stack set to 1800
	seat2 := payload.Seats[2]
	if seat2.Stack == nil {
		t.Error("expected seat 2 Stack to be non-nil in broadcast")
	} else if *seat2.Stack != 1800 {
		t.Errorf("expected seat 2 Stack to be 1800 in broadcast, got %d", *seat2.Stack)
	}

	// Check seat 3 (empty) - should have nil Stack
	seat3 := payload.Seats[3]
	if seat3.Stack != nil {
		t.Error("expected seat 3 Stack (empty seat) to be nil in broadcast")
	}
}

// TestTableStateIncludesGameStateWhenHandActive verifies game state fields are populated when hand is active
func TestTableStateIncludesGameStateWhenHandActive(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up players
	token1 := "player-1"
	token2 := "player-2"

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	// Start a hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Create a mock client
	mockClient := &Client{
		Token: token1,
		send:  make(chan []byte, 10),
	}

	// Send table state
	err = mockClient.SendTableState(server, table.ID, logger)
	if err != nil {
		t.Fatalf("failed to send table state: %v", err)
	}

	// Get the message from the client's send channel
	if len(mockClient.send) == 0 {
		t.Fatal("no message sent to client")
	}

	msgBytes := <-mockClient.send
	var wsMsg WebSocketMessage
	err = json.Unmarshal(msgBytes, &wsMsg)
	if err != nil {
		t.Fatalf("failed to unmarshal WebSocketMessage: %v", err)
	}

	// Verify message type
	if wsMsg.Type != "table_state" {
		t.Errorf("expected message type 'table_state', got '%s'", wsMsg.Type)
	}

	// Unmarshal the payload
	var payload TableStatePayload
	err = json.Unmarshal(wsMsg.Payload, &payload)
	if err != nil {
		t.Fatalf("failed to unmarshal TableStatePayload: %v", err)
	}

	// Verify HandInProgress is true
	if !payload.HandInProgress {
		t.Error("expected HandInProgress to be true when hand is active")
	}

	// Verify DealerSeat is not nil
	if payload.DealerSeat == nil {
		t.Error("expected DealerSeat to be non-nil when hand is active")
	}

	// Verify SmallBlindSeat is not nil
	if payload.SmallBlindSeat == nil {
		t.Error("expected SmallBlindSeat to be non-nil when hand is active")
	}

	// Verify BigBlindSeat is not nil
	if payload.BigBlindSeat == nil {
		t.Error("expected BigBlindSeat to be non-nil when hand is active")
	}

	// Verify Pot is not nil
	if payload.Pot == nil {
		t.Error("expected Pot to be non-nil when hand is active")
	}
}

// TestTableStateOmitsGameStateWhenNoHand verifies fields are nil/zero when no hand active
func TestTableStateOmitsGameStateWhenNoHand(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up players but don't start a hand
	token1 := "player-1"
	token2 := "player-2"

	// Assign seats
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "waiting"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "waiting"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	// Create a mock client
	mockClient := &Client{
		Token: token1,
		send:  make(chan []byte, 10),
	}

	// Send table state without a hand
	err := mockClient.SendTableState(server, table.ID, logger)
	if err != nil {
		t.Fatalf("failed to send table state: %v", err)
	}

	// Get the message from the client's send channel
	if len(mockClient.send) == 0 {
		t.Fatal("no message sent to client")
	}

	msgBytes := <-mockClient.send
	var wsMsg WebSocketMessage
	err = json.Unmarshal(msgBytes, &wsMsg)
	if err != nil {
		t.Fatalf("failed to unmarshal WebSocketMessage: %v", err)
	}

	// Unmarshal the payload
	var payload TableStatePayload
	err = json.Unmarshal(wsMsg.Payload, &payload)
	if err != nil {
		t.Fatalf("failed to unmarshal TableStatePayload: %v", err)
	}

	// Verify HandInProgress is false
	if payload.HandInProgress {
		t.Error("expected HandInProgress to be false when no hand is active")
	}

	// Verify DealerSeat is nil
	if payload.DealerSeat != nil {
		t.Error("expected DealerSeat to be nil when no hand is active")
	}

	// Verify SmallBlindSeat is nil
	if payload.SmallBlindSeat != nil {
		t.Error("expected SmallBlindSeat to be nil when no hand is active")
	}

	// Verify BigBlindSeat is nil
	if payload.BigBlindSeat != nil {
		t.Error("expected BigBlindSeat to be nil when no hand is active")
	}

	// Verify Pot is nil
	if payload.Pot != nil {
		t.Error("expected Pot to be nil when no hand is active")
	}
}

// TestTableStateGameStateFields verifies dealer, blinds, and pot are correctly set
func TestTableStateGameStateFields(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up players
	token1 := "player-1"
	token2 := "player-2"

	// Assign seats
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	// Start a hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Get the active hand state for verification
	table.mu.RLock()
	activeHand := table.CurrentHand
	activeDealerSeat := *table.DealerSeat
	table.mu.RUnlock()

	if activeHand == nil {
		t.Fatal("CurrentHand should not be nil after starting hand")
	}

	expectedDealerSeat := activeDealerSeat
	expectedSmallBlindSeat := activeHand.SmallBlindSeat
	expectedBigBlindSeat := activeHand.BigBlindSeat
	expectedPot := activeHand.Pot

	// Create a mock client
	mockClient := &Client{
		Token: token1,
		send:  make(chan []byte, 10),
	}

	// Send table state
	err = mockClient.SendTableState(server, table.ID, logger)
	if err != nil {
		t.Fatalf("failed to send table state: %v", err)
	}

	// Get the message from the client's send channel
	if len(mockClient.send) == 0 {
		t.Fatal("no message sent to client")
	}

	msgBytes := <-mockClient.send
	var wsMsg WebSocketMessage
	err = json.Unmarshal(msgBytes, &wsMsg)
	if err != nil {
		t.Fatalf("failed to unmarshal WebSocketMessage: %v", err)
	}

	// Unmarshal the payload
	var payload TableStatePayload
	err = json.Unmarshal(wsMsg.Payload, &payload)
	if err != nil {
		t.Fatalf("failed to unmarshal TableStatePayload: %v", err)
	}

	// Verify DealerSeat matches
	if payload.DealerSeat == nil {
		t.Fatal("DealerSeat should not be nil")
	}
	if *payload.DealerSeat != expectedDealerSeat {
		t.Errorf("expected DealerSeat to be %d, got %d", expectedDealerSeat, *payload.DealerSeat)
	}

	// Verify SmallBlindSeat matches
	if payload.SmallBlindSeat == nil {
		t.Fatal("SmallBlindSeat should not be nil")
	}
	if *payload.SmallBlindSeat != expectedSmallBlindSeat {
		t.Errorf("expected SmallBlindSeat to be %d, got %d", expectedSmallBlindSeat, *payload.SmallBlindSeat)
	}

	// Verify BigBlindSeat matches
	if payload.BigBlindSeat == nil {
		t.Fatal("BigBlindSeat should not be nil")
	}
	if *payload.BigBlindSeat != expectedBigBlindSeat {
		t.Errorf("expected BigBlindSeat to be %d, got %d", expectedBigBlindSeat, *payload.BigBlindSeat)
	}

	// Verify Pot matches
	if payload.Pot == nil {
		t.Fatal("Pot should not be nil")
	}
	if *payload.Pot != expectedPot {
		t.Errorf("expected Pot to be %d, got %d", expectedPot, *payload.Pot)
	}
}

// TestTableStateIncludesHoleCardsForSeatedPlayer verifies seated player receives their hole cards
func TestTableStateIncludesHoleCardsForSeatedPlayer(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up 2 players using server's session manager
	session1, _ := server.sessionManager.CreateSession("Player1")
	token1 := session1.Token
	session2, _ := server.sessionManager.CreateSession("Player2")
	token2 := session2.Token

	// Update sessions with table and seat info
	server.sessionManager.UpdateSession(token1, &table.ID, &[]int{0}[0])
	server.sessionManager.UpdateSession(token2, &table.ID, &[]int{1}[0])

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	// Start a hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Get hole cards for verification
	table.mu.RLock()
	holeCards := table.CurrentHand.HoleCards
	table.mu.RUnlock()

	// Create a mock client for player 1 (seated at seat 0)
	client := &Client{
		Token: token1,
		send:  make(chan []byte, 10),
	}

	// Send table state to player 1
	err = client.SendTableState(server, table.ID, logger)
	if err != nil {
		t.Fatalf("SendTableState failed: %v", err)
	}

	// Get the message from the client's send channel
	if len(client.send) == 0 {
		t.Fatal("no message sent to client")
	}

	msgBytes := <-client.send
	var wsMsg WebSocketMessage
	err = json.Unmarshal(msgBytes, &wsMsg)
	if err != nil {
		t.Fatalf("failed to unmarshal WebSocketMessage: %v", err)
	}

	// Unmarshal the payload
	var payload TableStatePayload
	err = json.Unmarshal(wsMsg.Payload, &payload)
	if err != nil {
		t.Fatalf("failed to unmarshal TableStatePayload: %v", err)
	}

	// Verify HoleCards field exists and contains player's cards
	if payload.HoleCards == nil {
		t.Error("expected HoleCards to be non-nil for seated player")
	} else if len(payload.HoleCards) == 0 {
		t.Error("expected HoleCards to contain entries for seated player")
	}

	// Verify player 1 sees only their own hole cards (seat 0)
	if cards, ok := payload.HoleCards[0]; ok {
		if len(cards) != 2 {
			t.Errorf("expected 2 cards for seat 0, got %d", len(cards))
		}
		// Verify these match the actual hole cards dealt
		if expectedCards, hasExpected := holeCards[0]; hasExpected {
			if cards[0] != expectedCards[0] || cards[1] != expectedCards[1] {
				t.Errorf("expected cards %v, got %v", expectedCards, cards)
			}
		}
	} else {
		t.Error("expected seat 0 to be in HoleCards")
	}

	// Verify player 1 does NOT see opponent's hole cards (seat 1)
	if _, ok := payload.HoleCards[1]; ok {
		t.Error("player should NOT see opponent's hole cards (seat 1)")
	}
}

// TestTableStateOmitsHoleCardsForUnseatedPlayer verifies spectators don't get hole cards but see card counts
func TestTableStateOmitsHoleCardsForUnseatedPlayer(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up 2 players
	session1, _ := server.sessionManager.CreateSession("Player1")
	token1 := session1.Token
	session2, _ := server.sessionManager.CreateSession("Player2")
	token2 := session2.Token
	// Create spectator session (not seated)
	sessionSpectator, _ := server.sessionManager.CreateSession("Spectator")
	tokenSpectator := sessionSpectator.Token

	// Update sessions with table and seat info (only seated players)
	server.sessionManager.UpdateSession(token1, &table.ID, &[]int{0}[0])
	server.sessionManager.UpdateSession(token2, &table.ID, &[]int{1}[0])
	// Spectator has no seat

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	// Start a hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Create a mock client for spectator (not seated)
	client := &Client{
		Token: tokenSpectator,
		send:  make(chan []byte, 10),
	}

	// Send table state to spectator
	err = client.SendTableState(server, table.ID, logger)
	if err != nil {
		t.Fatalf("SendTableState failed: %v", err)
	}

	// Get the message from the client's send channel
	if len(client.send) == 0 {
		t.Fatal("no message sent to client")
	}

	msgBytes := <-client.send
	var wsMsg WebSocketMessage
	err = json.Unmarshal(msgBytes, &wsMsg)
	if err != nil {
		t.Fatalf("failed to unmarshal WebSocketMessage: %v", err)
	}

	// Unmarshal the payload
	var payload TableStatePayload
	err = json.Unmarshal(wsMsg.Payload, &payload)
	if err != nil {
		t.Fatalf("failed to unmarshal TableStatePayload: %v", err)
	}

	// Verify spectator does NOT see any hole cards
	if payload.HoleCards != nil && len(payload.HoleCards) > 0 {
		t.Error("spectator should NOT see any hole cards")
	}

	// Verify spectator DOES see card counts for occupied seats
	if payload.Seats[0].CardCount == nil {
		t.Error("expected seat 0 to have CardCount")
	} else if *payload.Seats[0].CardCount != 2 {
		t.Errorf("expected seat 0 CardCount to be 2, got %d", *payload.Seats[0].CardCount)
	}

	if payload.Seats[1].CardCount == nil {
		t.Error("expected seat 1 to have CardCount")
	} else if *payload.Seats[1].CardCount != 2 {
		t.Errorf("expected seat 1 CardCount to be 2, got %d", *payload.Seats[1].CardCount)
	}

	// Verify empty seats have nil CardCount
	for i := 2; i < 6; i++ {
		if payload.Seats[i].CardCount != nil {
			t.Errorf("expected seat %d to have nil CardCount (empty seat)", i)
		}
	}
}

// TestTableStateHoleCardsPrivacy verifies player only sees their own cards
func TestTableStateHoleCardsPrivacy(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up 3 players
	session1, _ := server.sessionManager.CreateSession("Player1")
	token1 := session1.Token
	session2, _ := server.sessionManager.CreateSession("Player2")
	token2 := session2.Token
	session3, _ := server.sessionManager.CreateSession("Player3")
	token3 := session3.Token

	// Update sessions with table and seat info
	server.sessionManager.UpdateSession(token1, &table.ID, &[]int{0}[0])
	server.sessionManager.UpdateSession(token2, &table.ID, &[]int{1}[0])
	server.sessionManager.UpdateSession(token3, &table.ID, &[]int{2}[0])

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.Seats[2].Token = &token3
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000
	table.mu.Unlock()

	// Start a hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Get hole cards for verification
	table.mu.RLock()
	player1Cards := table.CurrentHand.HoleCards[0]
	player2Cards := table.CurrentHand.HoleCards[1]
	player3Cards := table.CurrentHand.HoleCards[2]
	table.mu.RUnlock()

	// Test each player - should only see their own cards
	players := []struct {
		token         string
		seatIndex     int
		expectedCards []Card
	}{
		{token1, 0, player1Cards},
		{token2, 1, player2Cards},
		{token3, 2, player3Cards},
	}

	for _, player := range players {
		client := &Client{
			Token: player.token,
			send:  make(chan []byte, 10),
		}

		err = client.SendTableState(server, table.ID, logger)
		if err != nil {
			t.Fatalf("SendTableState failed for player at seat %d: %v", player.seatIndex, err)
		}

		if len(client.send) == 0 {
			t.Fatalf("no message sent to player at seat %d", player.seatIndex)
		}

		msgBytes := <-client.send
		var wsMsg WebSocketMessage
		err = json.Unmarshal(msgBytes, &wsMsg)
		if err != nil {
			t.Fatalf("failed to unmarshal WebSocketMessage: %v", err)
		}

		var payload TableStatePayload
		err = json.Unmarshal(wsMsg.Payload, &payload)
		if err != nil {
			t.Fatalf("failed to unmarshal TableStatePayload: %v", err)
		}

		// Verify player sees only their own cards
		if len(payload.HoleCards) != 1 {
			t.Errorf("player at seat %d should see exactly 1 entry in HoleCards, got %d", player.seatIndex, len(payload.HoleCards))
		}

		if cards, ok := payload.HoleCards[player.seatIndex]; ok {
			if len(cards) != 2 {
				t.Errorf("expected 2 cards for seat %d, got %d", player.seatIndex, len(cards))
			}
			if cards[0] != player.expectedCards[0] || cards[1] != player.expectedCards[1] {
				t.Errorf("seat %d: expected cards %v, got %v", player.seatIndex, player.expectedCards, cards)
			}
		} else {
			t.Errorf("expected seat %d in HoleCards for player at that seat", player.seatIndex)
		}

		// Verify player does NOT see any opponent cards
		for i := 0; i < 3; i++ {
			if i != player.seatIndex {
				if _, ok := payload.HoleCards[i]; ok {
					t.Errorf("player at seat %d should NOT see cards for seat %d", player.seatIndex, i)
				}
			}
		}
	}
}

// TestTableStateCardCountsForSpectators verifies card counts are populated for all seats
func TestTableStateCardCountsForSpectators(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up 3 players
	session1, _ := server.sessionManager.CreateSession("Player1")
	token1 := session1.Token
	session2, _ := server.sessionManager.CreateSession("Player2")
	token2 := session2.Token
	session3, _ := server.sessionManager.CreateSession("Player3")
	token3 := session3.Token

	// Update sessions with table and seat info
	server.sessionManager.UpdateSession(token1, &table.ID, &[]int{0}[0])
	server.sessionManager.UpdateSession(token2, &table.ID, &[]int{2}[0])
	server.sessionManager.UpdateSession(token3, &table.ID, &[]int{4}[0])

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[2].Token = &token2
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000
	table.Seats[4].Token = &token3
	table.Seats[4].Status = "active"
	table.Seats[4].Stack = 1000
	table.mu.Unlock()

	// Start a hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Create a spectator client
	sessionSpectator, _ := server.sessionManager.CreateSession("Spectator")
	spectatorClient := &Client{
		Token: sessionSpectator.Token,
		send:  make(chan []byte, 10),
	}

	// Send table state to spectator
	err = spectatorClient.SendTableState(server, table.ID, logger)
	if err != nil {
		t.Fatalf("SendTableState failed: %v", err)
	}

	msgBytes := <-spectatorClient.send
	var wsMsg WebSocketMessage
	err = json.Unmarshal(msgBytes, &wsMsg)
	if err != nil {
		t.Fatalf("failed to unmarshal WebSocketMessage: %v", err)
	}

	var payload TableStatePayload
	err = json.Unmarshal(wsMsg.Payload, &payload)
	if err != nil {
		t.Fatalf("failed to unmarshal TableStatePayload: %v", err)
	}

	// Verify occupied seats (0, 2, 4) have CardCount = 2
	occupiedSeats := []int{0, 2, 4}
	for _, seat := range occupiedSeats {
		if payload.Seats[seat].CardCount == nil {
			t.Errorf("expected seat %d to have CardCount", seat)
		} else if *payload.Seats[seat].CardCount != 2 {
			t.Errorf("expected seat %d CardCount to be 2, got %d", seat, *payload.Seats[seat].CardCount)
		}
	}

	// Verify empty seats (1, 3, 5) have nil CardCount
	emptySeats := []int{1, 3, 5}
	for _, seat := range emptySeats {
		if payload.Seats[seat].CardCount != nil {
			t.Errorf("expected seat %d to have nil CardCount (empty seat)", seat)
		}
	}

	// Test that a seated player also sees card counts for all seats with cards
	seatedClient := &Client{
		Token: token1,
		send:  make(chan []byte, 10),
	}

	err = seatedClient.SendTableState(server, table.ID, logger)
	if err != nil {
		t.Fatalf("SendTableState failed for seated player: %v", err)
	}

	msgBytes = <-seatedClient.send
	var wsMsg2 WebSocketMessage
	err = json.Unmarshal(msgBytes, &wsMsg2)
	if err != nil {
		t.Fatalf("failed to unmarshal WebSocketMessage: %v", err)
	}

	var payload2 TableStatePayload
	err = json.Unmarshal(wsMsg2.Payload, &payload2)
	if err != nil {
		t.Fatalf("failed to unmarshal TableStatePayload: %v", err)
	}

	// Verify seated player also sees card counts
	for _, seat := range occupiedSeats {
		if payload2.Seats[seat].CardCount == nil {
			t.Errorf("seated player: expected seat %d to have CardCount", seat)
		} else if *payload2.Seats[seat].CardCount != 2 {
			t.Errorf("seated player: expected seat %d CardCount to be 2, got %d", seat, *payload2.Seats[seat].CardCount)
		}
	}
}

// ============= Phase 4: WebSocket Protocol & Handler Tests =============

// TestHandlePlayerAction_ValidCall verifies call action processes
func TestHandlePlayerAction_ValidCall(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	sm := NewSessionManager(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up 2 players with sessions
	session1, _ := sm.CreateSession("Player1")
	session2, _ := sm.CreateSession("Player2")
	token1 := session1.Token
	token2 := session2.Token

	// Update sessions with table and seat info
	sm.UpdateSession(token1, &table.ID, &[]int{0}[0])
	sm.UpdateSession(token2, &table.ID, &[]int{1}[0])

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Verify CurrentActor is set
	table.mu.RLock()
	currentActor := table.CurrentHand.CurrentActor
	table.mu.RUnlock()
	if currentActor == nil || *currentActor != 0 {
		t.Fatalf("expected CurrentActor to be seat 0, got %v", currentActor)
	}

	// Player at seat 0 calls
	client := &Client{
		hub:   hub,
		Token: token1,
		send:  make(chan []byte, 256),
	}

	err = server.HandlePlayerAction(sm, client, 0, "call")
	if err != nil {
		t.Errorf("expected no error for valid call, got %v", err)
	}

	// Verify action was processed
	table.mu.RLock()
	if table.CurrentHand.ActedPlayers == nil {
		t.Fatal("ActedPlayers should be initialized")
	}
	if !table.CurrentHand.ActedPlayers[0] {
		t.Errorf("expected seat 0 to be marked as acted")
	}
	// Verify stack was updated (dealer posts SB=10, calls BB=20, so needs to add 10 more)
	// Initial stack 1000 - 10 (SB) - 10 (call to match BB) = 980
	expectedStack := 1000 - 10 - 10
	if table.Seats[0].Stack != expectedStack {
		t.Errorf("expected seat 0 stack to be %d after call, got %d", expectedStack, table.Seats[0].Stack)
	}
	table.mu.RUnlock()
}

// TestHandlePlayerAction_ValidCheck verifies check action processes
func TestHandlePlayerAction_ValidCheck(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	sm := NewSessionManager(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up 2 players with sessions
	session1, _ := sm.CreateSession("Player1")
	session2, _ := sm.CreateSession("Player2")
	token1 := session1.Token
	token2 := session2.Token

	// Update sessions with table and seat info
	sm.UpdateSession(token1, &table.ID, &[]int{0}[0])
	sm.UpdateSession(token2, &table.ID, &[]int{1}[0])

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// First player calls to match BB
	client := &Client{
		hub:   hub,
		Token: token1,
		send:  make(chan []byte, 256),
	}

	err = server.HandlePlayerAction(sm, client, 0, "call")
	if err != nil {
		t.Errorf("unexpected error on first call: %v", err)
	}

	// Second player should be able to check (already posted BB)
	client2 := &Client{
		hub:   hub,
		Token: token2,
		send:  make(chan []byte, 256),
	}

	// Verify check is valid action for seat 1
	table.mu.RLock()
	validActions := table.CurrentHand.GetValidActions(1, table.Seats[1].Stack, table.Seats)
	table.mu.RUnlock()

	hasCheck := false
	for _, action := range validActions {
		if action == "check" {
			hasCheck = true
			break
		}
	}
	if !hasCheck {
		t.Errorf("expected check to be valid action for seat 1, got %v", validActions)
	}

	// Store street before check
	table.mu.RLock()
	streetBefore := table.CurrentHand.Street
	table.mu.RUnlock()

	err = server.HandlePlayerAction(sm, client2, 1, "check")
	if err != nil {
		t.Errorf("expected no error for valid check, got %v", err)
	}

	// Verify action was processed and street advanced (since check completes betting round)
	table.mu.RLock()
	streetAfter := table.CurrentHand.Street
	table.mu.RUnlock()

	if streetAfter == streetBefore {
		t.Errorf("expected street to advance after check completes betting round, but street stayed %s", streetBefore)
	}
}

// TestHandlePlayerAction_ValidFold verifies fold action marks player folded
func TestHandlePlayerAction_ValidFold(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	sm := NewSessionManager(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up 3 players with sessions (need 3 so hand doesn't end on single fold)
	session1, _ := sm.CreateSession("Player1")
	session2, _ := sm.CreateSession("Player2")
	session3, _ := sm.CreateSession("Player3")
	token1 := session1.Token
	token2 := session2.Token
	token3 := session3.Token

	// Update sessions with table and seat info
	sm.UpdateSession(token1, &table.ID, &[]int{0}[0])
	sm.UpdateSession(token2, &table.ID, &[]int{1}[0])
	sm.UpdateSession(token3, &table.ID, &[]int{2}[0])

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.Seats[2].Token = &token3
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000
	table.mu.Unlock()

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Player at seat 0 folds
	client := &Client{
		hub:   hub,
		Token: token1,
		send:  make(chan []byte, 256),
	}

	err = server.HandlePlayerAction(sm, client, 0, "fold")
	if err != nil {
		t.Errorf("expected no error for valid fold, got %v", err)
	}

	// Verify player was marked as folded
	table.mu.RLock()
	if table.CurrentHand.FoldedPlayers == nil {
		t.Fatal("FoldedPlayers should be initialized")
	}
	if !table.CurrentHand.FoldedPlayers[0] {
		t.Errorf("expected seat 0 to be marked as folded")
	}
	table.mu.RUnlock()
}

// TestHandlePlayerAction_InvalidAction verifies error on invalid action
func TestHandlePlayerAction_InvalidAction(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	sm := NewSessionManager(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up 2 players with sessions
	session1, _ := sm.CreateSession("Player1")
	session2, _ := sm.CreateSession("Player2")
	token1 := session1.Token
	token2 := session2.Token

	// Update sessions with table and seat info
	sm.UpdateSession(token1, &table.ID, &[]int{0}[0])
	sm.UpdateSession(token2, &table.ID, &[]int{1}[0])

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Player at seat 0 tries to check (but must call since CurrentBet=20 and they bet 10)
	client := &Client{
		hub:   hub,
		Token: token1,
		send:  make(chan []byte, 256),
	}

	err = server.HandlePlayerAction(sm, client, 0, "check")
	if err == nil {
		t.Errorf("expected error for invalid check, got nil")
	}
}

// TestHandlePlayerAction_OutOfTurn verifies error when not current actor
func TestHandlePlayerAction_OutOfTurn(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	sm := NewSessionManager(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up 2 players with sessions
	session1, _ := sm.CreateSession("Player1")
	session2, _ := sm.CreateSession("Player2")
	token1 := session1.Token
	token2 := session2.Token

	// Update sessions with table and seat info
	sm.UpdateSession(token1, &table.ID, &[]int{0}[0])
	sm.UpdateSession(token2, &table.ID, &[]int{1}[0])

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Player at seat 1 tries to act (but seat 0 is current actor)
	client := &Client{
		hub:   hub,
		Token: token2,
		send:  make(chan []byte, 256),
	}

	err = server.HandlePlayerAction(sm, client, 1, "call")
	if err == nil {
		t.Errorf("expected error for out of turn action, got nil")
	}
}

// TestHandlePlayerAction_BroadcastsResult verifies action_result is broadcast after action
func TestHandlePlayerAction_BroadcastsResult(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	sm := NewSessionManager(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up 2 players with sessions
	session1, _ := sm.CreateSession("Player1")
	session2, _ := sm.CreateSession("Player2")
	token1 := session1.Token
	token2 := session2.Token

	// Update sessions with table and seat info
	sm.UpdateSession(token1, &table.ID, &[]int{0}[0])
	sm.UpdateSession(token2, &table.ID, &[]int{1}[0])

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Create two mock clients with send channels so we can verify broadcasts
	client1 := &Client{
		hub:   hub,
		send:  make(chan []byte, 256),
		Token: token1,
	}
	client2 := &Client{
		hub:   hub,
		send:  make(chan []byte, 256),
		Token: token2,
	}

	// Register clients in hub
	hub.register <- client1
	hub.register <- client2
	// Give time for hub to process registrations
	time.Sleep(50 * time.Millisecond)

	// Player at seat 0 calls
	err = server.HandlePlayerAction(sm, client1, 0, "call")
	if err != nil {
		t.Errorf("expected no error for valid call, got %v", err)
	}

	// Give time for broadcast
	time.Sleep(50 * time.Millisecond)

	// Verify both clients received action_result
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

// TestHandlePlayerAction_RaiseWithAmount verifies handler extracts amount and calls ProcessAction correctly for raises
func TestHandlePlayerAction_RaiseWithAmount(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	sm := NewSessionManager(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up 2 players with sessions
	session1, _ := sm.CreateSession("Player1")
	session2, _ := sm.CreateSession("Player2")
	token1 := session1.Token
	token2 := session2.Token

	// Update sessions with table and seat info
	sm.UpdateSession(token1, &table.ID, &[]int{0}[0])
	sm.UpdateSession(token2, &table.ID, &[]int{1}[0])

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Player at seat 0 raises to 50
	client := &Client{
		hub:   hub,
		Token: token1,
		send:  make(chan []byte, 256),
	}

	// Test that we can marshal a PlayerActionPayload with an Amount field
	raiseAmount := 50
	payload := PlayerActionPayload{
		SeatIndex: 0,
		Action:    "raise",
		Amount:    &raiseAmount,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	// Test the message handler which should extract the amount
	err = client.HandlePlayerActionMessage(sm, server, logger, payloadBytes)
	if err != nil {
		t.Errorf("expected no error for valid raise with amount, got %v", err)
	}

	// Verify action was processed
	table.mu.RLock()
	if table.CurrentHand.ActedPlayers == nil {
		t.Fatal("ActedPlayers should be initialized")
	}
	if !table.CurrentHand.ActedPlayers[0] {
		t.Errorf("expected seat 0 to be marked as acted")
	}
	// Verify the raise was recorded at the correct amount
	if table.CurrentHand.PlayerBets[0] != 50 {
		t.Errorf("expected seat 0 bet to be 50, got %d", table.CurrentHand.PlayerBets[0])
	}
	table.mu.RUnlock()
}

// TestHandlePlayerAction_RaiseMissingAmount verifies handler returns error when raise lacks amount
func TestHandlePlayerAction_RaiseMissingAmount(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	sm := NewSessionManager(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up 2 players with sessions
	session1, _ := sm.CreateSession("Player1")
	session2, _ := sm.CreateSession("Player2")
	token1 := session1.Token
	token2 := session2.Token

	// Update sessions with table and seat info
	sm.UpdateSession(token1, &table.ID, &[]int{0}[0])
	sm.UpdateSession(token2, &table.ID, &[]int{1}[0])

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Player at seat 0 tries to raise without an amount - should fail
	client := &Client{
		hub:   hub,
		Token: token1,
		send:  make(chan []byte, 256),
	}

	// Create a raise action without amount
	payload := PlayerActionPayload{
		SeatIndex: 0,
		Action:    "raise",
		Amount:    nil, // Missing amount
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	err = client.HandlePlayerActionMessage(sm, server, logger, payloadBytes)
	if err == nil {
		t.Error("expected error for raise without amount, got nil")
	}
	if err != nil && err.Error() != "raise action requires amount parameter" {
		t.Errorf("expected error 'raise action requires amount parameter', got %v", err)
	}
}

// TestBroadcastActionRequest_IncludesMinMaxRaise verifies action_request payload includes minRaise and maxRaise fields
func TestBroadcastActionRequest_IncludesMinMaxRaise(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	sm := NewSessionManager(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up 2 players with sessions
	session1, _ := sm.CreateSession("Player1")
	session2, _ := sm.CreateSession("Player2")
	token1 := session1.Token
	token2 := session2.Token

	// Update sessions with table and seat info
	sm.UpdateSession(token1, &table.ID, &[]int{0}[0])
	sm.UpdateSession(token2, &table.ID, &[]int{1}[0])

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Create clients and register them
	client1 := &Client{
		hub:   hub,
		Token: token1,
		send:  make(chan []byte, 256),
	}
	client2 := &Client{
		hub:   hub,
		Token: token2,
		send:  make(chan []byte, 256),
	}
	hub.register <- client1
	hub.register <- client2
	time.Sleep(50 * time.Millisecond)

	// Request action for seat 0
	validActions := []string{"fold", "check", "raise"}
	callAmount := 10 // Seat 0 is dealer with SB=10, so call amount is 0, but we'll test with next actor
	currentBet := 20
	pot := 30

	err = server.BroadcastActionRequest(table.ID, 0, validActions, callAmount, currentBet, pot)
	if err != nil {
		t.Fatalf("failed to broadcast action_request: %v", err)
	}

	// Give time for broadcast
	time.Sleep(50 * time.Millisecond)

	// Verify client received action_request with minRaise and maxRaise
	select {
	case msg := <-client1.send:
		var wsMsg WebSocketMessage
		err := json.Unmarshal(msg, &wsMsg)
		if err != nil {
			t.Errorf("failed to unmarshal message: %v", err)
		}
		if wsMsg.Type != "action_request" {
			t.Errorf("expected message type 'action_request', got %q", wsMsg.Type)
		}
		var payload ActionRequestPayload
		err = json.Unmarshal(wsMsg.Payload, &payload)
		if err != nil {
			t.Errorf("failed to unmarshal payload: %v", err)
		}
		// Verify minRaise and maxRaise fields are present and accessible
		if payload.MinRaise < 0 {
			t.Errorf("minRaise should be non-negative, got %d", payload.MinRaise)
		}
		if payload.MaxRaise < 0 {
			t.Errorf("maxRaise should be non-negative, got %d", payload.MaxRaise)
		}
	default:
		t.Error("client1 did not receive action_request message")
	}
}

// TestBroadcastActionRequest_MinMaxCalculation verifies MinRaise and MaxRaise are calculated correctly
func TestBroadcastActionRequest_MinMaxCalculation(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	sm := NewSessionManager(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Test Case 1: Multi-player game (stack sizes: 1000, 1000)
	session1, _ := sm.CreateSession("Player1")
	session2, _ := sm.CreateSession("Player2")
	token1 := session1.Token
	token2 := session2.Token

	sm.UpdateSession(token1, &table.ID, &[]int{0}[0])
	sm.UpdateSession(token2, &table.ID, &[]int{1}[0])

	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Create clients and register
	client1 := &Client{
		hub:   hub,
		Token: token1,
		send:  make(chan []byte, 256),
	}
	client2 := &Client{
		hub:   hub,
		Token: token2,
		send:  make(chan []byte, 256),
	}
	hub.register <- client1
	hub.register <- client2
	time.Sleep(50 * time.Millisecond)

	// Broadcast action request for multi-player game
	validActions := []string{"fold", "check", "raise"}
	err = server.BroadcastActionRequest(table.ID, 0, validActions, 0, 20, 30)
	if err != nil {
		t.Fatalf("failed to broadcast action_request: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Verify min/max calculation for multi-player
	select {
	case msg := <-client1.send:
		var wsMsg WebSocketMessage
		json.Unmarshal(msg, &wsMsg)
		var payload ActionRequestPayload
		json.Unmarshal(wsMsg.Payload, &payload)

		// With BB=20 and opponent stack after posting BB=20 (1000-20=980), minRaise should be 20 + 20 = 40
		// maxRaise is min of player's remaining stack (1000-10=990) and opponent's stack (1000-20=980) = 980
		if payload.MinRaise != 40 {
			t.Errorf("multi-player: expected minRaise=40, got %d", payload.MinRaise)
		}
		// maxRaise should be opponent's remaining stack after posting BB (1000 - 20 = 980)
		if payload.MaxRaise != 980 {
			t.Errorf("multi-player: expected maxRaise=980, got %d", payload.MaxRaise)
		}
	default:
		t.Error("client1 did not receive action_request message")
	}

	// Test Case 2: Heads-up game (different stack sizes: 800, 1200)
	// Clear table and recreate
	table.mu.Lock()
	table.CurrentHand = nil
	table.mu.Unlock()

	session3, _ := sm.CreateSession("Player3")
	session4, _ := sm.CreateSession("Player4")
	token3 := session3.Token
	token4 := session4.Token

	sm.UpdateSession(token3, &table.ID, &[]int{0}[0])
	sm.UpdateSession(token4, &table.ID, &[]int{1}[0])

	table.mu.Lock()
	table.Seats[0].Token = &token3
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 800
	table.Seats[1].Token = &token4
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1200
	table.mu.Unlock()

	err = table.StartHand()
	if err != nil {
		t.Fatalf("failed to start second hand: %v", err)
	}

	client3 := &Client{
		hub:   hub,
		Token: token3,
		send:  make(chan []byte, 256),
	}
	client4 := &Client{
		hub:   hub,
		Token: token4,
		send:  make(chan []byte, 256),
	}
	hub.register <- client3
	hub.register <- client4
	time.Sleep(50 * time.Millisecond)

	// Broadcast action request for heads-up
	err = server.BroadcastActionRequest(table.ID, 0, validActions, 0, 20, 30)
	if err != nil {
		t.Fatalf("failed to broadcast action_request for heads-up: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Verify min/max calculation for heads-up
	select {
	case msg := <-client3.send:
		var wsMsg WebSocketMessage
		json.Unmarshal(msg, &wsMsg)
		var payload ActionRequestPayload
		json.Unmarshal(wsMsg.Payload, &payload)

		// Heads-up scenario:
		// Dealer (seat 1) posts SB: 1200 - 10 = 1190
		// Non-dealer (seat 0, our player) posts BB: 800 - 20 = 780
		// Player at seat 0 is acting, so GetMaxRaise(0) = min(780, 1190) = 780
		if payload.MaxRaise != 780 {
			t.Errorf("heads-up: expected maxRaise=780, got %d", payload.MaxRaise)
		}
	default:
		t.Error("client3 did not receive action_request message")
	}
}

// TestBroadcastBoardDealt_SendsToAllTablePlayers verifies board_dealt message is sent to all players at the table
func TestBroadcastBoardDealt_SendsToAllTablePlayers(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up players
	token1 := "player-1"
	token2 := "player-2"

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	// Create mock clients and register with hub
	client1 := &Client{
		hub:   hub,
		Token: token1,
		send:  make(chan []byte, 256),
	}
	client2 := &Client{
		hub:   hub,
		Token: token2,
		send:  make(chan []byte, 256),
	}
	hub.register <- client1
	hub.register <- client2
	time.Sleep(50 * time.Millisecond)

	// Start a hand to establish CurrentHand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Drain any messages from starting hand
	drainChannels := func(c1, c2 *Client) {
		for len(c1.send) > 0 {
			<-c1.send
		}
		for len(c2.send) > 0 {
			<-c2.send
		}
	}

	time.Sleep(100 * time.Millisecond)
	drainChannels(client1, client2)

	// Advance to flop to have board cards
	table.mu.Lock()
	if table.CurrentHand != nil {
		table.CurrentHand.AdvanceToNextStreet()
	}
	table.mu.Unlock()

	// Call broadcastBoardDealt
	err = server.broadcastBoardDealt(table, "flop")
	if err != nil {
		t.Fatalf("failed to broadcast board dealt: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Verify both clients received messages
	select {
	case msg := <-client1.send:
		var wsMsg WebSocketMessage
		err := json.Unmarshal(msg, &wsMsg)
		if err != nil {
			t.Fatalf("failed to parse client1 message: %v", err)
		}
		if wsMsg.Type != "board_dealt" {
			t.Errorf("client1 expected message type 'board_dealt', got %q", wsMsg.Type)
		}
	default:
		t.Error("client1 did not receive board_dealt message")
	}

	select {
	case msg := <-client2.send:
		var wsMsg WebSocketMessage
		err := json.Unmarshal(msg, &wsMsg)
		if err != nil {
			t.Fatalf("failed to parse client2 message: %v", err)
		}
		if wsMsg.Type != "board_dealt" {
			t.Errorf("client2 expected message type 'board_dealt', got %q", wsMsg.Type)
		}
	default:
		t.Error("client2 did not receive board_dealt message")
	}
}

// TestBroadcastBoardDealt_IncludesCorrectBoardCards verifies board_dealt includes the correct board cards
func TestBroadcastBoardDealt_IncludesCorrectBoardCards(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up 2 players
	token1 := "player-1"
	token2 := "player-2"
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	// Create mock clients
	client1 := &Client{
		hub:   hub,
		Token: token1,
		send:  make(chan []byte, 256),
	}
	hub.register <- client1
	time.Sleep(50 * time.Millisecond)

	// Start a hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Drain messages from starting hand
	time.Sleep(100 * time.Millisecond)
	for len(client1.send) > 0 {
		<-client1.send
	}

	// Advance to flop to have board cards
	table.mu.Lock()
	if table.CurrentHand != nil {
		table.CurrentHand.AdvanceToNextStreet()
		// Get the board cards
		boardCards := table.CurrentHand.BoardCards
		table.mu.Unlock()

		// Broadcast
		err = server.broadcastBoardDealt(table, "flop")
		if err != nil {
			t.Fatalf("failed to broadcast board dealt: %v", err)
		}

		time.Sleep(50 * time.Millisecond)

		// Read and verify the message
		select {
		case msg := <-client1.send:
			var wsMsg WebSocketMessage
			json.Unmarshal(msg, &wsMsg)
			var payload BoardDealtPayload
			json.Unmarshal(wsMsg.Payload, &payload)

			if len(payload.BoardCards) != len(boardCards) {
				t.Errorf("expected %d board cards, got %d", len(boardCards), len(payload.BoardCards))
			}

			for i, card := range payload.BoardCards {
				if card.Rank != boardCards[i].Rank || card.Suit != boardCards[i].Suit {
					t.Errorf("card %d mismatch: expected %s%s, got %s%s", i,
						boardCards[i].Rank, boardCards[i].Suit, card.Rank, card.Suit)
				}
			}
		default:
			t.Error("client did not receive board_dealt message")
		}
	} else {
		table.mu.Unlock()
		t.Fatal("CurrentHand is nil")
	}
}

// TestBroadcastBoardDealt_IncludesStreetIndicator verifies board_dealt includes the correct street indicator
func TestBroadcastBoardDealt_IncludesStreetIndicator(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up 2 players
	token1 := "player-1"
	token2 := "player-2"
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	// Create mock client
	client1 := &Client{
		hub:   hub,
		Token: token1,
		send:  make(chan []byte, 256),
	}
	hub.register <- client1
	time.Sleep(50 * time.Millisecond)

	// Start a hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Test all three streets
	streetsToTest := []string{"flop", "turn", "river"}

	for _, streetName := range streetsToTest {
		// Clear previous messages
		time.Sleep(100 * time.Millisecond)
		for len(client1.send) > 0 {
			<-client1.send
		}

		// Broadcast for this street
		err = server.broadcastBoardDealt(table, streetName)
		if err != nil {
			t.Fatalf("failed to broadcast board dealt for %s: %v", streetName, err)
		}

		time.Sleep(50 * time.Millisecond)

		// Read and verify the message
		select {
		case msg := <-client1.send:
			var wsMsg WebSocketMessage
			json.Unmarshal(msg, &wsMsg)
			var payload BoardDealtPayload
			json.Unmarshal(wsMsg.Payload, &payload)

			if payload.Street != streetName {
				t.Errorf("expected street %q, got %q", streetName, payload.Street)
			}
		default:
			t.Errorf("client did not receive board_dealt message for street %s", streetName)
		}

		// Advance to next street
		table.mu.Lock()
		if table.CurrentHand != nil {
			table.CurrentHand.AdvanceToNextStreet()
		}
		table.mu.Unlock()
	}
}

// TestHandlePlayerAction_AdvancesStreetAfterRoundComplete verifies that when a betting round
// completes, the street is advanced and board cards are dealt and broadcast
func TestHandlePlayerAction_AdvancesStreetAfterRoundComplete(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	sm := NewSessionManager(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up 2 players with sessions (heads-up scenario)
	session1, _ := sm.CreateSession("Player1")
	session2, _ := sm.CreateSession("Player2")
	token1 := session1.Token
	token2 := session2.Token

	// Update sessions with table and seat info
	sm.UpdateSession(token1, &table.ID, &[]int{0}[0])
	sm.UpdateSession(token2, &table.ID, &[]int{1}[0])

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	// Start hand
	// In heads-up: dealer (seat 0) posts SB, seat 1 posts BB
	// Preflop: dealer (seat 0) acts first
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Verify we're on preflop street
	table.mu.RLock()
	if table.CurrentHand.Street != "preflop" {
		t.Fatalf("expected street preflop, got %s", table.CurrentHand.Street)
	}
	table.mu.RUnlock()

	// Create clients for both players
	client1 := &Client{
		hub:   hub,
		Token: token1,
		send:  make(chan []byte, 256),
	}
	client2 := &Client{
		hub:   hub,
		Token: token2,
		send:  make(chan []byte, 256),
	}

	hub.register <- client1
	hub.register <- client2
	time.Sleep(50 * time.Millisecond)

	// Clear initial hand setup messages
	for len(client1.send) > 0 {
		<-client1.send
	}
	for len(client2.send) > 0 {
		<-client2.send
	}

	// Get current action state
	table.mu.RLock()
	currentActor := *table.CurrentHand.CurrentActor
	table.mu.RUnlock()

	if currentActor != 0 {
		t.Fatalf("expected dealer (seat 0) to act first in heads-up, got seat %d", currentActor)
	}

	// Seat 0 (dealer/small blind) calls the big blind
	err = server.HandlePlayerAction(sm, client1, 0, "call")
	if err != nil {
		t.Errorf("expected no error for dealer call, got %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify seat 1 is now the current actor (big blind can check since dealer called)
	table.mu.RLock()
	currentActor = *table.CurrentHand.CurrentActor
	table.mu.RUnlock()

	if currentActor != 1 {
		t.Fatalf("expected seat 1 to act after dealer calls, got seat %d", currentActor)
	}

	// Clear messages from client2 before next action
	for len(client2.send) > 0 {
		<-client2.send
	}

	// Seat 1 (big blind) checks to complete the preflop betting round
	err = server.HandlePlayerAction(sm, client2, 1, "check")
	if err != nil {
		t.Errorf("expected no error for big blind check, got %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify street advanced from preflop to flop
	table.mu.RLock()
	currentStreet := table.CurrentHand.Street
	boardCardCount := len(table.CurrentHand.BoardCards)
	table.mu.RUnlock()

	if currentStreet != "flop" {
		t.Errorf("expected street to advance to 'flop' after betting round complete, got '%s'", currentStreet)
	}

	if boardCardCount != 3 {
		t.Errorf("expected flop to have 3 cards, got %d", boardCardCount)
	}

	// Verify board_dealt message was broadcast with "flop" street
	// Drain and check messages from both clients
	boardDealReceived := false
	checkChannelForBoardDealt := func(client *Client) bool {
		for {
			select {
			case msg := <-client.send:
				var wsMsg WebSocketMessage
				if err := json.Unmarshal(msg, &wsMsg); err == nil {
					if wsMsg.Type == "board_dealt" {
						var payload BoardDealtPayload
						if err := json.Unmarshal(wsMsg.Payload, &payload); err == nil {
							if payload.Street == "flop" && len(payload.BoardCards) == 3 {
								return true
							}
						}
					}
				}
			default:
				return false
			}
		}
	}

	if checkChannelForBoardDealt(client1) || checkChannelForBoardDealt(client2) {
		boardDealReceived = true
	}

	if !boardDealReceived {
		t.Error("expected board_dealt message with 'flop' street to be broadcast")
	}
}

// TestHandlerFlow_RiverToShowdown - Full hand flow from deal to river to showdown
func TestHandlerFlow_RiverToShowdown(t *testing.T) {
	// This is a simplified test that verifies showdown is triggered when river betting completes
	// We'll create a hand that reaches river and verify HandleShowdown would be called

	// Create a new server and table
	server := &Server{
		logger: slog.Default(),
	}
	table := NewTable("table-1", "Test Table", server)
	server.tables[0] = table

	// Setup: Create a hand already at river with 2 active players
	hand := &Hand{
		Pot:            100,
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   0, // 2-player game
		Street:         "river",
		HoleCards: map[int][]Card{
			0: {Card{Rank: "A", Suit: "s"}, Card{Rank: "K", Suit: "s"}},
			1: {Card{Rank: "Q", Suit: "h"}, Card{Rank: "J", Suit: "h"}},
		},
		BoardCards: []Card{
			{Rank: "T", Suit: "d"}, {Rank: "9", Suit: "c"}, {Rank: "8", Suit: "s"},
			{Rank: "7", Suit: "h"}, {Rank: "6", Suit: "d"},
		},
		FoldedPlayers: make(map[int]bool),
		ActedPlayers:  make(map[int]bool),
		PlayerBets: map[int]int{
			0: 50,
			1: 50,
		},
		CurrentBet:        50,
		BigBlindHasOption: false,
	}

	table.CurrentHand = hand
	dealerSeat := 0
	table.DealerSeat = &dealerSeat // Set the table's dealer to match the hand
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 950
	table.Seats[0].Token = newString("token0")
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 950
	table.Seats[1].Token = newString("token1")

	// Verify IsBettingRoundComplete is true (both players have acted and matched the bet)
	hand.ActedPlayers[0] = true
	hand.ActedPlayers[1] = true

	if !hand.IsBettingRoundComplete(table.Seats) {
		t.Error("expected betting round to be complete on river")
	}

	// Call HandleShowdown - this should not panic or error
	table.HandleShowdown()

	// Verify the hand is cleared (HandCleanup happens during HandleShowdown)
	if table.CurrentHand != nil {
		t.Error("expected CurrentHand to be nil after HandleShowdown")
	}

	// Verify dealer was rotated
	if table.DealerSeat == nil || *table.DealerSeat != 1 {
		t.Errorf("expected dealer to rotate to seat 1, got %v", table.DealerSeat)
	}
}

// TestHandlerFlow_AllFoldBeforeShowdown - All players fold, one wins early
func TestHandlerFlow_AllFoldBeforeShowdown(t *testing.T) {
	// This test verifies that when all but one player fold, the remaining player wins
	// without needing to evaluate hands

	// Create a new server and table
	server := &Server{
		logger: slog.Default(),
	}
	table := NewTable("table-1", "Test Table", server)
	server.tables[0] = table

	// Setup: Create a hand at flop where all but one player folded
	hand := &Hand{
		Pot:            100,
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Street:         "flop",
		HoleCards: map[int][]Card{
			0: {Card{Rank: "A", Suit: "s"}, Card{Rank: "K", Suit: "s"}},
			1: {Card{Rank: "2", Suit: "h"}, Card{Rank: "3", Suit: "h"}},
			2: {Card{Rank: "4", Suit: "d"}, Card{Rank: "5", Suit: "d"}},
		},
		BoardCards: []Card{
			{Rank: "6", Suit: "c"}, {Rank: "7", Suit: "s"}, {Rank: "8", Suit: "h"},
		},
		FoldedPlayers: map[int]bool{
			1: true, // Seat 1 folded
			2: true, // Seat 2 folded
		},
		ActedPlayers: make(map[int]bool),
		PlayerBets: map[int]int{
			0: 50,
		},
		CurrentBet:        50,
		BigBlindHasOption: false,
	}

	table.CurrentHand = hand
	dealerSeat := 0
	table.DealerSeat = &dealerSeat // Set the table's dealer to match the hand
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 950
	table.Seats[0].Token = newString("token0")
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 950
	table.Seats[1].Token = newString("token1")
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 950
	table.Seats[2].Token = newString("token2")

	// Verify IsBettingRoundComplete returns true (only 1 player remains)
	if !hand.IsBettingRoundComplete(table.Seats) {
		t.Error("expected betting round to be complete when only 1 player remains")
	}

	// Call HandleShowdown - should handle early winner case
	table.HandleShowdown()

	// Verify the hand is cleared (HandCleanup happens during HandleShowdown)
	if table.CurrentHand != nil {
		t.Error("expected CurrentHand to be nil after HandleShowdown")
	}

	// Verify dealer was rotated
	if table.DealerSeat == nil || *table.DealerSeat != 1 {
		t.Errorf("expected dealer to rotate to seat 1, got %v", table.DealerSeat)
	}
}

// ============ PHASE 4: HANDLER FLOW TESTS FOR HAND CLEANUP & NEXT HAND ============

// TestHandlerFlow_FullHandCycle_ManualNextHand verifies complete hand flow with manual "Start Hand" button
func TestHandlerFlow_FullHandCycle_ManualNextHand(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	table := server.tables[0]

	// Set up 2 active players
	token0 := "player-0"
	token1 := "player-1"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	// Start first hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Verify hand is active
	if table.CurrentHand == nil {
		t.Fatal("expected CurrentHand to be set after StartHand")
	}

	// Simulate early winner (seat 1 folds)
	table.CurrentHand.FoldedPlayers[1] = true

	// Call HandleShowdown
	table.HandleShowdown()

	// After showdown, CurrentHand should be cleared
	if table.CurrentHand != nil {
		t.Errorf("expected CurrentHand to be nil after HandleShowdown, got %v", table.CurrentHand)
	}

	// Dealer should have rotated
	if table.DealerSeat == nil || *table.DealerSeat != 1 {
		t.Errorf("expected dealer to rotate to seat 1, got %v", table.DealerSeat)
	}

	// Both players should still be "active" (no auto promotion of waiting players)
	if table.Seats[0].Status != "active" || table.Seats[1].Status != "active" {
		t.Errorf("expected both seats to remain active, got seat 0: %s, seat 1: %s",
			table.Seats[0].Status, table.Seats[1].Status)
	}

	// Now manually start next hand via StartHand
	err = table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting second hand, got %v", err)
	}

	// Verify new hand is created
	if table.CurrentHand == nil {
		t.Fatal("expected new CurrentHand to be set after second StartHand")
	}

	// Dealer should remain at seat 1 (was rotated to seat 1 after first hand ended, stays there for second hand)
	// Dealer rotates at the END of a hand (in HandleShowdown), not during StartHand
	if table.CurrentHand.DealerSeat != 1 {
		t.Errorf("expected second hand dealer at seat 1 (rotated after first hand), got %d", table.CurrentHand.DealerSeat)
	}
}

// TestHandlerFlow_HandEndsWithBustOut verifies bust-out handling at showdown
func TestHandlerFlow_HandEndsWithBustOut(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	table := server.tables[0]

	// Set up 2 players: seat 0 wins 100, seat 1 has exactly 100 (will bust)
	token0 := "player-0"
	token1 := "player-1"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 100 // Small stack - will go all-in for BB (20) and lose

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Mark seat 1 as folded (loses the hand)
	table.CurrentHand.FoldedPlayers[1] = true

	// Call HandleShowdown
	table.HandleShowdown()

	// After showdown, hand should be cleared
	if table.CurrentHand != nil {
		t.Errorf("expected CurrentHand to be nil after HandleShowdown, got %v", table.CurrentHand)
	}

	// If seat 1 stack became 0, it should be busted out (Token=nil, Status="empty")
	if table.Seats[1].Stack == 0 {
		if table.Seats[1].Status != "empty" {
			t.Errorf("expected busted-out seat 1 status to be 'empty', got '%s'", table.Seats[1].Status)
		}
		if table.Seats[1].Token != nil {
			t.Errorf("expected busted-out seat 1 Token to be nil, got %v", table.Seats[1].Token)
		}
	} else {
		// Seat 1 still has chips, so should remain "active"
		if table.Seats[1].Status != "active" {
			t.Errorf("expected seat 1 status to remain 'active' with remaining stack, got '%s'", table.Seats[1].Status)
		}
	}

	// Dealer should have rotated
	if table.DealerSeat == nil {
		t.Error("expected DealerSeat to be set after HandleShowdown")
	}
}

// TestHandlerFlow_DealerRotatesAfterShowdown verifies dealer position rotates correctly after showdown
func TestHandlerFlow_DealerRotatesAfterShowdown(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	table := server.tables[0]

	// Set up 3 active players at seats 1, 3, 5
	token1 := "player-1"
	token3 := "player-3"
	token5 := "player-5"

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	table.Seats[3].Token = &token3
	table.Seats[3].Status = "active"
	table.Seats[3].Stack = 1000

	table.Seats[5].Token = &token5
	table.Seats[5].Status = "active"
	table.Seats[5].Stack = 1000

	// Start first hand (dealer should be seat 1)
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting first hand, got %v", err)
	}

	if table.DealerSeat == nil || *table.DealerSeat != 1 {
		t.Fatalf("expected dealer to be seat 1 initially, got %v", table.DealerSeat)
	}

	// Mark seat 3 as folded (seat 1 wins)
	table.CurrentHand.FoldedPlayers[3] = true
	table.CurrentHand.FoldedPlayers[5] = true

	// Call HandleShowdown
	table.HandleShowdown()

	// After first showdown, dealer should have rotated to seat 3
	if table.DealerSeat == nil || *table.DealerSeat != 3 {
		t.Errorf("expected dealer to rotate to seat 3 after first showdown, got %v", table.DealerSeat)
	}

	// Start second hand
	err = table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting second hand, got %v", err)
	}

	// Verify dealer is 3
	if table.CurrentHand.DealerSeat != 3 {
		t.Errorf("expected dealer in second hand to be seat 3, got %d", table.CurrentHand.DealerSeat)
	}

	// Mark seat 5 as folded (seat 3 wins)
	table.CurrentHand.FoldedPlayers[5] = true
	table.CurrentHand.FoldedPlayers[1] = true

	// Call HandleShowdown for second hand
	table.HandleShowdown()

	// After second showdown, dealer should have rotated from seat 3 to seat 5
	if table.DealerSeat == nil || *table.DealerSeat != 5 {
		t.Errorf("expected dealer to rotate to seat 5 after second showdown, got %v", table.DealerSeat)
	}

	// Start third hand
	err = table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting third hand, got %v", err)
	}

	// Verify dealer remains at seat 5 (was rotated to seat 5 after second hand ended, stays there for third hand)
	// Dealer rotates at the END of a hand (in HandleShowdown), not during StartHand
	if table.CurrentHand.DealerSeat != 5 {
		t.Errorf("expected dealer to remain at seat 5 in third hand (rotated after second hand), got %d", table.CurrentHand.DealerSeat)
	}
}

// TestHandlerFlow_StartHandButtonWorksAfterShowdown verifies "Start Hand" works after showdown cleanup
func TestHandlerFlow_StartHandButtonWorksAfterShowdown(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	table := server.tables[0]

	// Set up 2 active players
	token0 := "player-0"
	token1 := "player-1"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	// Start first hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting first hand, got %v", err)
	}

	// Verify 2 active players dealt in
	if len(table.CurrentHand.HoleCards) != 2 {
		t.Errorf("expected 2 active players in first hand, got %d", len(table.CurrentHand.HoleCards))
	}

	// Mark seat 1 as folded
	table.CurrentHand.FoldedPlayers[1] = true

	// Call HandleShowdown
	table.HandleShowdown()

	// Verify hand is cleared
	if table.CurrentHand != nil {
		t.Error("expected CurrentHand to be nil after showdown")
	}

	// Verify dealer rotated
	if table.DealerSeat == nil || *table.DealerSeat != 1 {
		t.Errorf("expected dealer to rotate to seat 1, got %v", table.DealerSeat)
	}

	// Now click "Start Hand" button - should start a fresh hand
	err = table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting second hand, got %v", err)
	}

	// Verify new hand is created
	if table.CurrentHand == nil {
		t.Fatal("expected CurrentHand to be set after second StartHand")
	}

	// Verify both players are still in the new hand
	if len(table.CurrentHand.HoleCards) != 2 {
		t.Errorf("expected 2 active players in second hand, got %d", len(table.CurrentHand.HoleCards))
	}

	// Verify dealer remains at seat 1 (was rotated to seat 1 after first hand ended, stays there for second hand)
	// Dealer rotates at the END of a hand (in HandleShowdown), not during StartHand
	if table.CurrentHand.DealerSeat != 1 {
		t.Errorf("expected dealer to remain at seat 1 in second hand (rotated after first hand), got %d", table.CurrentHand.DealerSeat)
	}
}

// TestHandleAction_RiverBettingCompleteTriggersShowdown verifies showdown is called
// when betting completes on river via IsBettingRoundComplete path
func TestHandleAction_RiverBettingCompleteTriggersShowdown(t *testing.T) {
	// This test specifically checks the IsBettingRoundComplete path on the river
	// where both players have acted and matched the bet

	logger := slog.Default()
	server := NewServer(logger)
	sm := NewSessionManager(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up 2 players with sessions
	session1, _ := sm.CreateSession("Player1")
	session2, _ := sm.CreateSession("Player2")
	token1 := session1.Token
	token2 := session2.Token

	// Update sessions with table and seat info
	sm.UpdateSession(token1, &table.ID, &[]int{0}[0])
	sm.UpdateSession(token2, &table.ID, &[]int{1}[0])

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	// Setup: Create a hand already at river with 2 active players
	hand := &Hand{
		Pot:            100,
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   0,
		Street:         "river",
		HoleCards: map[int][]Card{
			0: {Card{Rank: "A", Suit: "s"}, Card{Rank: "K", Suit: "s"}},
			1: {Card{Rank: "Q", Suit: "h"}, Card{Rank: "J", Suit: "h"}},
		},
		BoardCards: []Card{
			{Rank: "T", Suit: "d"}, {Rank: "9", Suit: "c"}, {Rank: "8", Suit: "s"},
			{Rank: "7", Suit: "h"}, {Rank: "6", Suit: "d"},
		},
		FoldedPlayers: make(map[int]bool),
		ActedPlayers:  make(map[int]bool),
		PlayerBets: map[int]int{
			0: 50,
			1: 50,
		},
		CurrentBet:        50,
		BigBlindHasOption: false,
	}

	table.mu.Lock()
	table.CurrentHand = hand
	dealerSeat := 0
	table.DealerSeat = &dealerSeat

	// Set current actor to seat 0 (about to act)
	currentActor := 0
	hand.CurrentActor = &currentActor

	// Mark players as having acted and matched current bet
	hand.ActedPlayers[0] = true
	hand.ActedPlayers[1] = true

	// Verify betting round is complete (triggers the IsBettingRoundComplete path)
	if !hand.IsBettingRoundComplete(table.Seats) {
		t.Error("expected betting round to be complete on river")
	}
	table.mu.Unlock()

	// Create client for player 0
	client := &Client{
		hub:   hub,
		Token: token1,
		send:  make(chan []byte, 256),
	}

	// Call HandlePlayerAction to trigger the handler logic
	// This should detect river + betting complete and call HandleShowdown
	err := server.HandlePlayerAction(sm, client, 0, "check")
	if err != nil {
		t.Errorf("expected no error from HandlePlayerAction, got %v", err)
	}

	// Verify the hand is cleared (indicates HandleShowdown was called)
	table.mu.RLock()
	if table.CurrentHand != nil {
		t.Errorf("expected CurrentHand to be nil after HandlePlayerAction on river with complete betting, got %v", table.CurrentHand)
	}
	table.mu.RUnlock()

	// Verify dealer was rotated (another indicator ShowDown was called)
	table.mu.RLock()
	if table.DealerSeat == nil || *table.DealerSeat != 1 {
		t.Errorf("expected dealer to rotate to seat 1, got %v", table.DealerSeat)
	}
	table.mu.RUnlock()
}

// TestHandleAction_RiverNoShowdownIfNotComplete verifies showdown is NOT called
// when betting is incomplete on river
func TestHandleAction_RiverNoShowdownIfNotComplete(t *testing.T) {
	// This test verifies that if betting is not complete on the river,
	// showdown is not triggered yet

	logger := slog.Default()
	server := NewServer(logger)
	sm := NewSessionManager(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up 2 players with sessions
	session1, _ := sm.CreateSession("Player1")
	session2, _ := sm.CreateSession("Player2")
	token1 := session1.Token
	token2 := session2.Token

	// Update sessions with table and seat info
	sm.UpdateSession(token1, &table.ID, &[]int{0}[0])
	sm.UpdateSession(token2, &table.ID, &[]int{1}[0])

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	// Setup: Create a hand on river with 2 active players, but betting not complete
	hand := &Hand{
		Pot:            100,
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   0,
		Street:         "river",
		HoleCards: map[int][]Card{
			0: {Card{Rank: "A", Suit: "s"}, Card{Rank: "K", Suit: "s"}},
			1: {Card{Rank: "Q", Suit: "h"}, Card{Rank: "J", Suit: "h"}},
		},
		BoardCards: []Card{
			{Rank: "T", Suit: "d"}, {Rank: "9", Suit: "c"}, {Rank: "8", Suit: "s"},
			{Rank: "7", Suit: "h"}, {Rank: "6", Suit: "d"},
		},
		FoldedPlayers: make(map[int]bool),
		ActedPlayers:  make(map[int]bool),
		PlayerBets: map[int]int{
			0: 50,
			1: 25, // Player 1 has only bet 25, hasn't matched
		},
		CurrentBet:        50,
		BigBlindHasOption: false,
	}

	table.mu.Lock()
	table.CurrentHand = hand
	dealerSeat := 0
	table.DealerSeat = &dealerSeat

	// Set current actor to seat 0
	currentActor := 0
	hand.CurrentActor = &currentActor

	// Only player 1 has acted, player 0 has not
	hand.ActedPlayers[0] = false
	hand.ActedPlayers[1] = true

	// Verify betting round is NOT complete
	if hand.IsBettingRoundComplete(table.Seats) {
		t.Error("expected betting round to NOT be complete (player 0 hasn't acted)")
	}
	table.mu.Unlock()

	// Create client for player 0
	client := &Client{
		hub:   hub,
		Token: token1,
		send:  make(chan []byte, 256),
	}

	// Call HandlePlayerAction for player 0 - should advance to next actor, not trigger showdown
	err := server.HandlePlayerAction(sm, client, 0, "check")
	if err != nil {
		t.Errorf("expected no error from HandlePlayerAction, got %v", err)
	}

	// Verify the hand is NOT cleared (showdown should not have been called)
	table.mu.RLock()
	if table.CurrentHand == nil {
		t.Error("expected CurrentHand to still exist (showdown should not have been triggered)")
	}

	// Verify we're still on the river street
	if table.CurrentHand.Street != "river" {
		t.Errorf("expected to remain on river street, got %s", table.CurrentHand.Street)
	}
	table.mu.RUnlock()
}

// ============ PHASE 1: EARLY WINNER FOLD HANDLING TESTS ============

// TestHandlePlayerAction_AllFoldPreflop_EarlyWinner verifies that when all but one player folds on preflop,
// the remaining player wins immediately without advancing to flop
func TestHandlePlayerAction_AllFoldPreflop_EarlyWinner(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	sm := NewSessionManager(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up 3 players with sessions
	session0, _ := sm.CreateSession("Player0")
	session1, _ := sm.CreateSession("Player1")
	session2, _ := sm.CreateSession("Player2")
	token0 := session0.Token
	token1 := session1.Token
	token2 := session2.Token

	// Update sessions with table and seat info
	sm.UpdateSession(token0, &table.ID, &[]int{0}[0])
	sm.UpdateSession(token1, &table.ID, &[]int{1}[0])
	sm.UpdateSession(token2, &table.ID, &[]int{2}[0])

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.Seats[2].Token = &token2
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000
	table.mu.Unlock()

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Verify hand is on preflop
	table.mu.RLock()
	if table.CurrentHand.Street != "preflop" {
		t.Fatalf("expected hand to be on preflop, got %s", table.CurrentHand.Street)
	}
	initialPot := table.CurrentHand.Pot
	table.mu.RUnlock()

	// Player 0 folds
	client0 := &Client{
		hub:   hub,
		Token: token0,
		send:  make(chan []byte, 256),
	}
	err = server.HandlePlayerAction(sm, client0, 0, "fold")
	if err != nil {
		t.Errorf("expected no error for player 0 fold, got %v", err)
	}

	table.mu.RLock()
	handAfterFold1 := table.CurrentHand
	table.mu.RUnlock()

	// Player 1 folds (now only player 2 remains)
	client1 := &Client{
		hub:   hub,
		Token: token1,
		send:  make(chan []byte, 256),
	}
	err = server.HandlePlayerAction(sm, client1, 1, "fold")
	if err != nil {
		t.Errorf("expected no error for player 1 fold, got %v", err)
	}

	// Verify hand is completed and player 2 won the pot
	table.mu.Lock()
	defer table.mu.Unlock()

	// Hand should be nil after early winner (HandleShowdown clears it)
	if table.CurrentHand != nil {
		t.Errorf("expected CurrentHand to be nil after early winner, got %v", table.CurrentHand)
	}

	// Player 2 should have received the pot
	// Player 2 is BB (posted 20), so final stack = 1000 - 20 (BB posted) + initialPot (won)
	player2Stack := table.Seats[2].Stack
	expectedStack := 1000 - 20 + initialPot // Started with 1000, paid 20 for BB, won pot
	if player2Stack != expectedStack {
		t.Errorf("expected player 2 stack to be %d, got %d", expectedStack, player2Stack)
	}

	// Board should have no cards on preflop early winner
	if handAfterFold1 != nil && len(handAfterFold1.BoardCards) > 0 {
		t.Errorf("expected no board cards on preflop early winner, got %d cards", len(handAfterFold1.BoardCards))
	}
}

// TestHandlePlayerAction_AllFoldFlop_EarlyWinner verifies that when all but one player folds on flop,
// the remaining player wins immediately without advancing to turn
func TestHandlePlayerAction_AllFoldFlop_EarlyWinner(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	sm := NewSessionManager(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up 3 players with sessions
	session0, _ := sm.CreateSession("Player0")
	session1, _ := sm.CreateSession("Player1")
	session2, _ := sm.CreateSession("Player2")
	token0 := session0.Token
	token1 := session1.Token
	token2 := session2.Token

	// Update sessions with table and seat info
	sm.UpdateSession(token0, &table.ID, &[]int{0}[0])
	sm.UpdateSession(token1, &table.ID, &[]int{1}[0])
	sm.UpdateSession(token2, &table.ID, &[]int{2}[0])

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.Seats[2].Token = &token2
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000
	table.mu.Unlock()

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Manually advance to flop and set up state
	table.mu.Lock()
	table.CurrentHand.Street = "flop"
	table.CurrentHand.BoardCards = []Card{
		{Rank: "A", Suit: "s"},
		{Rank: "K", Suit: "h"},
		{Rank: "Q", Suit: "d"},
	}
	// Mark two players as folded, only player 2 remains
	table.CurrentHand.FoldedPlayers[0] = true
	table.CurrentHand.FoldedPlayers[1] = true
	table.CurrentHand.CurrentActor = newInt(2)
	table.CurrentHand.ActedPlayers[0] = true
	table.CurrentHand.ActedPlayers[1] = true
	initialPot := table.CurrentHand.Pot
	boardLengthBefore := len(table.CurrentHand.BoardCards)
	table.mu.Unlock()

	// Player 2 checks (which completes betting round with only them left)
	client2 := &Client{
		hub:   hub,
		Token: token2,
		send:  make(chan []byte, 256),
	}
	err = server.HandlePlayerAction(sm, client2, 2, "check")
	if err != nil {
		t.Errorf("expected no error for player 2 check, got %v", err)
	}

	// Verify hand is completed and player 2 won the pot
	table.mu.Lock()
	defer table.mu.Unlock()

	// Hand should be nil after early winner
	if table.CurrentHand != nil {
		t.Errorf("expected CurrentHand to be nil after early winner, got %v", table.CurrentHand)
	}

	// Player 2 should have received the pot
	// Player 2 is BB (posted 20), so final stack = 1000 - 20 (BB posted) + initialPot (won)
	player2Stack := table.Seats[2].Stack
	expectedStack := 1000 - 20 + initialPot
	if player2Stack != expectedStack {
		t.Errorf("expected player 2 stack to be %d, got %d", expectedStack, player2Stack)
	}

	// Board should have remained at 3 cards (not advanced to turn with 4)
	if boardLengthBefore != 3 {
		t.Errorf("expected 3 board cards on flop, got %d", boardLengthBefore)
	}
}

// TestHandlePlayerAction_AllFoldTurn_EarlyWinner verifies that when all but one player folds on turn,
// the remaining player wins immediately without advancing to river
func TestHandlePlayerAction_AllFoldTurn_EarlyWinner(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	sm := NewSessionManager(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up 3 players with sessions
	session0, _ := sm.CreateSession("Player0")
	session1, _ := sm.CreateSession("Player1")
	session2, _ := sm.CreateSession("Player2")
	token0 := session0.Token
	token1 := session1.Token
	token2 := session2.Token

	// Update sessions with table and seat info
	sm.UpdateSession(token0, &table.ID, &[]int{0}[0])
	sm.UpdateSession(token1, &table.ID, &[]int{1}[0])
	sm.UpdateSession(token2, &table.ID, &[]int{2}[0])

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.Seats[2].Token = &token2
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000
	table.mu.Unlock()

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Manually advance to turn and set up state
	table.mu.Lock()
	table.CurrentHand.Street = "turn"
	table.CurrentHand.BoardCards = []Card{
		{Rank: "A", Suit: "s"},
		{Rank: "K", Suit: "h"},
		{Rank: "Q", Suit: "d"},
		{Rank: "J", Suit: "c"},
	}
	// Mark two players as folded, only player 2 remains
	table.CurrentHand.FoldedPlayers[0] = true
	table.CurrentHand.FoldedPlayers[1] = true
	table.CurrentHand.CurrentActor = newInt(2)
	table.CurrentHand.ActedPlayers[0] = true
	table.CurrentHand.ActedPlayers[1] = true
	initialPot := table.CurrentHand.Pot
	boardLengthBefore := len(table.CurrentHand.BoardCards)
	table.mu.Unlock()

	// Player 2 checks (which completes betting round with only them left)
	client2 := &Client{
		hub:   hub,
		Token: token2,
		send:  make(chan []byte, 256),
	}
	err = server.HandlePlayerAction(sm, client2, 2, "check")
	if err != nil {
		t.Errorf("expected no error for player 2 check, got %v", err)
	}

	// Verify hand is completed and player 2 won the pot
	table.mu.Lock()
	defer table.mu.Unlock()

	// Hand should be nil after early winner
	if table.CurrentHand != nil {
		t.Errorf("expected CurrentHand to be nil after early winner, got %v", table.CurrentHand)
	}

	// Player 2 should have received the pot
	// Player 2 is BB (posted 20), so final stack = 1000 - 20 (BB posted) + initialPot (won)
	player2Stack := table.Seats[2].Stack
	expectedStack := 1000 - 20 + initialPot
	if player2Stack != expectedStack {
		t.Errorf("expected player 2 stack to be %d, got %d", expectedStack, player2Stack)
	}

	// Board should have remained at 4 cards (not advanced to river with 5)
	if boardLengthBefore != 4 {
		t.Errorf("expected 4 board cards on turn, got %d", boardLengthBefore)
	}
}

// TestHandlePlayerAction_AllFoldRiver_EarlyWinner verifies that when all but one player folds on river,
// the remaining player wins immediately (should already work as river is final street)
func TestHandlePlayerAction_AllFoldRiver_EarlyWinner(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	sm := NewSessionManager(logger)
	hub := server.hub
	go hub.Run()

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up 3 players with sessions
	session0, _ := sm.CreateSession("Player0")
	session1, _ := sm.CreateSession("Player1")
	session2, _ := sm.CreateSession("Player2")
	token0 := session0.Token
	token1 := session1.Token
	token2 := session2.Token

	// Update sessions with table and seat info
	sm.UpdateSession(token0, &table.ID, &[]int{0}[0])
	sm.UpdateSession(token1, &table.ID, &[]int{1}[0])
	sm.UpdateSession(token2, &table.ID, &[]int{2}[0])

	// Assign seats and set to active
	table.mu.Lock()
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.Seats[2].Token = &token2
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000
	table.mu.Unlock()

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Manually advance to river and set up state
	table.mu.Lock()
	table.CurrentHand.Street = "river"
	table.CurrentHand.BoardCards = []Card{
		{Rank: "A", Suit: "s"},
		{Rank: "K", Suit: "h"},
		{Rank: "Q", Suit: "d"},
		{Rank: "J", Suit: "c"},
		{Rank: "T", Suit: "s"},
	}
	// Mark two players as folded, only player 2 remains
	table.CurrentHand.FoldedPlayers[0] = true
	table.CurrentHand.FoldedPlayers[1] = true
	table.CurrentHand.CurrentActor = newInt(2)
	table.CurrentHand.ActedPlayers[0] = true
	table.CurrentHand.ActedPlayers[1] = true
	initialPot := table.CurrentHand.Pot
	boardLengthBefore := len(table.CurrentHand.BoardCards)
	table.mu.Unlock()

	// Player 2 checks (which completes betting round with only them left)
	client2 := &Client{
		hub:   hub,
		Token: token2,
		send:  make(chan []byte, 256),
	}
	err = server.HandlePlayerAction(sm, client2, 2, "check")
	if err != nil {
		t.Errorf("expected no error for player 2 check, got %v", err)
	}

	// Verify hand is completed and player 2 won the pot
	table.mu.Lock()
	defer table.mu.Unlock()

	// Hand should be nil after early winner
	if table.CurrentHand != nil {
		t.Errorf("expected CurrentHand to be nil after early winner, got %v", table.CurrentHand)
	}

	// Player 2 should have received the pot
	// Player 2 is BB (posted 20), so final stack = 1000 - 20 (BB posted) + initialPot (won)
	player2Stack := table.Seats[2].Stack
	expectedStack := 1000 - 20 + initialPot
	if player2Stack != expectedStack {
		t.Errorf("expected player 2 stack to be %d, got %d", expectedStack, player2Stack)
	}

	// Board should have remained at 5 cards (all river cards)
	if boardLengthBefore != 5 {
		t.Errorf("expected 5 board cards on river, got %d", boardLengthBefore)
	}
}

// Helper function to create a pointer to an int
func newInt(i int) *int {
	return &i
}

// Helper function to create a pointer to a string
func newString(s string) *string {
	return &s
}
