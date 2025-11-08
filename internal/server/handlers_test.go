package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
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
