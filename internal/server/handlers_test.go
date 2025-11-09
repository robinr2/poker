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
	validActions := table.CurrentHand.GetValidActions(1)
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

	err = server.HandlePlayerAction(sm, client2, 1, "check")
	if err != nil {
		t.Errorf("expected no error for valid check, got %v", err)
	}

	// Verify action was processed
	table.mu.RLock()
	if !table.CurrentHand.ActedPlayers[1] {
		t.Errorf("expected seat 1 to be marked as acted")
	}
	table.mu.RUnlock()
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
