package server

import (
	"log/slog"
	"sync"
	"testing"
)

// TestNewTable verifies table creation with correct ID, name, and 6 empty seats
func TestNewTable(t *testing.T) {
	table := NewTable("table-1", "Table 1")

	if table == nil {
		t.Fatal("expected table to be created, got nil")
	}

	if table.ID != "table-1" {
		t.Errorf("expected ID 'table-1', got '%s'", table.ID)
	}

	if table.Name != "Table 1" {
		t.Errorf("expected Name 'Table 1', got '%s'", table.Name)
	}

	if table.MaxSeats != 6 {
		t.Errorf("expected MaxSeats 6, got %d", table.MaxSeats)
	}
}

// TestSeatInitialization verifies all seats have correct Index and nil Token
func TestSeatInitialization(t *testing.T) {
	table := NewTable("table-1", "Table 1")

	if len(table.Seats) != 6 {
		t.Errorf("expected 6 seats, got %d", len(table.Seats))
	}

	for i := 0; i < 6; i++ {
		if table.Seats[i].Index != i {
			t.Errorf("seat %d: expected Index %d, got %d", i, i, table.Seats[i].Index)
		}

		if table.Seats[i].Token != nil {
			t.Errorf("seat %d: expected Token nil, got %v", i, table.Seats[i].Token)
		}
	}
}

// TestSeatStatusField verifies Seat has Status field with valid "empty" value for new tables
func TestSeatStatusField(t *testing.T) {
	table := NewTable("table-1", "Table 1")

	if len(table.Seats) != 6 {
		t.Errorf("expected 6 seats, got %d", len(table.Seats))
	}

	for i := 0; i < 6; i++ {
		if table.Seats[i].Status != "empty" {
			t.Errorf("seat %d: expected Status 'empty', got '%s'", i, table.Seats[i].Status)
		}
	}
}

// TestGetOccupiedSeatCount verifies returns 0 for empty table
func TestGetOccupiedSeatCount(t *testing.T) {
	table := NewTable("table-1", "Table 1")

	count := table.GetOccupiedSeatCount()
	if count != 0 {
		t.Errorf("expected 0 occupied seats, got %d", count)
	}
}

// TestGetOccupiedSeatCountWithOccupiedSeats verifies count with manually set tokens
func TestGetOccupiedSeatCountWithOccupiedSeats(t *testing.T) {
	table := NewTable("table-1", "Table 1")

	// Manually set some tokens
	token1 := "player1"
	token2 := "player2"
	token3 := "player3"

	table.Seats[0].Token = &token1
	table.Seats[2].Token = &token2
	table.Seats[5].Token = &token3

	count := table.GetOccupiedSeatCount()
	if count != 3 {
		t.Errorf("expected 3 occupied seats, got %d", count)
	}

	// Set all seats
	token4 := "player4"
	token5 := "player5"
	token6 := "player6"

	table.Seats[1].Token = &token4
	table.Seats[3].Token = &token5
	table.Seats[4].Token = &token6

	count = table.GetOccupiedSeatCount()
	if count != 6 {
		t.Errorf("expected 6 occupied seats, got %d", count)
	}
}

// TestTableThreadSafety verifies concurrent reads/writes with RWMutex
func TestTableThreadSafety(t *testing.T) {
	table := NewTable("table-1", "Table 1")

	const numGoroutines = 5
	const operationsPerGoroutine = 50

	var wg sync.WaitGroup

	// Writer goroutines - set tokens
	for i := 0; i < numGoroutines/2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				seatIdx := (id + j) % 6
				token := "player"
				table.mu.Lock()
				table.Seats[seatIdx].Token = &token
				table.mu.Unlock()
			}
		}(i)
	}

	// Reader goroutines - call GetOccupiedSeatCount
	for i := 0; i < numGoroutines/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				_ = table.GetOccupiedSeatCount()
			}
		}()
	}

	wg.Wait()

	// Verify test completed successfully
	count := table.GetOccupiedSeatCount()
	if count < 0 || count > 6 {
		t.Errorf("invalid occupied seat count: %d", count)
	}
}

// TestServerTablesPreseeded verifies NewServer creates 4 tables with correct IDs/names
func TestServerTablesPreseeded(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)

	if server == nil {
		t.Fatal("expected server to be initialized, got nil")
	}

	if len(server.tables) != 4 {
		t.Errorf("expected 4 tables, got %d", len(server.tables))
	}

	expectedTables := []struct {
		id   string
		name string
	}{
		{"table-1", "Table 1"},
		{"table-2", "Table 2"},
		{"table-3", "Table 3"},
		{"table-4", "Table 4"},
	}

	for i, expected := range expectedTables {
		if server.tables[i] == nil {
			t.Errorf("table %d: expected table to exist, got nil", i)
			continue
		}

		if server.tables[i].ID != expected.id {
			t.Errorf("table %d: expected ID '%s', got '%s'", i, expected.id, server.tables[i].ID)
		}

		if server.tables[i].Name != expected.name {
			t.Errorf("table %d: expected Name '%s', got '%s'", i, expected.name, server.tables[i].Name)
		}

		if server.tables[i].MaxSeats != 6 {
			t.Errorf("table %d: expected MaxSeats 6, got %d", i, server.tables[i].MaxSeats)
		}

		// Verify all seats are empty
		occupiedCount := server.tables[i].GetOccupiedSeatCount()
		if occupiedCount != 0 {
			t.Errorf("table %d: expected 0 occupied seats, got %d", i, occupiedCount)
		}
	}
}

// TestTableAssignSeat verifies assigns to first empty seat (0-5 sequential)
func TestTableAssignSeat(t *testing.T) {
	table := NewTable("table-1", "Table 1")
	token := "player-token-1"

	// Assign to seat 0 (first empty)
	seat, err := table.AssignSeat(&token)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if seat.Index != 0 {
		t.Errorf("expected seat index 0, got %d", seat.Index)
	}

	if seat.Token == nil || *seat.Token != token {
		t.Errorf("expected token '%s', got %v", token, seat.Token)
	}

	// Verify Status is set to "waiting"
	if seat.Status != "waiting" {
		t.Errorf("expected Status 'waiting', got '%s'", seat.Status)
	}

	// Verify it's in the table's seats with correct status
	if table.Seats[0].Token == nil || *table.Seats[0].Token != token {
		t.Errorf("expected table.Seats[0].Token to be '%s'", token)
	}

	if table.Seats[0].Status != "waiting" {
		t.Errorf("expected table.Seats[0].Status to be 'waiting', got '%s'", table.Seats[0].Status)
	}

	// Assign to seat 1 (next empty)
	token2 := "player-token-2"
	seat2, err := table.AssignSeat(&token2)
	if err != nil {
		t.Fatalf("expected no error for second assignment, got %v", err)
	}

	if seat2.Index != 1 {
		t.Errorf("expected seat index 1, got %d", seat2.Index)
	}

	if seat2.Token == nil || *seat2.Token != token2 {
		t.Errorf("expected token '%s', got %v", token2, seat2.Token)
	}

	if seat2.Status != "waiting" {
		t.Errorf("expected Status 'waiting' for seat2, got '%s'", seat2.Status)
	}
}

// TestTableAssignSeatSequential verifies seats are assigned 0-5 sequentially
func TestTableAssignSeatSequential(t *testing.T) {
	table := NewTable("table-1", "Table 1")

	for i := 0; i < 6; i++ {
		token := "player-" + string(rune('0'+i))
		seat, err := table.AssignSeat(&token)
		if err != nil {
			t.Fatalf("assignment %d: expected no error, got %v", i, err)
		}

		if seat.Index != i {
			t.Errorf("assignment %d: expected seat index %d, got %d", i, i, seat.Index)
		}
	}

	// Verify all 6 seats are occupied
	if table.GetOccupiedSeatCount() != 6 {
		t.Errorf("expected 6 occupied seats, got %d", table.GetOccupiedSeatCount())
	}
}

// TestTableAssignSeatWhenFull verifies returns error when all 6 seats occupied
func TestTableAssignSeatWhenFull(t *testing.T) {
	table := NewTable("table-1", "Table 1")

	// Fill all 6 seats
	for i := 0; i < 6; i++ {
		token := "player-" + string(rune('0'+i))
		_, err := table.AssignSeat(&token)
		if err != nil {
			t.Fatalf("seat %d: expected no error, got %v", i, err)
		}
	}

	// Try to assign 7th seat
	token7 := "player-7"
	seat, err := table.AssignSeat(&token7)
	if err == nil {
		t.Fatal("expected error when table is full, got nil")
	}

	if seat != (Seat{}) {
		t.Errorf("expected empty seat when table is full, got %v", seat)
	}

	if err.Error() != "table is full" {
		t.Errorf("expected error message 'table is full', got '%s'", err.Error())
	}
}

// TestTableClearSeat verifies clears seat by token, sets Token to nil and Status to "empty"
func TestTableClearSeat(t *testing.T) {
	table := NewTable("table-1", "Table 1")
	token1 := "player-1"
	token2 := "player-2"

	// Assign two seats
	_, _ = table.AssignSeat(&token1)
	_, _ = table.AssignSeat(&token2)

	// Verify they're assigned with "waiting" status
	if table.Seats[0].Token == nil || *table.Seats[0].Token != token1 {
		t.Fatal("expected seat 0 to have token1")
	}

	if table.Seats[0].Status != "waiting" {
		t.Errorf("expected seat 0 Status to be 'waiting', got '%s'", table.Seats[0].Status)
	}

	if table.Seats[1].Token == nil || *table.Seats[1].Token != token2 {
		t.Fatal("expected seat 1 to have token2")
	}

	if table.Seats[1].Status != "waiting" {
		t.Errorf("expected seat 1 Status to be 'waiting', got '%s'", table.Seats[1].Status)
	}

	// Clear seat 1 (token2)
	err := table.ClearSeat(&token2)
	if err != nil {
		t.Fatalf("expected no error when clearing seat, got %v", err)
	}

	// Verify seat 1 is now empty with "empty" status
	if table.Seats[1].Token != nil {
		t.Errorf("expected seat 1 Token to be nil after clearing, got %v", table.Seats[1].Token)
	}

	if table.Seats[1].Status != "empty" {
		t.Errorf("expected seat 1 Status to be 'empty', got '%s'", table.Seats[1].Status)
	}

	// Verify seat 0 is still occupied with "waiting" status
	if table.Seats[0].Token == nil || *table.Seats[0].Token != token1 {
		t.Errorf("expected seat 0 to still have token1")
	}

	if table.Seats[0].Status != "waiting" {
		t.Errorf("expected seat 0 Status to still be 'waiting', got '%s'", table.Seats[0].Status)
	}

	// Verify occupied count is 1
	if table.GetOccupiedSeatCount() != 1 {
		t.Errorf("expected 1 occupied seat, got %d", table.GetOccupiedSeatCount())
	}
}

// TestTableClearSeatNotFound verifies returns error when token not found
func TestTableClearSeatNotFound(t *testing.T) {
	table := NewTable("table-1", "Table 1")
	token1 := "player-1"
	tokenNotAssigned := "player-not-assigned"

	// Assign one seat
	_, _ = table.AssignSeat(&token1)

	// Try to clear a non-existent seat
	err := table.ClearSeat(&tokenNotAssigned)
	if err == nil {
		t.Fatal("expected error when clearing non-existent seat, got nil")
	}

	if err.Error() != "seat not found" {
		t.Errorf("expected error message 'seat not found', got '%s'", err.Error())
	}

	// Verify seat 0 is still occupied
	if table.Seats[0].Token == nil || *table.Seats[0].Token != token1 {
		t.Errorf("expected seat 0 to still have token1")
	}

	// Verify seat 1's Stack is reset to 0 after clearing
	if table.Seats[1].Stack != 0 {
		t.Errorf("expected seat 1 Stack to be 0 after clearing, got %d", table.Seats[1].Stack)
	}
}

// TestTableGetSeatByToken verifies returns seat if player is seated at table
func TestTableGetSeatByToken(t *testing.T) {
	table := NewTable("table-1", "Table 1")
	token1 := "player-1"
	token2 := "player-2"

	// Assign two seats
	_, _ = table.AssignSeat(&token1)
	_, _ = table.AssignSeat(&token2)

	// Get seat by token1
	seat, found := table.GetSeatByToken(&token1)
	if !found {
		t.Fatal("expected seat to be found for token1, got not found")
	}

	if seat.Index != 0 {
		t.Errorf("expected seat index 0, got %d", seat.Index)
	}

	if seat.Token == nil || *seat.Token != token1 {
		t.Errorf("expected token '%s', got %v", token1, seat.Token)
	}

	// Get seat by token2
	seat, found = table.GetSeatByToken(&token2)
	if !found {
		t.Fatal("expected seat to be found for token2, got not found")
	}

	if seat.Index != 1 {
		t.Errorf("expected seat index 1 for token2, got %d", seat.Index)
	}

	if seat.Token == nil || *seat.Token != token2 {
		t.Errorf("expected token '%s', got %v", token2, seat.Token)
	}
}

// TestTableGetSeatByTokenNotFound verifies returns found=false when token not found
func TestTableGetSeatByTokenNotFound(t *testing.T) {
	table := NewTable("table-1", "Table 1")
	token1 := "player-1"
	tokenNotAssigned := "player-not-assigned"

	// Assign one seat
	_, _ = table.AssignSeat(&token1)

	// Get seat by non-existent token
	seat, found := table.GetSeatByToken(&tokenNotAssigned)
	if found {
		t.Errorf("expected not found for non-existent token, got found")
	}
	if seat != (Seat{}) {
		t.Errorf("expected empty seat for non-existent token, got %v", seat)
	}

	// Empty table should return not found
	emptyTable := NewTable("empty", "Empty")
	seat, found = emptyTable.GetSeatByToken(&token1)
	if found {
		t.Errorf("expected not found for empty table, got found")
	}
	if seat != (Seat{}) {
		t.Errorf("expected empty seat for empty table, got %v", seat)
	}
}

// TestTableConcurrentAssignments verifies multiple goroutines assign seats safely
func TestTableConcurrentAssignments(t *testing.T) {
	table := NewTable("table-1", "Table 1")

	const numGoroutines = 6
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			token := "player-" + string(rune('0'+id))
			seat, err := table.AssignSeat(&token)
			if err == nil && seat != (Seat{}) {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	if successCount != 6 {
		t.Errorf("expected 6 successful assignments, got %d", successCount)
	}

	if table.GetOccupiedSeatCount() != 6 {
		t.Errorf("expected 6 occupied seats, got %d", table.GetOccupiedSeatCount())
	}
}

// TestCardString verifies card representation (e.g., "As" for Ace of Spades, "Kh" for King of Hearts)
func TestCardString(t *testing.T) {
	tests := []struct {
		rank string
		suit string
		want string
	}{
		{"A", "s", "As"},
		{"K", "h", "Kh"},
		{"Q", "d", "Qd"},
		{"J", "c", "Jc"},
		{"T", "s", "Ts"},
		{"9", "h", "9h"},
		{"2", "d", "2d"},
	}

	for _, tt := range tests {
		card := Card{Rank: tt.rank, Suit: tt.suit}
		if got := card.String(); got != tt.want {
			t.Errorf("Card{Rank: %q, Suit: %q}.String() = %q, want %q", tt.rank, tt.suit, got, tt.want)
		}
	}
}

// TestNewDeck verifies 52-card deck generation with all unique cards
func TestNewDeck(t *testing.T) {
	deck := NewDeck()

	// Verify exactly 52 cards
	if len(deck) != 52 {
		t.Errorf("expected 52 cards in deck, got %d", len(deck))
	}

	// Verify all cards are unique
	cardMap := make(map[string]bool)
	for _, card := range deck {
		cardStr := card.String()
		if cardMap[cardStr] {
			t.Errorf("duplicate card found: %s", cardStr)
		}
		cardMap[cardStr] = true
	}

	// Verify all 13 ranks are present for each suit
	ranks := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K"}
	suits := []string{"s", "h", "d", "c"}

	for _, suit := range suits {
		for _, rank := range ranks {
			cardStr := rank + suit
			if !cardMap[cardStr] {
				t.Errorf("expected card %s in deck, not found", cardStr)
			}
		}
	}
}

// TestHandInitialization verifies Hand struct fields are properly initialized
func TestHandInitialization(t *testing.T) {
	hand := &Hand{
		DealerSeat:     2,
		SmallBlindSeat: 3,
		BigBlindSeat:   4,
		Pot:            0,
		Deck:           NewDeck(),
		HoleCards:      make(map[int][]Card),
	}

	if hand.DealerSeat != 2 {
		t.Errorf("expected DealerSeat 2, got %d", hand.DealerSeat)
	}

	if hand.SmallBlindSeat != 3 {
		t.Errorf("expected SmallBlindSeat 3, got %d", hand.SmallBlindSeat)
	}

	if hand.BigBlindSeat != 4 {
		t.Errorf("expected BigBlindSeat 4, got %d", hand.BigBlindSeat)
	}

	if hand.Pot != 0 {
		t.Errorf("expected Pot 0, got %d", hand.Pot)
	}

	if len(hand.Deck) != 52 {
		t.Errorf("expected 52 cards in deck, got %d", len(hand.Deck))
	}

	if hand.HoleCards == nil {
		t.Error("expected HoleCards to be initialized, got nil")
	}

	if len(hand.HoleCards) != 0 {
		t.Errorf("expected HoleCards to be empty, got %d entries", len(hand.HoleCards))
	}
}

// TestSeatWithStack verifies Stack field is added to Seat struct and defaults to 1000
func TestSeatWithStack(t *testing.T) {
	table := NewTable("table-1", "Table 1")

	// Verify Stack field exists and defaults to 0 on new table
	for i := 0; i < 6; i++ {
		if table.Seats[i].Stack != 0 {
			t.Errorf("seat %d: expected Stack 0 on empty seat, got %d", i, table.Seats[i].Stack)
		}
	}

	// Assign a seat and verify Stack is set to 1000
	token := "player-1"
	seat, err := table.AssignSeat(&token)
	if err != nil {
		t.Fatalf("expected no error assigning seat, got %v", err)
	}

	if seat.Stack != 1000 {
		t.Errorf("expected Stack 1000 on assigned seat, got %d", seat.Stack)
	}

	// Verify it's persisted in the table
	if table.Seats[0].Stack != 1000 {
		t.Errorf("expected table.Seats[0].Stack to be 1000, got %d", table.Seats[0].Stack)
	}
}

// TestTableClearSeatResetStack verifies Stack is reset to 0 after clearing a seat
func TestTableClearSeatResetStack(t *testing.T) {
	table := NewTable("table-1", "Table 1")
	token1 := "player-1"

	// Assign a seat
	seat, err := table.AssignSeat(&token1)
	if err != nil {
		t.Fatalf("expected no error assigning seat, got %v", err)
	}

	// Verify Stack is 1000 after assignment
	if seat.Stack != 1000 {
		t.Errorf("expected Stack 1000 after assignment, got %d", seat.Stack)
	}

	if table.Seats[0].Stack != 1000 {
		t.Errorf("expected table.Seats[0].Stack to be 1000, got %d", table.Seats[0].Stack)
	}

	// Clear the seat
	err = table.ClearSeat(&token1)
	if err != nil {
		t.Fatalf("expected no error clearing seat, got %v", err)
	}

	// Verify Stack is reset to 0 after clearing
	if table.Seats[0].Stack != 0 {
		t.Errorf("expected Stack to be 0 after clearing, got %d", table.Seats[0].Stack)
	}
}
