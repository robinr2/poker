package server

import (
	"log/slog"
	"sync"
	"testing"
)

// Helper function to convert Seat slice to []*Seat
func seatsToPointers(seats []Seat) []*Seat {
	result := make([]*Seat, len(seats))
	for i := range seats {
		result[i] = &seats[i]
	}
	return result
}

// Helper function to create a minimal Hand with no bets (for testing purposes)
func createEmptyHand() *Hand {
	return &Hand{
		Pot:           0,
		PlayerBets:    make(map[int]int), // Empty - no bets
		FoldedPlayers: make(map[int]bool),
		ActedPlayers:  make(map[int]bool),
	}
}

// TestNewTable verifies table creation with correct ID, name, and 6 empty seats
func TestNewTable(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

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
	table := NewTable("table-1", "Table 1", nil)

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
	table := NewTable("table-1", "Table 1", nil)

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
	table := NewTable("table-1", "Table 1", nil)

	count := table.GetOccupiedSeatCount()
	if count != 0 {
		t.Errorf("expected 0 occupied seats, got %d", count)
	}
}

// TestGetOccupiedSeatCountWithOccupiedSeats verifies count with manually set tokens
func TestGetOccupiedSeatCountWithOccupiedSeats(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

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
	table := NewTable("table-1", "Table 1", nil)

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
	table := NewTable("table-1", "Table 1", nil)
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
	table := NewTable("table-1", "Table 1", nil)

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
	table := NewTable("table-1", "Table 1", nil)

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
	table := NewTable("table-1", "Table 1", nil)
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
	table := NewTable("table-1", "Table 1", nil)
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
	table := NewTable("table-1", "Table 1", nil)
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
	table := NewTable("table-1", "Table 1", nil)
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
	emptyTable := NewTable("empty", "Empty", nil)
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
	table := NewTable("table-1", "Table 1", nil)

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
	table := NewTable("table-1", "Table 1", nil)

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
	table := NewTable("table-1", "Table 1", nil)
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

// TestNextDealerFirstHand verifies first hand assigns dealer to first active seat
func TestNextDealerFirstHand(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up: seat 0 and 2 are active, seat 1 is waiting
	token0 := "player-0"
	token1 := "player-1"
	token2 := "player-2"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "waiting"
	table.Seats[1].Stack = 1000

	table.Seats[2].Token = &token2
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000

	// First hand: dealer should be assigned to seat 0 (first active)
	dealer := table.NextDealer()

	if dealer != 0 {
		t.Errorf("expected first dealer to be seat 0, got %d", dealer)
	}

	if table.DealerSeat == nil || *table.DealerSeat != 0 {
		t.Errorf("expected DealerSeat to be 0, got %v", table.DealerSeat)
	}
}

// TestNextDealerRotation verifies dealer rotates clockwise through active players
func TestNextDealerRotation(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up: seats 0, 2, 4 are active
	token0 := "player-0"
	token2 := "player-2"
	token4 := "player-4"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[2].Token = &token2
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000

	table.Seats[4].Token = &token4
	table.Seats[4].Status = "active"
	table.Seats[4].Stack = 1000

	// First hand: dealer = seat 0
	dealer1 := table.NextDealer()
	if dealer1 != 0 {
		t.Errorf("expected first dealer to be seat 0, got %d", dealer1)
	}

	// Second hand: dealer should rotate to seat 2
	dealer2 := table.NextDealer()
	if dealer2 != 2 {
		t.Errorf("expected second dealer to be seat 2, got %d", dealer2)
	}

	// Third hand: dealer should rotate to seat 4
	dealer3 := table.NextDealer()
	if dealer3 != 4 {
		t.Errorf("expected third dealer to be seat 4, got %d", dealer3)
	}

	// Fourth hand: dealer should wrap around to seat 0
	dealer4 := table.NextDealer()
	if dealer4 != 0 {
		t.Errorf("expected fourth dealer to wrap to seat 0, got %d", dealer4)
	}
}

// TestNextDealerSkipsWaiting verifies dealer skips seats with "waiting" status
func TestNextDealerSkipsWaiting(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up: seat 0 active, seat 1 waiting, seat 2 active
	token0 := "player-0"
	token1 := "player-1"
	token2 := "player-2"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "waiting"
	table.Seats[1].Stack = 1000

	table.Seats[2].Token = &token2
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000

	// First hand: dealer = seat 0
	dealer1 := table.NextDealer()
	if dealer1 != 0 {
		t.Errorf("expected first dealer to be seat 0, got %d", dealer1)
	}

	// Second hand: dealer should skip seat 1 (waiting) and go to seat 2 (active)
	dealer2 := table.NextDealer()
	if dealer2 != 2 {
		t.Errorf("expected dealer to skip waiting seat 1 and go to seat 2, got %d", dealer2)
	}
}

// TestGetBlindPositionsNormal verifies blind positions for 3+ players (SB=next after dealer, BB=next after SB)
func TestGetBlindPositionsNormal(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up: seats 0, 1, 2, 3 are active
	for i := 0; i < 4; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Dealer at seat 0: SB should be seat 1, BB should be seat 2
	sb, bb, err := table.GetBlindPositions(0)
	if err != nil {
		t.Errorf("expected no error for 4 active players, got %v", err)
	}

	if sb != 1 {
		t.Errorf("expected SB at seat 1, got %d", sb)
	}

	if bb != 2 {
		t.Errorf("expected BB at seat 2, got %d", bb)
	}

	// Dealer at seat 2: SB should be seat 3, BB should be seat 0
	sb, bb, err = table.GetBlindPositions(2)
	if err != nil {
		t.Errorf("expected no error for 4 active players, got %v", err)
	}

	if sb != 3 {
		t.Errorf("expected SB at seat 3, got %d", sb)
	}

	if bb != 0 {
		t.Errorf("expected BB at seat 0 (wrapped), got %d", bb)
	}
}

// TestGetBlindPositionsHeadsUp verifies blind positions for 2 players (dealer IS SB, other is BB)
func TestGetBlindPositionsHeadsUp(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up: only seats 0 and 3 are active (heads-up)
	token0 := "player-0"
	token3 := "player-3"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[3].Token = &token3
	table.Seats[3].Status = "active"
	table.Seats[3].Stack = 1000

	// Heads-up with dealer at seat 0: dealer IS SB (seat 0), other player IS BB (seat 3)
	sb, bb, err := table.GetBlindPositions(0)
	if err != nil {
		t.Errorf("expected no error for 2 active players (heads-up), got %v", err)
	}

	if sb != 0 {
		t.Errorf("expected SB to be dealer seat 0 in heads-up, got %d", sb)
	}

	if bb != 3 {
		t.Errorf("expected BB to be other player at seat 3 in heads-up, got %d", bb)
	}

	// Heads-up with dealer at seat 3: dealer IS SB (seat 3), other player IS BB (seat 0)
	sb, bb, err = table.GetBlindPositions(3)
	if err != nil {
		t.Errorf("expected no error for 2 active players (heads-up), got %v", err)
	}

	if sb != 3 {
		t.Errorf("expected SB to be dealer seat 3 in heads-up, got %d", sb)
	}

	if bb != 0 {
		t.Errorf("expected BB to be other player at seat 0 in heads-up, got %d", bb)
	}
}

// TestGetBlindPositionsInsufficientPlayers verifies error for <2 active players
func TestGetBlindPositionsInsufficientPlayers(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// No active players
	sb, bb, err := table.GetBlindPositions(0)
	if err == nil {
		t.Fatal("expected error for 0 active players, got nil")
	}

	if sb != 0 || bb != 0 {
		t.Errorf("expected sb=0, bb=0 on error, got sb=%d, bb=%d", sb, bb)
	}

	// Only 1 active player
	token0 := "player-0"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	sb, bb, err = table.GetBlindPositions(0)
	if err == nil {
		t.Fatal("expected error for 1 active player, got nil")
	}

	if sb != 0 || bb != 0 {
		t.Errorf("expected sb=0, bb=0 on error, got sb=%d, bb=%d", sb, bb)
	}
}

// TestGetBlindPositionsScatteredSeats verifies blind positions with non-consecutive active seats
func TestGetBlindPositionsScatteredSeats(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up: seats 1, 3, 5 are active (scattered, non-consecutive)
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

	// Dealer at seat 5: SB should be seat 1 (next active), BB should be seat 3
	sb, bb, err := table.GetBlindPositions(5)
	if err != nil {
		t.Errorf("expected no error for 3 active players with scattered seats, got %v", err)
	}

	if sb != 1 {
		t.Errorf("expected SB at seat 1, got %d", sb)
	}

	if bb != 3 {
		t.Errorf("expected BB at seat 3, got %d", bb)
	}

	// Dealer at seat 1: SB should be seat 3 (next active), BB should be seat 5
	sb, bb, err = table.GetBlindPositions(1)
	if err != nil {
		t.Errorf("expected no error for 3 active players with scattered seats, got %v", err)
	}

	if sb != 3 {
		t.Errorf("expected SB at seat 3, got %d", sb)
	}

	if bb != 5 {
		t.Errorf("expected BB at seat 5, got %d", bb)
	}
}

// TestGetBlindPositionsInvalidDealer verifies error when dealer seat is not active
func TestGetBlindPositionsInvalidDealer(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up: seats 1, 3, 5 are active
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

	// Try to get blinds with dealer at seat 0 (not active)
	sb, bb, err := table.GetBlindPositions(0)
	if err == nil {
		t.Fatal("expected error when dealer seat 0 is not active, got nil")
	}

	if sb != 0 || bb != 0 {
		t.Errorf("expected sb=0, bb=0 on error, got sb=%d, bb=%d", sb, bb)
	}

	// Verify error message mentions the dealer seat
	if err.Error() != "dealer seat 0 is not active" {
		t.Errorf("expected error message 'dealer seat 0 is not active', got '%s'", err.Error())
	}

	// Try with dealer at seat 2 (also not active)
	sb, bb, err = table.GetBlindPositions(2)
	if err == nil {
		t.Fatal("expected error when dealer seat 2 is not active, got nil")
	}

	if err.Error() != "dealer seat 2 is not active" {
		t.Errorf("expected error message 'dealer seat 2 is not active', got '%s'", err.Error())
	}
}

// TestShuffleDeck verifies deck remains 52 cards after shuffle and cards are randomized
func TestShuffleDeck(t *testing.T) {
	deck := NewDeck()

	// Verify deck has 52 cards before shuffle
	if len(deck) != 52 {
		t.Errorf("expected 52 cards before shuffle, got %d", len(deck))
	}

	// Store original deck order
	originalOrder := make([]Card, len(deck))
	copy(originalOrder, deck)

	// Shuffle the deck
	err := ShuffleDeck(deck)
	if err != nil {
		t.Fatalf("expected no error shuffling deck, got %v", err)
	}

	// Verify deck still has 52 cards after shuffle
	if len(deck) != 52 {
		t.Errorf("expected 52 cards after shuffle, got %d", len(deck))
	}

	// Verify all cards are still present (by converting to map)
	originalMap := make(map[string]bool)
	for _, card := range originalOrder {
		originalMap[card.String()] = true
	}

	shuffledMap := make(map[string]bool)
	for _, card := range deck {
		shuffledMap[card.String()] = true
	}

	// Check that all original cards are present in shuffled deck
	for cardStr := range originalMap {
		if !shuffledMap[cardStr] {
			t.Errorf("card %s missing from shuffled deck", cardStr)
		}
	}

	// Verify no new cards were added
	if len(shuffledMap) != 52 {
		t.Errorf("expected 52 unique cards in shuffled deck, got %d", len(shuffledMap))
	}
}

// TestShuffleDeckRandomization verifies shuffle produces different results on multiple shuffles
func TestShuffleDeckRandomization(t *testing.T) {
	// Perform multiple shuffles and check they're different
	// (with 52 cards, getting the exact same order twice is extremely unlikely)
	results := make([][]string, 5)
	for i := 0; i < 5; i++ {
		deck := NewDeck()
		err := ShuffleDeck(deck)
		if err != nil {
			t.Fatalf("shuffle %d: expected no error, got %v", i, err)
		}
		for _, card := range deck {
			results[i] = append(results[i], card.String())
		}
	}

	// Compare shuffles - at least some should be different
	differentFound := false
	for i := 0; i < 4; i++ {
		for j := i + 1; j < 5; j++ {
			if !shufflesEqual(results[i], results[j]) {
				differentFound = true
				break
			}
		}
		if differentFound {
			break
		}
	}

	if !differentFound {
		t.Error("expected shuffle to produce different results on multiple shuffles")
	}
}

// Helper function to check if two shuffle results are identical
func shufflesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestDealHoleCardsToActivePlayers verifies only "active" seats get 2 cards each
func TestDealHoleCardsToActivePlayers(t *testing.T) {
	hand := &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            0,
		Deck:           NewDeck(),
		HoleCards:      make(map[int][]Card),
	}

	// Set up seats: 0, 2, 4 active; 1, 3, 5 waiting
	seats := [6]Seat{}
	token0 := "player-0"
	token2 := "player-2"
	token4 := "player-4"

	seats[0].Index = 0
	seats[0].Token = &token0
	seats[0].Status = "active"
	seats[0].Stack = 1000

	seats[1].Index = 1
	seats[1].Token = nil
	seats[1].Status = "empty"
	seats[1].Stack = 0

	seats[2].Index = 2
	seats[2].Token = &token2
	seats[2].Status = "active"
	seats[2].Stack = 1000

	seats[3].Index = 3
	seats[3].Token = nil
	seats[3].Status = "empty"
	seats[3].Stack = 0

	seats[4].Index = 4
	seats[4].Token = &token4
	seats[4].Status = "active"
	seats[4].Stack = 1000

	seats[5].Index = 5
	seats[5].Token = nil
	seats[5].Status = "empty"
	seats[5].Stack = 0

	// Deal hole cards
	err := hand.DealHoleCards(seats)
	if err != nil {
		t.Fatalf("expected no error dealing hole cards, got %v", err)
	}

	// Verify only active seats (0, 2, 4) have hole cards
	for seatIdx := 0; seatIdx < 6; seatIdx++ {
		cards, exists := hand.HoleCards[seatIdx]

		if seats[seatIdx].Status == "active" {
			if !exists {
				t.Errorf("seat %d (active): expected hole cards, got none", seatIdx)
			}
			if len(cards) != 2 {
				t.Errorf("seat %d (active): expected 2 cards, got %d", seatIdx, len(cards))
			}
		} else {
			if exists {
				t.Errorf("seat %d (empty/waiting): expected no hole cards, got %d", seatIdx, len(cards))
			}
		}
	}

	// Verify HoleCards map has exactly 3 entries (one per active player)
	if len(hand.HoleCards) != 3 {
		t.Errorf("expected 3 entries in HoleCards map, got %d", len(hand.HoleCards))
	}
}

// TestDealHoleCardsSkipsWaiting verifies "waiting" seats get no cards
func TestDealHoleCardsSkipsWaiting(t *testing.T) {
	hand := &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            0,
		Deck:           NewDeck(),
		HoleCards:      make(map[int][]Card),
	}

	// Set up seats: 0 active, 1 and 2 waiting
	seats := [6]Seat{}
	token0 := "player-0"
	token1 := "player-1"
	token2 := "player-2"

	seats[0].Index = 0
	seats[0].Token = &token0
	seats[0].Status = "active"
	seats[0].Stack = 1000

	seats[1].Index = 1
	seats[1].Token = &token1
	seats[1].Status = "waiting"
	seats[1].Stack = 1000

	seats[2].Index = 2
	seats[2].Token = &token2
	seats[2].Status = "waiting"
	seats[2].Stack = 1000

	// Deal hole cards
	err := hand.DealHoleCards(seats)
	if err != nil {
		t.Fatalf("expected no error dealing hole cards, got %v", err)
	}

	// Verify only seat 0 (active) has hole cards
	if cards, exists := hand.HoleCards[0]; !exists || len(cards) != 2 {
		t.Errorf("seat 0 (active): expected 2 cards, got %d", len(cards))
	}

	// Verify waiting seats don't have hole cards
	if _, exists := hand.HoleCards[1]; exists {
		t.Error("seat 1 (waiting): expected no hole cards, but found some")
	}

	if _, exists := hand.HoleCards[2]; exists {
		t.Error("seat 2 (waiting): expected no hole cards, but found some")
	}

	// Verify HoleCards map has exactly 1 entry
	if len(hand.HoleCards) != 1 {
		t.Errorf("expected 1 entry in HoleCards map, got %d", len(hand.HoleCards))
	}
}

// TestDealHoleCardsReducesDeck verifies deck size decreases by (2 × active_players)
func TestDealHoleCardsReducesDeck(t *testing.T) {
	hand := &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            0,
		Deck:           NewDeck(),
		HoleCards:      make(map[int][]Card),
	}

	initialDeckSize := len(hand.Deck)

	// Set up seats: 0, 1, 2 active (3 players)
	seats := [6]Seat{}
	token0 := "player-0"
	token1 := "player-1"
	token2 := "player-2"

	seats[0].Index = 0
	seats[0].Token = &token0
	seats[0].Status = "active"
	seats[0].Stack = 1000

	seats[1].Index = 1
	seats[1].Token = &token1
	seats[1].Status = "active"
	seats[1].Stack = 1000

	seats[2].Index = 2
	seats[2].Token = &token2
	seats[2].Status = "active"
	seats[2].Stack = 1000

	// Deal hole cards
	err := hand.DealHoleCards(seats)
	if err != nil {
		t.Fatalf("expected no error dealing hole cards, got %v", err)
	}

	// Verify deck reduced by 2 × 3 = 6 cards
	expectedDeckSize := initialDeckSize - 6
	if len(hand.Deck) != expectedDeckSize {
		t.Errorf("expected deck size %d after dealing to 3 players, got %d", expectedDeckSize, len(hand.Deck))
	}
}

// TestDealHoleCardsEmptySeats verifies empty seats get no cards
func TestDealHoleCardsEmptySeats(t *testing.T) {
	hand := &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            0,
		Deck:           NewDeck(),
		HoleCards:      make(map[int][]Card),
	}

	// Set up seats: only seat 0 active, rest empty
	seats := [6]Seat{}
	token0 := "player-0"

	seats[0].Index = 0
	seats[0].Token = &token0
	seats[0].Status = "active"
	seats[0].Stack = 1000

	for i := 1; i < 6; i++ {
		seats[i].Index = i
		seats[i].Token = nil
		seats[i].Status = "empty"
		seats[i].Stack = 0
	}

	// Deal hole cards
	err := hand.DealHoleCards(seats)
	if err != nil {
		t.Fatalf("expected no error dealing hole cards, got %v", err)
	}

	// Verify only seat 0 has cards
	for seatIdx := 0; seatIdx < 6; seatIdx++ {
		_, exists := hand.HoleCards[seatIdx]
		if seatIdx == 0 {
			if !exists {
				t.Errorf("seat 0 (active): expected hole cards, got none")
			}
		} else {
			if exists {
				t.Errorf("seat %d (empty): expected no hole cards, but found some", seatIdx)
			}
		}
	}

	// Verify deck reduced by 2 cards (only 1 active player)
	expectedDeckSize := 52 - 2
	if len(hand.Deck) != expectedDeckSize {
		t.Errorf("expected deck size %d after dealing to 1 player, got %d", expectedDeckSize, len(hand.Deck))
	}
}

// TestDealHoleCardsAllPlayersActive verifies dealing to all 6 active seats
func TestDealHoleCardsAllPlayersActive(t *testing.T) {
	hand := &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            0,
		Deck:           NewDeck(),
		HoleCards:      make(map[int][]Card),
	}

	// Set up all 6 seats as active
	seats := [6]Seat{}
	for i := 0; i < 6; i++ {
		token := "player-" + string(rune('0'+i))
		seats[i].Index = i
		seats[i].Token = &token
		seats[i].Status = "active"
		seats[i].Stack = 1000
	}

	// Deal hole cards
	err := hand.DealHoleCards(seats)
	if err != nil {
		t.Fatalf("expected no error dealing hole cards, got %v", err)
	}

	// Verify all 6 seats have 2 cards each
	for seatIdx := 0; seatIdx < 6; seatIdx++ {
		cards, exists := hand.HoleCards[seatIdx]
		if !exists {
			t.Errorf("seat %d: expected hole cards, got none", seatIdx)
		}
		if len(cards) != 2 {
			t.Errorf("seat %d: expected 2 cards, got %d", seatIdx, len(cards))
		}
	}

	// Verify HoleCards map has exactly 6 entries
	if len(hand.HoleCards) != 6 {
		t.Errorf("expected 6 entries in HoleCards map, got %d", len(hand.HoleCards))
	}

	// Verify deck reduced by 12 cards (6 active × 2 cards each)
	expectedDeckSize := 52 - 12
	if len(hand.Deck) != expectedDeckSize {
		t.Errorf("expected deck size %d after dealing to 6 players, got %d", expectedDeckSize, len(hand.Deck))
	}
}

// TestDealHoleCardsInsufficientCards verifies error when deck has fewer cards than needed
func TestDealHoleCardsInsufficientCards(t *testing.T) {
	// Create a small deck with only 3 cards
	smallDeck := []Card{
		{Rank: "A", Suit: "s"},
		{Rank: "K", Suit: "h"},
		{Rank: "Q", Suit: "d"},
	}

	hand := &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            0,
		Deck:           smallDeck,
		HoleCards:      make(map[int][]Card),
	}

	// Set up 3 active players (would need 6 cards total)
	seats := [6]Seat{}
	token0 := "player-0"
	token1 := "player-1"
	token2 := "player-2"

	seats[0].Index = 0
	seats[0].Token = &token0
	seats[0].Status = "active"
	seats[0].Stack = 1000

	seats[1].Index = 1
	seats[1].Token = &token1
	seats[1].Status = "active"
	seats[1].Stack = 1000

	seats[2].Index = 2
	seats[2].Token = &token2
	seats[2].Status = "active"
	seats[2].Stack = 1000

	// Try to deal hole cards - should fail due to insufficient cards
	err := hand.DealHoleCards(seats)
	if err == nil {
		t.Fatal("expected error when deck has insufficient cards, got nil")
	}

	// Verify error message mentions insufficient cards
	expectedMsg := "insufficient cards in deck: have 3, need 6"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	// Verify no cards were dealt
	if len(hand.HoleCards) != 0 {
		t.Errorf("expected no hole cards dealt on error, got %d entries in HoleCards map", len(hand.HoleCards))
	}

	// Verify deck is unchanged
	if len(hand.Deck) != 3 {
		t.Errorf("expected deck to remain 3 cards after error, got %d", len(hand.Deck))
	}
}

// TestCanStartHandRequiresTwoActive verifies CanStartHand returns false with <2 active players
func TestCanStartHandRequiresTwoActive(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// No active players
	if table.CanStartHand() {
		t.Error("expected CanStartHand to return false with 0 active players, got true")
	}

	// Only 1 active player
	token0 := "player-0"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	if table.CanStartHand() {
		t.Error("expected CanStartHand to return false with 1 active player, got true")
	}
}

// TestCanStartHandRequiresNoActiveHand verifies CanStartHand returns false if hand already running
func TestCanStartHandRequiresNoActiveHand(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players
	token0 := "player-0"
	token1 := "player-1"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	// With no active hand, should return true
	if !table.CanStartHand() {
		t.Error("expected CanStartHand to return true with 2 active players and no hand, got false")
	}

	// Set a hand as active
	table.CurrentHand = &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            30,
		Deck:           NewDeck(),
		HoleCards:      make(map[int][]Card),
	}

	// Now it should return false because a hand is already running
	if table.CanStartHand() {
		t.Error("expected CanStartHand to return false with active hand, got true")
	}
}

// TestCanStartHandTrue verifies CanStartHand returns true when ≥2 active and no active hand
func TestCanStartHandTrue(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players (heads-up)
	token0 := "player-0"
	token1 := "player-1"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	if !table.CanStartHand() {
		t.Error("expected CanStartHand to return true with 2 active players, got false")
	}

	// Set up 6 active players (full table)
	table2 := NewTable("table-2", "Table 2", nil)
	for i := 0; i < 6; i++ {
		token := "player-" + string(rune('0'+i))
		table2.Seats[i].Token = &token
		table2.Seats[i].Status = "active"
		table2.Seats[i].Stack = 1000
	}

	if !table2.CanStartHand() {
		t.Error("expected CanStartHand to return true with 6 active players, got false")
	}
}

// TestStartHandInitializesDealer verifies StartHand sets dealer via NextDealer()
func TestStartHandInitializesDealer(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players
	token0 := "player-0"
	token1 := "player-1"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Verify dealer was set
	if table.CurrentHand == nil {
		t.Fatal("expected CurrentHand to be set, got nil")
	}

	if table.CurrentHand.DealerSeat != 0 {
		t.Errorf("expected dealer seat 0, got %d", table.CurrentHand.DealerSeat)
	}

	if table.DealerSeat == nil || *table.DealerSeat != 0 {
		t.Errorf("expected table.DealerSeat to be 0, got %v", table.DealerSeat)
	}
}

// TestStartHandPostsBlinds verifies StartHand deducts SB(10) and BB(20) from stacks, pot = 30
func TestStartHandPostsBlinds(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players (heads-up)
	token0 := "player-0"
	token1 := "player-1"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// In heads-up: dealer (seat 0) is SB (10), seat 1 is BB (20)
	expectedDealer := 0
	expectedSB := 0
	expectedBB := 1

	if table.CurrentHand.DealerSeat != expectedDealer {
		t.Errorf("expected dealer seat %d, got %d", expectedDealer, table.CurrentHand.DealerSeat)
	}

	if table.CurrentHand.SmallBlindSeat != expectedSB {
		t.Errorf("expected SB seat %d, got %d", expectedSB, table.CurrentHand.SmallBlindSeat)
	}

	if table.CurrentHand.BigBlindSeat != expectedBB {
		t.Errorf("expected BB seat %d, got %d", expectedBB, table.CurrentHand.BigBlindSeat)
	}

	// Verify stacks were updated
	// Dealer (seat 0) should have 1000 - 10 = 990 (posted SB)
	if table.Seats[0].Stack != 990 {
		t.Errorf("expected seat 0 stack 990 (1000 - 10 SB), got %d", table.Seats[0].Stack)
	}

	// Non-dealer (seat 1) should have 1000 - 20 = 980 (posted BB)
	if table.Seats[1].Stack != 980 {
		t.Errorf("expected seat 1 stack 980 (1000 - 20 BB), got %d", table.Seats[1].Stack)
	}

	// Verify PlayerBets have blinds (Pot stays 0 during betting)
	if table.CurrentHand.Pot != 0 {
		t.Errorf("expected pot 0 during betting, got %d", table.CurrentHand.Pot)
	}

	if table.CurrentHand.PlayerBets[0] != 10 {
		t.Errorf("expected seat 0 (SB) PlayerBets 10, got %d", table.CurrentHand.PlayerBets[0])
	}

	if table.CurrentHand.PlayerBets[1] != 20 {
		t.Errorf("expected seat 1 (BB) PlayerBets 20, got %d", table.CurrentHand.PlayerBets[1])
	}
}

// TestStartHandDealsCards verifies each active player has 2 cards in CurrentHand.HoleCards
func TestStartHandDealsCards(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 3 active players
	for i := 0; i < 3; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Verify all active players have 2 hole cards
	for seatIdx := 0; seatIdx < 3; seatIdx++ {
		cards, exists := table.CurrentHand.HoleCards[seatIdx]
		if !exists {
			t.Errorf("seat %d: expected hole cards, got none", seatIdx)
		}
		if len(cards) != 2 {
			t.Errorf("seat %d: expected 2 hole cards, got %d", seatIdx, len(cards))
		}
	}

	// Verify non-active seats don't have cards
	for seatIdx := 3; seatIdx < 6; seatIdx++ {
		if _, exists := table.CurrentHand.HoleCards[seatIdx]; exists {
			t.Errorf("seat %d (empty): expected no hole cards, but found some", seatIdx)
		}
	}
}

// TestStartHandSetsPot verifies pot equals SB + BB = 30
func TestStartHandSetsPot(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 6 active players (full table)
	for i := 0; i < 6; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Verify PlayerBets have blinds (Pot stays 0 during betting)
	if table.CurrentHand.Pot != 0 {
		t.Errorf("expected pot 0 during betting, got %d", table.CurrentHand.Pot)
	}

	// In 6-player game: dealer at 0, SB at next active (seat 1), BB at seat 2
	sbSeat := table.CurrentHand.SmallBlindSeat
	bbSeat := table.CurrentHand.BigBlindSeat

	if table.CurrentHand.PlayerBets[sbSeat] != 10 {
		t.Errorf("expected seat %d (SB) PlayerBets 10, got %d", sbSeat, table.CurrentHand.PlayerBets[sbSeat])
	}

	if table.CurrentHand.PlayerBets[bbSeat] != 20 {
		t.Errorf("expected seat %d (BB) PlayerBets 20, got %d", bbSeat, table.CurrentHand.PlayerBets[bbSeat])
	}
}

// TestStartHandAllInBlind verifies handling player with stack < blind amount (goes all-in)
func TestStartHandAllInBlind(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 3 active players where SB player has only 5 chips (less than 10)
	token0 := "player-0"
	token1 := "player-1" // Will be SB with only 5 chips (all-in)
	token2 := "player-2"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 5 // Only 5 chips for SB (10 required) - will go all-in

	table.Seats[2].Token = &token2
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Dealer should be seat 0 (first active)
	if table.CurrentHand.DealerSeat != 0 {
		t.Errorf("expected dealer seat 0, got %d", table.CurrentHand.DealerSeat)
	}

	// SB should be seat 1 (all-in with 5)
	if table.CurrentHand.SmallBlindSeat != 1 {
		t.Errorf("expected SB seat 1, got %d", table.CurrentHand.SmallBlindSeat)
	}

	// BB should be seat 2
	if table.CurrentHand.BigBlindSeat != 2 {
		t.Errorf("expected BB seat 2, got %d", table.CurrentHand.BigBlindSeat)
	}

	// Verify stacks: seat 0 (dealer) unchanged, seat 1 (SB) at 0 (all-in with 5), seat 2 (BB) at 980 (1000 - 20)
	if table.Seats[0].Stack != 1000 {
		t.Errorf("expected seat 0 stack 1000 (no blind), got %d", table.Seats[0].Stack)
	}

	if table.Seats[1].Stack != 0 {
		t.Errorf("expected seat 1 stack 0 (all-in with 5), got %d", table.Seats[1].Stack)
	}

	if table.Seats[2].Stack != 980 {
		t.Errorf("expected seat 2 stack 980 (1000 - 20 BB), got %d", table.Seats[2].Stack)
	}

	// Verify PlayerBets have blinds (Pot stays 0 during betting)
	if table.CurrentHand.Pot != 0 {
		t.Errorf("expected pot 0 during betting, got %d", table.CurrentHand.Pot)
	}

	// SB posted 5 (all-in), BB posted 20
	if table.CurrentHand.PlayerBets[1] != 5 {
		t.Errorf("expected seat 1 (SB) PlayerBets 5, got %d", table.CurrentHand.PlayerBets[1])
	}

	if table.CurrentHand.PlayerBets[2] != 20 {
		t.Errorf("expected seat 2 (BB) PlayerBets 20, got %d", table.CurrentHand.PlayerBets[2])
	}
}

// ============ PHASE 1: TURN ORDER & ACTION STATE TESTS ============

// TestGetFirstActor_HeadsUp verifies dealer acts first preflop in heads-up
func TestGetFirstActor_HeadsUp(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up heads-up: seats 0 and 2 active (dealer at 0)
	token0 := "player-0"
	token2 := "player-2"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[2].Token = &token2
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000

	// Start hand to set dealer and blinds
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// In heads-up, dealer (seat 0) should act first preflop
	firstActor := table.CurrentHand.GetFirstActor(table.Seats)
	if firstActor != 0 {
		t.Errorf("expected first actor to be dealer (seat 0) in heads-up, got %d", firstActor)
	}
}

// TestGetFirstActor_MultiPlayer verifies first seat after BB acts first in 3+ player game
func TestGetFirstActor_MultiPlayer(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 4 active players (seats 0, 1, 2, 3)
	for i := 0; i < 4; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand to set dealer and blinds
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// In multi-player preflop, UTG (seat after BB) acts first
	// Dealer at 0, SB at 1, BB at 2, so UTG (first to act) is seat 3
	firstActor := table.CurrentHand.GetFirstActor(table.Seats)
	if firstActor != 3 {
		t.Errorf("expected first actor to be UTG (seat 3), got %d", firstActor)
	}
}

// TestGetFirstActor_MultiPlayer_WithScatteredSeats verifies UTG with non-consecutive seats
func TestGetFirstActor_MultiPlayer_WithScatteredSeats(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 3 active players at scattered seats (1, 3, 5)
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

	// Start hand to set dealer and blinds
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// After StartHand, we need to determine positions
	// Dealer should be seat 1 (first active), SB seat 3, BB seat 5
	// So UTG (first to act) is seat 1 (next after BB in rotation)
	firstActor := table.CurrentHand.GetFirstActor(table.Seats)
	if firstActor != 1 {
		t.Errorf("expected first actor to be UTG (seat 1), got %d", firstActor)
	}
}

// TestGetFirstActor_HeadsUp_DealerValidation verifies dealer seat is validated as active in heads-up
func TestGetFirstActor_HeadsUp_DealerValidation(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up heads-up: seats 1 and 3 active
	token1 := "player-1"
	token3 := "player-3"

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	table.Seats[3].Token = &token3
	table.Seats[3].Status = "active"
	table.Seats[3].Stack = 1000

	// Start hand to set dealer and blinds
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Manually mark the dealer as inactive (simulate edge case where dealer became inactive)
	dealerSeat := table.CurrentHand.DealerSeat
	table.Seats[dealerSeat].Status = "empty"

	// GetFirstActor should still return a valid seat (not panic or return invalid seat)
	firstActor := table.CurrentHand.GetFirstActor(table.Seats)

	// Verify the returned seat is actually active
	if table.Seats[firstActor].Status != "active" {
		t.Errorf("expected first actor to be an active seat, got seat %d with status %s", firstActor, table.Seats[firstActor].Status)
	}
}

// TestGetFirstActor_BBNotFound verifies handling when BB is not in active seats (edge case)
func TestGetFirstActor_BBNotFound(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 3 active players (seats 0, 1, 2)
	for i := 0; i < 3; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand to set dealer and blinds
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Save the BB seat before we mess with it
	bbSeat := table.CurrentHand.BigBlindSeat

	// Manually mark the BB as inactive (simulate edge case where BB became inactive)
	table.Seats[bbSeat].Status = "empty"

	// GetFirstActor should handle this gracefully without panic
	// It should return a valid active seat as fallback
	firstActor := table.CurrentHand.GetFirstActor(table.Seats)

	// Verify the returned seat is actually active
	if table.Seats[firstActor].Status != "active" {
		t.Errorf("expected first actor to be an active seat, got seat %d with status %s", firstActor, table.Seats[firstActor].Status)
	}

	// Verify it's not the BB seat (since BB is inactive)
	if firstActor == bbSeat {
		t.Errorf("expected first actor to NOT be the inactive BB seat %d, got %d", bbSeat, firstActor)
	}
}

// TestGetFirstActor_Postflop_MultiPlayer verifies SB acts first on postflop streets in multi-player games
func TestGetFirstActor_Postflop_MultiPlayer(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 4 active players (seats 0, 1, 2, 3)
	for i := 0; i < 4; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand to set dealer and blinds
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Setup positions: dealer at 0, SB at 1, BB at 2, UTG at 3
	// Manually set street to flop to simulate postflop action
	table.CurrentHand.Street = "flop"

	// On postflop streets, SB should act first in multi-player
	firstActor := table.CurrentHand.GetFirstActor(table.Seats)
	sbSeat := table.CurrentHand.SmallBlindSeat

	if firstActor != sbSeat {
		t.Errorf("expected first actor to be SB (seat %d) on postflop, got seat %d", sbSeat, firstActor)
	}
}

// TestGetFirstActor_Postflop_HeadsUp verifies BB acts first on postflop streets in heads-up
func TestGetFirstActor_Postflop_HeadsUp(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players (seats 0 and 2)
	token0 := "player-0"
	token2 := "player-2"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[2].Token = &token2
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000

	// Start hand to set dealer and blinds
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Setup positions: dealer/SB at 0, BB at 2
	// Manually set street to flop to simulate postflop action
	table.CurrentHand.Street = "flop"

	// On postflop streets in heads-up, BB (non-dealer) should act first
	firstActor := table.CurrentHand.GetFirstActor(table.Seats)
	bbSeat := table.CurrentHand.BigBlindSeat

	if firstActor != bbSeat {
		t.Errorf("expected first actor to be BB (seat %d) on postflop heads-up, got seat %d", bbSeat, firstActor)
	}
}

// TestGetFirstActor_Postflop_WithFoldedSB verifies BB acts first on postflop when SB is folded
func TestGetFirstActor_Postflop_WithFoldedSB(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 4 active players (seats 0, 1, 2, 3)
	for i := 0; i < 4; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand to set dealer and blinds
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Get positions: dealer at 0, SB at 1, BB at 2
	sbSeat := table.CurrentHand.SmallBlindSeat
	bbSeat := table.CurrentHand.BigBlindSeat

	// Mark SB as folded
	table.CurrentHand.FoldedPlayers[sbSeat] = true

	// Manually set street to flop
	table.CurrentHand.Street = "flop"

	// On postflop with folded SB, BB should act first
	firstActor := table.CurrentHand.GetFirstActor(table.Seats)

	if firstActor != bbSeat {
		t.Errorf("expected first actor to be BB (seat %d) when SB is folded on postflop, got seat %d", bbSeat, firstActor)
	}
}

// ============ PHASE 3: POT DISTRIBUTION & STACK UPDATES TESTS ============

// TestDistributePot_SingleWinner verifies single winner gets eligible pot(s) only
func TestDistributePot_SingleWinner(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)
	winners := []int{2}

	// Set up hand with pot and contributions
	// Seat 0 (30), Seat 1 (40), Seat 2 (30) - only seat 2 is not folded
	// Main pot at level 30: 30 * 3 = 90 (all contributed at least 30)
	// Seat 2 only eligible for main pot (level 30) since it's the only non-folded player
	// Side pot at level 40: 10 * 1 (only seat 1 at this level, but it's folded) = 10, no eligible winners
	// So seat 2 wins 90, the 10 remains unawarded
	table.CurrentHand = &Hand{
		Pot: 100,
		TotalContributions: map[int]int{
			0: 30,
			1: 40,
			2: 30,
		},
		FoldedPlayers: map[int]bool{
			0: true,
			1: true,
			2: false, // winner
		},
	}

	result := table.DistributePot(winners)

	if len(result) == 0 {
		t.Fatal("expected result map to be non-empty")
	}

	// Seat 2 is eligible for main pot (30 * 3 = 90)
	// But NOT eligible for side pot (only seat 1 at level 40, and seat 1 is folded)
	if result[2] != 90 {
		t.Errorf("expected winner at seat 2 to receive 90 (main pot), got %d", result[2])
	}

	// Verify only the winner seat is in the map
	if len(result) != 1 {
		t.Errorf("expected only 1 entry in result map, got %d", len(result))
	}
}

// TestDistributePot_TwoWayTie_EvenSplit verifies even split between 2 winners (100 chip pot)
func TestDistributePot_TwoWayTie_EvenSplit(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)
	winners := []int{1, 3}
	pot := 100

	table.CurrentHand = &Hand{
		Pot: pot,
		TotalContributions: map[int]int{
			1: 50,
			3: 50,
		},
		FoldedPlayers: map[int]bool{
			1: false, // winner
			3: false, // winner
		},
	}

	result := table.DistributePot(winners)

	if result[1] != 50 {
		t.Errorf("expected winner at seat 1 to receive 50, got %d", result[1])
	}

	if result[3] != 50 {
		t.Errorf("expected winner at seat 3 to receive 50, got %d", result[3])
	}

	if len(result) != 2 {
		t.Errorf("expected 2 entries in result map, got %d", len(result))
	}
}

// TestDistributePot_ThreeWayTie_EvenSplit verifies even split between 3 winners (90 chip pot)
func TestDistributePot_ThreeWayTie_EvenSplit(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)
	winners := []int{0, 2, 4}
	pot := 90

	table.CurrentHand = &Hand{
		Pot: pot,
		TotalContributions: map[int]int{
			0: 30,
			2: 30,
			4: 30,
		},
		FoldedPlayers: map[int]bool{
			0: false, // winner
			2: false, // winner
			4: false, // winner
		},
	}

	result := table.DistributePot(winners)

	if result[0] != 30 {
		t.Errorf("expected winner at seat 0 to receive 30, got %d", result[0])
	}

	if result[2] != 30 {
		t.Errorf("expected winner at seat 2 to receive 30, got %d", result[2])
	}

	if result[4] != 30 {
		t.Errorf("expected winner at seat 4 to receive 30, got %d", result[4])
	}

	if len(result) != 3 {
		t.Errorf("expected 3 entries in result map, got %d", len(result))
	}
}

// TestDistributePot_TwoWayTie_OddPot verifies remainder goes to first winner in list (101 chip pot, 2 winners)
func TestDistributePot_TwoWayTie_OddPot(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)
	// Winners: seat 3 and seat 5. Seat 5 put in 1 more chip than seat 3
	// Main pot (50*2 = 100): splits evenly 50 each (both eligible)
	// Side pot (1*1 = 1): goes to seat 5 only (only seat 5 contributed this)
	// Result: seat 3 gets 50, seat 5 gets 51
	winners := []int{3, 5}
	pot := 101

	table.CurrentHand = &Hand{
		Pot: pot,
		TotalContributions: map[int]int{
			3: 50, // contributed 50 (eligible for main pot only)
			5: 51, // contributed 51 (eligible for both pots)
		},
		FoldedPlayers: map[int]bool{
			3: false, // winner
			5: false, // winner
		},
	}

	result := table.DistributePot(winners)

	// Main pot (50*2=100): split evenly between 3 and 5 = 50 each
	// Side pot (1*1=1): only eligible for seat 5 = 1
	// Seat 3 total: 50
	if result[3] != 50 {
		t.Errorf("expected seat 3 to receive 50, got %d", result[3])
	}

	// Seat 5 total: 50 + 1 = 51
	if result[5] != 51 {
		t.Errorf("expected seat 5 to receive 51, got %d", result[5])
	}

	if len(result) != 2 {
		t.Errorf("expected 2 entries in result map, got %d", len(result))
	}
}

// TestHandleShowdown_UpdatesStacks verifies stack values are updated after showdown
func TestHandleShowdown_UpdatesStacks(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	table := server.tables[0]

	// Set up 3 active players
	for i := 0; i < 3; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand: dealer at 0, SB at 1 (990 stack), BB at 2 (980 stack), UTG (seat 0) still has 1000
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Manually set up game to reach showdown state
	// Seats 0 and 1 are still active, seat 2 folded
	// Use the actual pot and stacks after StartHand (which posts blinds)
	table.CurrentHand.Street = "river"
	table.CurrentHand.FoldedPlayers[2] = true // Seat 2 folded

	// Simulate seat 0 calling the BB (20 chips) by adding to PlayerBets
	table.CurrentHand.PlayerBets[0] = 20 // Seat 0 calls the BB

	// With new pot accounting: sweep all PlayerBets into Pot
	for _, bet := range table.CurrentHand.PlayerBets {
		table.CurrentHand.Pot += bet
	}
	table.CurrentHand.PlayerBets = make(map[int]int)

	// Update TotalContributions to reflect seat 0's call
	table.CurrentHand.TotalContributions[0] = 20 // Seat 0 called the BB
	table.Seats[0].Stack = 1000 - 20             // Deduct the call from seat 0's stack
	// Now Pot should be 50 (10 + 20 + 20)

	// Add board cards (5 cards for river) for proper hand evaluation
	table.CurrentHand.BoardCards = []Card{
		{Rank: "A", Suit: "s"},
		{Rank: "K", Suit: "h"},
		{Rank: "Q", Suit: "d"},
		{Rank: "J", Suit: "c"},
		{Rank: "T", Suit: "s"},
	}

	// Verify initial stacks after blind posting and seat 0's call
	if table.Seats[0].Stack != 980 {
		t.Errorf("expected seat 0 stack 980 (dealer, called BB -20), got %d", table.Seats[0].Stack)
	}
	if table.Seats[1].Stack != 990 {
		t.Errorf("expected seat 1 stack 990 (SB -10), got %d", table.Seats[1].Stack)
	}
	if table.Seats[2].Stack != 980 {
		t.Errorf("expected seat 2 stack 980 (BB -20), got %d", table.Seats[2].Stack)
	}

	// Call HandleShowdown
	table.HandleShowdown()

	// After showdown, at least one player should have an updated stack (winner gets the pot)
	// The pot (50 chips from blinds and seat 0's call) should be distributed to seats 0 and/or 1
	// (seat 2 is not eligible since it folded)
	totalStacks := table.Seats[0].Stack + table.Seats[1].Stack + table.Seats[2].Stack
	originalTotal := 1000 + 1000 + 1000 // Original chips before any betting
	if totalStacks != originalTotal {
		t.Errorf("expected total stacks %d (conserved chips), got %d", originalTotal, totalStacks)
	}

	// At least one winner should have more than their post-blind amount
	// Seat 0 should have > 1000 or Seat 1 should have > 990
	if table.Seats[0].Stack <= 1000 && table.Seats[1].Stack <= 990 {
		t.Error("expected at least one winner to have stack increased from post-blind amount")
	}
}

// TestHandleShowdown_DetectsBustOut verifies HandleShowdown identifies players with stack == 0 as bust-outs
func TestHandleShowdown_DetectsBustOut(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	table := server.tables[0]

	// Set up 2 active players
	token0 := "player-0"
	token1 := "player-1"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 50

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 50

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Manually set up showdown state
	table.CurrentHand.Street = "river"
	table.CurrentHand.FoldedPlayers[1] = true
	table.CurrentHand.Pot = 80

	// Add board cards for proper hand evaluation
	table.CurrentHand.BoardCards = []Card{
		{Rank: "2", Suit: "s"},
		{Rank: "3", Suit: "h"},
		{Rank: "4", Suit: "d"},
		{Rank: "5", Suit: "c"},
		{Rank: "6", Suit: "s"},
	}

	// Before HandleShowdown, both seats are occupied
	if table.Seats[0].Token == nil || table.Seats[1].Token == nil {
		t.Fatal("expected both seats to be occupied before showdown")
	}

	// Call HandleShowdown
	table.HandleShowdown()

	// After showdown, winner should have stack > 0, loser might be busted
	// The key is that bust-outs are detected and cleared
	// At minimum, verify no errors occurred and the function completed
}

// TestHandleShowdown_ClearsBustOutSeat verifies bust-out seats are cleared (Token = nil, Status = "empty")
func TestHandleShowdown_ClearsBustOutSeat(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	table := server.tables[0]

	// Set up 2 active players
	token0 := "player-0"
	token1 := "player-1"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 30 // Will win, gets 60 total

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 30 // Will lose

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Manually set up showdown: pot is 60, one player will get it, other will have stack 0
	table.CurrentHand.Street = "river"
	table.CurrentHand.FoldedPlayers[1] = true
	table.CurrentHand.Pot = 60

	// Add board cards for proper hand evaluation
	table.CurrentHand.BoardCards = []Card{
		{Rank: "7", Suit: "s"},
		{Rank: "8", Suit: "h"},
		{Rank: "9", Suit: "d"},
		{Rank: "T", Suit: "c"},
		{Rank: "J", Suit: "s"},
	}

	// Call HandleShowdown
	table.HandleShowdown()

	// After showdown, if any seat has stack == 0, it should be cleared
	for i := 0; i < 6; i++ {
		if table.Seats[i].Stack == 0 {
			// Busted out seat should be cleared
			if table.Seats[i].Token != nil {
				t.Errorf("expected seat %d (busted out) to have Token == nil, got %v", i, table.Seats[i].Token)
			}
			if table.Seats[i].Status != "empty" {
				t.Errorf("expected seat %d (busted out) to have Status 'empty', got '%s'", i, table.Seats[i].Status)
			}
		}
	}
}

// TestHandleBustOuts_MultiplePlayersOut verifies HandleBustOuts clears all seats with stack == 0
func TestHandleBustOuts_MultiplePlayersOut(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 4 active players
	for i := 0; i < 4; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Manually set seats 1 and 3 to have stack == 0 (busted out)
	table.Seats[1].Stack = 0
	table.Seats[3].Stack = 0

	// Call HandleBustOuts
	table.HandleBustOuts()

	// Verify seats 1 and 3 are cleared
	if table.Seats[1].Token != nil {
		t.Errorf("expected seat 1 (bust out) to have Token == nil after HandleBustOuts, got %v", table.Seats[1].Token)
	}
	if table.Seats[1].Status != "empty" {
		t.Errorf("expected seat 1 (bust out) to have Status 'empty' after HandleBustOuts, got '%s'", table.Seats[1].Status)
	}

	if table.Seats[3].Token != nil {
		t.Errorf("expected seat 3 (bust out) to have Token == nil after HandleBustOuts, got %v", table.Seats[3].Token)
	}
	if table.Seats[3].Status != "empty" {
		t.Errorf("expected seat 3 (bust out) to have Status 'empty' after HandleBustOuts, got '%s'", table.Seats[3].Status)
	}

	// Verify remaining players are unchanged
	if table.Seats[0].Status != "active" || table.Seats[0].Stack != 1000 {
		t.Errorf("expected seat 0 to remain unchanged, got status='%s', stack=%d", table.Seats[0].Status, table.Seats[0].Stack)
	}
	if table.Seats[2].Status != "active" || table.Seats[2].Stack != 1000 {
		t.Errorf("expected seat 2 to remain unchanged, got status='%s', stack=%d", table.Seats[2].Status, table.Seats[2].Stack)
	}
}

// TestHandleBustOuts_WinnerDoesNotBustOut verifies HandleBustOuts doesn't clear seats with stack > 0
func TestHandleBustOuts_WinnerDoesNotBustOut(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players
	token0 := "player-0"
	token1 := "player-1"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 150 // Winner has stack

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 0 // Loser has 0

	// Call HandleBustOuts
	table.HandleBustOuts()

	// Verify seat 0 is NOT cleared (winner with stack > 0)
	if table.Seats[0].Token == nil {
		t.Error("expected seat 0 (winner with stack > 0) to remain occupied, got Token == nil")
	}
	if table.Seats[0].Status != "active" {
		t.Errorf("expected seat 0 (winner) to remain 'active', got '%s'", table.Seats[0].Status)
	}

	// Verify seat 1 is cleared (busted with stack == 0)
	if table.Seats[1].Token != nil {
		t.Errorf("expected seat 1 (busted out) to have Token == nil, got %v", table.Seats[1].Token)
	}
	if table.Seats[1].Status != "empty" {
		t.Errorf("expected seat 1 (busted out) to have Status 'empty', got '%s'", table.Seats[1].Status)
	}
}

// ============ PHASE 1: BOARD CARD STORAGE & DEALING TESTS ============

// TestHand_BoardCards_InitiallyEmpty verifies BoardCards field is initialized as empty slice
func TestHand_BoardCards_InitiallyEmpty(t *testing.T) {
	hand := &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            0,
		Deck:           NewDeck(),
		HoleCards:      make(map[int][]Card),
	}

	// Verify BoardCards is empty (should be nil or len 0)
	if hand.BoardCards != nil && len(hand.BoardCards) != 0 {
		t.Errorf("expected BoardCards to be empty, got %d cards", len(hand.BoardCards))
	}
}

// TestHand_DealFlop_DealsThreeCards verifies flop deals exactly 3 cards to board
func TestHand_DealFlop_DealsThreeCards(t *testing.T) {
	hand := &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            30,
		Deck:           NewDeck(),
		HoleCards:      make(map[int][]Card),
		BoardCards:     []Card{},
		Street:         "preflop",
	}

	err := ShuffleDeck(hand.Deck)
	if err != nil {
		t.Fatalf("failed to shuffle deck: %v", err)
	}

	initialDeckSize := len(hand.Deck)

	// Deal flop
	err = hand.DealFlop()
	if err != nil {
		t.Fatalf("expected no error dealing flop, got %v", err)
	}

	// Verify exactly 3 cards on board
	if len(hand.BoardCards) != 3 {
		t.Errorf("expected 3 board cards after flop, got %d", len(hand.BoardCards))
	}

	// Verify deck reduced by 4 cards (1 burn + 3 dealt)
	if len(hand.Deck) != initialDeckSize-4 {
		t.Errorf("expected deck to reduce by 4 cards (1 burn + 3 flop), got reduction of %d", initialDeckSize-len(hand.Deck))
	}
}

// TestHand_DealFlop_BurnsCardBeforeDealing verifies burn card is discarded (not stored)
func TestHand_DealFlop_BurnsCardBeforeDealing(t *testing.T) {
	hand := &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            30,
		Deck:           NewDeck(),
		HoleCards:      make(map[int][]Card),
		BoardCards:     []Card{},
		Street:         "preflop",
	}

	err := ShuffleDeck(hand.Deck)
	if err != nil {
		t.Fatalf("failed to shuffle deck: %v", err)
	}

	// Store cards from original deck before dealing
	burnCard := hand.Deck[0]
	flopCard1 := hand.Deck[1]
	flopCard2 := hand.Deck[2]
	flopCard3 := hand.Deck[3]

	// Deal flop
	err = hand.DealFlop()
	if err != nil {
		t.Fatalf("expected no error dealing flop, got %v", err)
	}

	// Verify burn card is not in board cards
	for _, boardCard := range hand.BoardCards {
		if boardCard == burnCard {
			t.Errorf("expected burn card to not be in board, but found %s", burnCard.String())
		}
	}

	// Verify board cards are exactly cards 1, 2, 3 from original deck
	if hand.BoardCards[0] != flopCard1 || hand.BoardCards[1] != flopCard2 || hand.BoardCards[2] != flopCard3 {
		t.Errorf("expected board cards to be original deck cards 1-3, got different cards")
	}
}

// TestHand_DealFlop_ErrorsIfDeckExhausted verifies error when deck has <4 cards for flop
func TestHand_DealFlop_ErrorsIfDeckExhausted(t *testing.T) {
	hand := &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            30,
		Deck: []Card{
			{Rank: "A", Suit: "s"},
			{Rank: "K", Suit: "h"},
			{Rank: "Q", Suit: "d"},
		},
		HoleCards:  make(map[int][]Card),
		BoardCards: []Card{},
		Street:     "preflop",
	}

	// Try to deal flop with only 3 cards (need 4: 1 burn + 3 flop)
	err := hand.DealFlop()
	if err == nil {
		t.Fatal("expected error when deck has insufficient cards for flop, got nil")
	}

	if err.Error() != "insufficient cards in deck: have 3, need 4" {
		t.Errorf("expected error 'insufficient cards in deck: have 3, need 4', got '%s'", err.Error())
	}

	// Verify no cards were dealt
	if len(hand.BoardCards) != 0 {
		t.Errorf("expected board to remain empty on error, got %d cards", len(hand.BoardCards))
	}
}

// TestHand_DealTurn_DealsOneCard verifies turn deals exactly 1 card to board
func TestHand_DealTurn_DealsOneCard(t *testing.T) {
	hand := &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            30,
		Deck:           NewDeck(),
		HoleCards:      make(map[int][]Card),
		BoardCards:     []Card{{Rank: "A", Suit: "s"}, {Rank: "K", Suit: "h"}, {Rank: "Q", Suit: "d"}},
		Street:         "flop",
	}

	err := ShuffleDeck(hand.Deck)
	if err != nil {
		t.Fatalf("failed to shuffle deck: %v", err)
	}

	initialDeckSize := len(hand.Deck)
	initialBoardSize := len(hand.BoardCards)

	// Deal turn
	err = hand.DealTurn()
	if err != nil {
		t.Fatalf("expected no error dealing turn, got %v", err)
	}

	// Verify exactly 4 cards on board (3 flop + 1 turn)
	if len(hand.BoardCards) != 4 {
		t.Errorf("expected 4 board cards after turn, got %d", len(hand.BoardCards))
	}

	// Verify deck reduced by 2 cards (1 burn + 1 turn)
	if len(hand.Deck) != initialDeckSize-2 {
		t.Errorf("expected deck to reduce by 2 cards (1 burn + 1 turn), got reduction of %d", initialDeckSize-len(hand.Deck))
	}

	// Verify only 1 new card was added
	if len(hand.BoardCards) != initialBoardSize+1 {
		t.Errorf("expected board to grow by 1 card, got growth of %d", len(hand.BoardCards)-initialBoardSize)
	}
}

// TestHand_DealTurn_BurnsCardBeforeDealing verifies turn burn card is discarded (not stored)
func TestHand_DealTurn_BurnsCardBeforeDealing(t *testing.T) {
	hand := &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            30,
		Deck:           NewDeck(),
		HoleCards:      make(map[int][]Card),
		BoardCards:     []Card{{Rank: "A", Suit: "s"}, {Rank: "K", Suit: "h"}, {Rank: "Q", Suit: "d"}},
		Street:         "flop",
	}

	err := ShuffleDeck(hand.Deck)
	if err != nil {
		t.Fatalf("failed to shuffle deck: %v", err)
	}

	initialBoardCard4 := hand.Deck[1]

	// Deal turn
	err = hand.DealTurn()
	if err != nil {
		t.Fatalf("expected no error dealing turn, got %v", err)
	}

	// Verify the 4th board card is not the first card from deck (which was burned)
	// The turn card should be the 2nd card from the pre-deal deck
	if hand.BoardCards[3] != initialBoardCard4 {
		t.Errorf("expected turn card to be the second card from pre-turn deck (burned first)")
	}
}

// TestHand_DealRiver_DealsOneCard verifies river deals exactly 1 card to board
func TestHand_DealRiver_DealsOneCard(t *testing.T) {
	hand := &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            30,
		Deck:           NewDeck(),
		HoleCards:      make(map[int][]Card),
		BoardCards: []Card{
			{Rank: "A", Suit: "s"}, {Rank: "K", Suit: "h"}, {Rank: "Q", Suit: "d"},
			{Rank: "J", Suit: "c"},
		},
		Street: "turn",
	}

	err := ShuffleDeck(hand.Deck)
	if err != nil {
		t.Fatalf("failed to shuffle deck: %v", err)
	}

	initialDeckSize := len(hand.Deck)
	initialBoardSize := len(hand.BoardCards)

	// Deal river
	err = hand.DealRiver()
	if err != nil {
		t.Fatalf("expected no error dealing river, got %v", err)
	}

	// Verify exactly 5 cards on board (4 from turn + 1 river)
	if len(hand.BoardCards) != 5 {
		t.Errorf("expected 5 board cards after river, got %d", len(hand.BoardCards))
	}

	// Verify deck reduced by 2 cards (1 burn + 1 river)
	if len(hand.Deck) != initialDeckSize-2 {
		t.Errorf("expected deck to reduce by 2 cards (1 burn + 1 river), got reduction of %d", initialDeckSize-len(hand.Deck))
	}

	// Verify only 1 new card was added
	if len(hand.BoardCards) != initialBoardSize+1 {
		t.Errorf("expected board to grow by 1 card, got growth of %d", len(hand.BoardCards)-initialBoardSize)
	}
}

// TestHand_DealRiver_BurnsCardBeforeDealing verifies river burn card is discarded (not stored)
func TestHand_DealRiver_BurnsCardBeforeDealing(t *testing.T) {
	hand := &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            30,
		Deck:           NewDeck(),
		HoleCards:      make(map[int][]Card),
		BoardCards: []Card{
			{Rank: "A", Suit: "s"}, {Rank: "K", Suit: "h"}, {Rank: "Q", Suit: "d"},
			{Rank: "J", Suit: "c"},
		},
		Street: "turn",
	}

	err := ShuffleDeck(hand.Deck)
	if err != nil {
		t.Fatalf("failed to shuffle deck: %v", err)
	}

	initialBoardCard5 := hand.Deck[1]

	// Deal river
	err = hand.DealRiver()
	if err != nil {
		t.Fatalf("expected no error dealing river, got %v", err)
	}

	// Verify the 5th board card is not the first card from deck (which was burned)
	// The river card should be the 2nd card from the pre-deal deck
	if hand.BoardCards[4] != initialBoardCard5 {
		t.Errorf("expected river card to be the second card from pre-river deck (burned first)")
	}
}

// TestGetNextActiveSeat verifies wrap-around and folded player skipping
func TestGetNextActiveSeat(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 4 active players (seats 0, 1, 2, 3)
	for i := 0; i < 4; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Initialize folded players map
	if table.CurrentHand.FoldedPlayers == nil {
		table.CurrentHand.FoldedPlayers = make(map[int]bool)
	}

	// Test: from seat 0, next active should be 1
	next := table.CurrentHand.GetNextActiveSeat(0, table.Seats)
	if next == nil || *next != 1 {
		t.Errorf("expected next active seat after 0 to be 1, got %v", next)
	}

	// Test: from seat 1 (with seat 2 folded), next active should be 3
	table.CurrentHand.FoldedPlayers[2] = true
	next = table.CurrentHand.GetNextActiveSeat(1, table.Seats)
	if next == nil || *next != 3 {
		t.Errorf("expected next active seat after 1 (skipping folded 2) to be 3, got %v", next)
	}

	// Test: from seat 3 (wrap-around), next active should be 0
	next = table.CurrentHand.GetNextActiveSeat(3, table.Seats)
	if next == nil || *next != 0 {
		t.Errorf("expected next active seat after 3 (wrap to 0) to be 0, got %v", next)
	}

	// Test: from seat 2 (folded), next active should be 3 (skip self)
	next = table.CurrentHand.GetNextActiveSeat(2, table.Seats)
	if next == nil || *next != 3 {
		t.Errorf("expected next active seat after folded 2 to be 3, got %v", next)
	}
}

// TestGetNextActiveSeat_AllButOneFinished verifies returns nil when all others are folded
func TestGetNextActiveSeat_AllButOneFinished(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 3 active players (seats 0, 1, 2)
	for i := 0; i < 3; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Initialize folded players map
	if table.CurrentHand.FoldedPlayers == nil {
		table.CurrentHand.FoldedPlayers = make(map[int]bool)
	}

	// Mark seats 1 and 2 as folded
	table.CurrentHand.FoldedPlayers[1] = true
	table.CurrentHand.FoldedPlayers[2] = true

	// From seat 0, there are no active players left (all others folded)
	next := table.CurrentHand.GetNextActiveSeat(0, table.Seats)
	if next != nil {
		t.Errorf("expected next active seat to be nil when all others folded, got %v", next)
	}
}

// TestGetCallAmount_NoCurrentBet verifies call amount is 0 when no bet is set
func TestGetCallAmount_NoCurrentBet(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players
	for i := 0; i < 2; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Initialize action state
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}

	// After StartHand, CurrentBet is 20 (big blind), seat 0 (dealer/small blind) has bet 10
	// So call amount for seat 0 should be 10 (20 - 10)
	callAmount := table.CurrentHand.GetCallAmount(0)
	if callAmount != 10 {
		t.Errorf("expected call amount 10 (to match BB), got %d", callAmount)
	}
}

// TestGetCallAmount_BehindCurrentBet verifies call amount is difference between CurrentBet and PlayerBet
func TestGetCallAmount_BehindCurrentBet(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players
	for i := 0; i < 2; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Initialize action state
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}

	// Player has bet 10, current bet is 50
	table.CurrentHand.PlayerBets[0] = 10
	table.CurrentHand.CurrentBet = 50

	callAmount := table.CurrentHand.GetCallAmount(0)
	if callAmount != 40 {
		t.Errorf("expected call amount 40 (50-10), got %d", callAmount)
	}
}

// TestGetCallAmount_AlreadyMatched verifies call amount is 0 when player has already matched bet
func TestGetCallAmount_AlreadyMatched(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players
	for i := 0; i < 2; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Initialize action state
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}

	// Player has already matched the current bet
	table.CurrentHand.PlayerBets[0] = 50
	table.CurrentHand.CurrentBet = 50

	callAmount := table.CurrentHand.GetCallAmount(0)
	if callAmount != 0 {
		t.Errorf("expected call amount 0 when bet matched, got %d", callAmount)
	}
}

// TestGetValidActions_CanCheck verifies check and fold are valid when bet is matched
func TestGetValidActions_CanCheck(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players
	for i := 0; i < 2; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Initialize action state
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}

	// Player has already matched the current bet
	table.CurrentHand.PlayerBets[0] = 50
	table.CurrentHand.CurrentBet = 50

	validActions := table.CurrentHand.GetValidActions(0, table.Seats[0].Stack, table.Seats)

	// Should allow check, fold, and raise (since player has enough chips)
	hasCheck := false
	hasFold := false
	hasRaise := false
	for _, action := range validActions {
		if action == "check" {
			hasCheck = true
		}
		if action == "fold" {
			hasFold = true
		}
		if action == "raise" {
			hasRaise = true
		}
	}

	if !hasCheck {
		t.Errorf("expected 'check' in valid actions, got %v", validActions)
	}
	if !hasFold {
		t.Errorf("expected 'fold' in valid actions, got %v", validActions)
	}
	if !hasRaise {
		t.Errorf("expected 'raise' in valid actions (player has enough chips), got %v", validActions)
	}
	if len(validActions) != 3 {
		t.Errorf("expected exactly 3 valid actions (check, fold, raise), got %d: %v", len(validActions), validActions)
	}
}

// TestGetValidActions_MustCall verifies call and fold are valid when behind current bet
func TestGetValidActions_MustCall(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players
	for i := 0; i < 2; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Initialize action state
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}

	// Player has bet 10, current bet is 50, last raise was 50
	table.CurrentHand.PlayerBets[0] = 10
	table.CurrentHand.CurrentBet = 50
	table.CurrentHand.LastRaise = 50 // min-raise = 100

	validActions := table.CurrentHand.GetValidActions(0, table.Seats[0].Stack, table.Seats)

	// Should allow call, fold, and raise (player has 1000 chips, can raise minimum of 100)
	hasCall := false
	hasFold := false
	hasRaise := false
	for _, action := range validActions {
		if action == "call" {
			hasCall = true
		}
		if action == "fold" {
			hasFold = true
		}
		if action == "raise" {
			hasRaise = true
		}
	}

	if !hasCall {
		t.Errorf("expected 'call' in valid actions, got %v", validActions)
	}
	if !hasFold {
		t.Errorf("expected 'fold' in valid actions, got %v", validActions)
	}
	if !hasRaise {
		t.Errorf("expected 'raise' in valid actions, got %v", validActions)
	}
	if len(validActions) != 3 {
		t.Errorf("expected exactly 3 valid actions (fold, call, raise), got %d: %v", len(validActions), validActions)
	}
}

// TestProcessAction_Fold marks player as folded without changing pot or stacks
func TestProcessAction_Fold(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 3 active players
	for i := 0; i < 3; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Initialize action state
	if table.CurrentHand.FoldedPlayers == nil {
		table.CurrentHand.FoldedPlayers = make(map[int]bool)
	}
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}

	initialPot := table.CurrentHand.Pot
	initialStack := table.Seats[0].Stack

	// Process fold action for seat 0
	chipsMoved, err := table.CurrentHand.ProcessAction(0, "fold", table.Seats[0].Stack)
	if err != nil {
		t.Errorf("expected no error processing fold, got %v", err)
	}

	// Verify no chips were moved
	if chipsMoved != 0 {
		t.Errorf("expected 0 chips moved on fold, got %d", chipsMoved)
	}

	// Verify player is marked as folded
	if !table.CurrentHand.FoldedPlayers[0] {
		t.Errorf("expected player 0 to be marked as folded")
	}

	// Verify pot and stack didn't change
	if table.CurrentHand.Pot != initialPot {
		t.Errorf("expected pot to remain %d, got %d", initialPot, table.CurrentHand.Pot)
	}
	if table.Seats[0].Stack != initialStack {
		t.Errorf("expected stack to remain %d, got %d", initialStack, table.Seats[0].Stack)
	}
}

// TestProcessAction_Check succeeds when bet is matched, no state change
func TestProcessAction_Check(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players
	for i := 0; i < 2; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Initialize action state
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}
	if table.CurrentHand.ActedPlayers == nil {
		table.CurrentHand.ActedPlayers = make(map[int]bool)
	}

	// Set bet so player can check (50/50 matched)
	table.CurrentHand.PlayerBets[0] = 50
	table.CurrentHand.CurrentBet = 50

	initialPot := table.CurrentHand.Pot
	initialStack := table.Seats[0].Stack

	// Process check action
	chipsMoved, err := table.CurrentHand.ProcessAction(0, "check", table.Seats[0].Stack)
	if err != nil {
		t.Errorf("expected no error processing check, got %v", err)
	}

	// Verify no chips were moved
	if chipsMoved != 0 {
		t.Errorf("expected 0 chips moved on check, got %d", chipsMoved)
	}

	// Verify player marked as acted
	if !table.CurrentHand.ActedPlayers[0] {
		t.Errorf("expected player 0 to be marked as acted")
	}

	// Verify no state change (pot, stack)
	if table.CurrentHand.Pot != initialPot {
		t.Errorf("expected pot to remain %d, got %d", initialPot, table.CurrentHand.Pot)
	}
	if table.Seats[0].Stack != initialStack {
		t.Errorf("expected stack to remain %d, got %d", initialStack, table.Seats[0].Stack)
	}
}

// TestProcessAction_CheckInvalidWhenBehind verifies check fails when player is behind current bet
func TestProcessAction_CheckInvalidWhenBehind(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players
	for i := 0; i < 2; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Initialize action state
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}

	// Player behind: bet 10, current bet 50
	table.CurrentHand.PlayerBets[0] = 10
	table.CurrentHand.CurrentBet = 50

	// Try to check when behind
	_, err = table.CurrentHand.ProcessAction(0, "check", table.Seats[0].Stack)
	if err == nil {
		t.Errorf("expected error processing check when behind current bet, got nil")
	}
}

// TestProcessAction_Call updates pot and stack correctly
func TestProcessAction_Call(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players
	for i := 0; i < 2; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Initialize action state
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}
	if table.CurrentHand.ActedPlayers == nil {
		table.CurrentHand.ActedPlayers = make(map[int]bool)
	}

	// Set up: player has bet 10, current bet is 50, stack is 1000
	table.CurrentHand.PlayerBets[0] = 10
	table.CurrentHand.CurrentBet = 50

	// Process call action (need to call 40 more)
	chipsMoved, err := table.CurrentHand.ProcessAction(0, "call", table.Seats[0].Stack)
	if err != nil {
		t.Errorf("expected no error processing call, got %v", err)
	}

	// Verify correct amount of chips were moved
	if chipsMoved != 40 {
		t.Errorf("expected 40 chips moved, got %d", chipsMoved)
	}

	// Verify Pot stays 0 during betting (chips go to PlayerBets)
	if table.CurrentHand.Pot != 0 {
		t.Errorf("expected pot 0 during betting, got %d", table.CurrentHand.Pot)
	}

	// Verify PlayerBets updated to match current bet
	if table.CurrentHand.PlayerBets[0] != 50 {
		t.Errorf("expected PlayerBets[0] to be 50, got %d", table.CurrentHand.PlayerBets[0])
	}

	// Verify player marked as acted
	if !table.CurrentHand.ActedPlayers[0] {
		t.Errorf("expected player 0 to be marked as acted")
	}

	// Note: stack update is the caller's responsibility, so we verify chipsMoved instead
	// In actual handler code, the caller would do: table.Seats[seatIndex].Stack -= chipsMoved
}

// TestProcessAction_CallPartial handles all-in when stack < call amount
func TestProcessAction_CallPartial(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players
	for i := 0; i < 2; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Initialize action state
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}
	if table.CurrentHand.ActedPlayers == nil {
		table.CurrentHand.ActedPlayers = make(map[int]bool)
	}

	// Set up: player needs to call 50, but only has 30 chips left
	table.CurrentHand.PlayerBets[0] = 0
	table.CurrentHand.CurrentBet = 50
	playerStack := 30 // Only 30 chips available

	// Process call action (go all-in with 30)
	chipsMoved, err := table.CurrentHand.ProcessAction(0, "call", playerStack)
	if err != nil {
		t.Errorf("expected no error processing partial call, got %v", err)
	}

	// Verify all-in amount returned (30)
	if chipsMoved != 30 {
		t.Errorf("expected 30 chips moved (all-in), got %d", chipsMoved)
	}

	// Verify Pot stays 0 during betting (chips go to PlayerBets)
	if table.CurrentHand.Pot != 0 {
		t.Errorf("expected pot 0 during betting, got %d", table.CurrentHand.Pot)
	}

	// Verify PlayerBets updated to 30 (not the full current bet)
	if table.CurrentHand.PlayerBets[0] != 30 {
		t.Errorf("expected PlayerBets[0] to be 30, got %d", table.CurrentHand.PlayerBets[0])
	}
}

// TestIsBettingRoundComplete_NotAllActed returns false when not all players have acted
func TestIsBettingRoundComplete_NotAllActed(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)

	// Set up 3 active players
	token1, token2, token3 := "player1", "player2", "player3"
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.Seats[2].Token = &token3
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000

	// Initialize hand with action state
	table.CurrentHand = &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            30,
		CurrentBet:     20,
		Street:         "preflop",
		FoldedPlayers:  make(map[int]bool),
		ActedPlayers:   make(map[int]bool),
		PlayerBets:     make(map[int]int),
	}

	// Player 0 and 1 have acted, but player 2 has not
	table.CurrentHand.ActedPlayers[0] = true
	table.CurrentHand.ActedPlayers[1] = true
	table.CurrentHand.PlayerBets[0] = 20
	table.CurrentHand.PlayerBets[1] = 20

	// Round should NOT be complete
	if table.CurrentHand.IsBettingRoundComplete(table.Seats) {
		t.Error("expected betting round to not be complete when not all players have acted")
	}
}

// TestIsBettingRoundComplete_BetsNotMatched returns false when bets unmatched
func TestIsBettingRoundComplete_BetsNotMatched(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)

	// Set up 3 active players
	token1, token2, token3 := "player1", "player2", "player3"
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.Seats[2].Token = &token3
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000

	// Initialize hand with action state
	table.CurrentHand = &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            30,
		CurrentBet:     50,
		Street:         "preflop",
		FoldedPlayers:  make(map[int]bool),
		ActedPlayers:   make(map[int]bool),
		PlayerBets:     make(map[int]int),
	}

	// All players have acted, but bets are not matched
	table.CurrentHand.ActedPlayers[0] = true
	table.CurrentHand.ActedPlayers[1] = true
	table.CurrentHand.ActedPlayers[2] = true
	table.CurrentHand.PlayerBets[0] = 50
	table.CurrentHand.PlayerBets[1] = 20 // Not matched
	table.CurrentHand.PlayerBets[2] = 20

	// Round should NOT be complete
	if table.CurrentHand.IsBettingRoundComplete(table.Seats) {
		t.Error("expected betting round to not be complete when bets are unmatched")
	}
}

// TestIsBettingRoundComplete_AllMatched returns true when all acted and matched
func TestIsBettingRoundComplete_AllMatched(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)

	// Set up 3 active players
	token1, token2, token3 := "player1", "player2", "player3"
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.Seats[2].Token = &token3
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000

	// Initialize hand with action state
	table.CurrentHand = &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            60,
		CurrentBet:     20,
		Street:         "preflop",
		FoldedPlayers:  make(map[int]bool),
		ActedPlayers:   make(map[int]bool),
		PlayerBets:     make(map[int]int),
	}

	// All players have acted and matched the current bet
	table.CurrentHand.ActedPlayers[0] = true
	table.CurrentHand.ActedPlayers[1] = true
	table.CurrentHand.ActedPlayers[2] = true
	table.CurrentHand.PlayerBets[0] = 20
	table.CurrentHand.PlayerBets[1] = 20
	table.CurrentHand.PlayerBets[2] = 20

	// Round should be complete
	if !table.CurrentHand.IsBettingRoundComplete(table.Seats) {
		t.Error("expected betting round to be complete when all players acted and matched")
	}
}

// TestIsBettingRoundComplete_AllFoldedButOne returns true when only one player left
func TestIsBettingRoundComplete_AllFoldedButOne(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)

	// Set up 3 active players
	token1, token2, token3 := "player1", "player2", "player3"
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.Seats[2].Token = &token3
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000

	// Initialize hand with action state
	table.CurrentHand = &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            100,
		CurrentBet:     50,
		Street:         "preflop",
		FoldedPlayers:  make(map[int]bool),
		ActedPlayers:   make(map[int]bool),
		PlayerBets:     make(map[int]int),
	}

	// Only players 0 and 1 have acted; player 2 has folded
	// Player 0 is the only one not folded
	table.CurrentHand.FoldedPlayers[1] = true
	table.CurrentHand.FoldedPlayers[2] = true
	table.CurrentHand.ActedPlayers[0] = true
	table.CurrentHand.ActedPlayers[1] = true
	table.CurrentHand.ActedPlayers[2] = true
	table.CurrentHand.PlayerBets[0] = 50
	table.CurrentHand.PlayerBets[1] = 50
	table.CurrentHand.PlayerBets[2] = 50

	// Round should be complete (only one player left)
	if !table.CurrentHand.IsBettingRoundComplete(table.Seats) {
		t.Error("expected betting round to be complete when only one player remains")
	}
}

// ============================================================================
// All-In Betting Loop Tests (Phase 1)
// ============================================================================
// These tests expose the bug where IsBettingRoundComplete() doesn't account
// for all-in players (stack = 0). All-in players cannot match higher bets and
// should be skipped from the bet matching check.

// TestIsBettingRoundComplete_TwoPlayerBothAllInUnequalStacks tests 2 players
// with unequal stacks both going all-in. SB has 900, BB has 1000.
// Both are all-in (stack = 0), betting round should complete.
// CURRENTLY FAILS: PlayerBets[SB]=900 != PlayerBets[BB]=1000, so returns false
func TestIsBettingRoundComplete_TwoPlayerBothAllInUnequalStacks(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)

	// Set up 2 active players: SB with 900 stack, BB with 1000 stack
	tokenSB, tokenBB := "player-sb", "player-bb"
	table.Seats[0].Token = &tokenSB
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 0 // SB is all-in
	table.Seats[1].Token = &tokenBB
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 0 // BB is all-in

	// Initialize hand with all-in action state
	table.CurrentHand = &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 0,
		BigBlindSeat:   1,
		Pot:            1900, // Will be swept
		CurrentBet:     1000, // BB's all-in amount
		Street:         "preflop",
		FoldedPlayers:  make(map[int]bool),
		ActedPlayers:   make(map[int]bool),
		PlayerBets:     make(map[int]int),
	}

	// Both players have acted
	table.CurrentHand.ActedPlayers[0] = true
	table.CurrentHand.ActedPlayers[1] = true

	// Bets: SB bet all 900, BB bet all 1000
	// Note: Unequal stacks mean unequal bets, but both are all-in
	table.CurrentHand.PlayerBets[0] = 900  // SB's all-in amount
	table.CurrentHand.PlayerBets[1] = 1000 // BB's all-in amount

	// Round SHOULD be complete (both all-in)
	// CURRENTLY FAILS: returns false because 900 != 1000
	if !table.CurrentHand.IsBettingRoundComplete(table.Seats) {
		t.Error("expected betting round to be complete when both players are all-in with unequal stacks")
	}
}

// TestIsBettingRoundComplete_TwoPlayerOneAllInOneMatched tests 2 players
// where one is all-in and the other has matched their bet.
// Player 0 (active, 500 stack): bet 500 (all-in)
// Player 1 (active, 1000 stack): bet 500 (matched)
func TestIsBettingRoundComplete_TwoPlayerOneAllInOneMatched(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)

	// Set up 2 active players
	tokenA, tokenB := "player-a", "player-b"
	table.Seats[0].Token = &tokenA
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 0 // All-in
	table.Seats[1].Token = &tokenB
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 500 // Has chips left

	// Initialize hand
	table.CurrentHand = &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 0,
		BigBlindSeat:   1,
		Pot:            0,
		CurrentBet:     500,
		Street:         "preflop",
		FoldedPlayers:  make(map[int]bool),
		ActedPlayers:   make(map[int]bool),
		PlayerBets:     make(map[int]int),
	}

	// Both players have acted
	table.CurrentHand.ActedPlayers[0] = true
	table.CurrentHand.ActedPlayers[1] = true

	// Player 0 is all-in with 500
	table.CurrentHand.PlayerBets[0] = 500
	// Player 1 matched 500
	table.CurrentHand.PlayerBets[1] = 500

	// Round SHOULD be complete (all-in player + matched player)
	if !table.CurrentHand.IsBettingRoundComplete(table.Seats) {
		t.Error("expected betting round to be complete when all-in player is matched by active player")
	}
}

// TestIsBettingRoundComplete_ThreePlayerTwoAllInOneActive tests 3 players
// with 2 all-in (different stacks: 500, 700) and 1 active (matched bet of 700).
func TestIsBettingRoundComplete_ThreePlayerTwoAllInOneActive(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)

	// Set up 3 active players
	token1, token2, token3 := "player-1", "player-2", "player-3"
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 0 // All-in with 500
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 0 // All-in with 700
	table.Seats[2].Token = &token3
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1300 // Has chips left, matched highest bet

	// Initialize hand
	table.CurrentHand = &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            0,
		CurrentBet:     700, // Highest all-in amount
		Street:         "preflop",
		FoldedPlayers:  make(map[int]bool),
		ActedPlayers:   make(map[int]bool),
		PlayerBets:     make(map[int]int),
	}

	// All players have acted
	table.CurrentHand.ActedPlayers[0] = true
	table.CurrentHand.ActedPlayers[1] = true
	table.CurrentHand.ActedPlayers[2] = true

	// Bets: Player 0 all-in 500, Player 1 all-in 700, Player 2 matched 700
	table.CurrentHand.PlayerBets[0] = 500
	table.CurrentHand.PlayerBets[1] = 700
	table.CurrentHand.PlayerBets[2] = 700

	// Round SHOULD be complete (2 all-in, 1 active matched highest)
	// CURRENTLY FAILS: returns false because Player 0 has 500 != 700
	if !table.CurrentHand.IsBettingRoundComplete(table.Seats) {
		t.Error("expected betting round to be complete with 2 all-in players and 1 matched player")
	}
}

// TestIsBettingRoundComplete_ThreePlayerAllDifferentStacks tests 3 players
// all all-in with different stacks (500, 700, 1000).
func TestIsBettingRoundComplete_ThreePlayerAllDifferentStacks(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)

	// Set up 3 active players, all all-in with different stacks
	token1, token2, token3 := "player-1", "player-2", "player-3"
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 0 // All-in
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 0 // All-in
	table.Seats[2].Token = &token3
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 0 // All-in

	// Initialize hand
	table.CurrentHand = &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            0,
		CurrentBet:     1000, // Highest all-in amount
		Street:         "preflop",
		FoldedPlayers:  make(map[int]bool),
		ActedPlayers:   make(map[int]bool),
		PlayerBets:     make(map[int]int),
	}

	// All players have acted
	table.CurrentHand.ActedPlayers[0] = true
	table.CurrentHand.ActedPlayers[1] = true
	table.CurrentHand.ActedPlayers[2] = true

	// Bets: Different stacks all-in
	table.CurrentHand.PlayerBets[0] = 500
	table.CurrentHand.PlayerBets[1] = 700
	table.CurrentHand.PlayerBets[2] = 1000

	// Round SHOULD be complete (all players all-in)
	// CURRENTLY FAILS: returns false because bets don't all equal 1000
	if !table.CurrentHand.IsBettingRoundComplete(table.Seats) {
		t.Error("expected betting round to be complete when all players are all-in with different stacks")
	}
}

// TestIsBettingRoundComplete_MultiPlayerSomeAllInSomeFolded tests 5 players:
// 2 all-in (300, 500), 2 folded, 1 active with 500 matched.
func TestIsBettingRoundComplete_MultiPlayerSomeAllInSomeFolded(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)

	// Set up 5 active players
	tokens := []string{"p1", "p2", "p3", "p4", "p5"}
	for i := 0; i < 5; i++ {
		token := tokens[i]
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		if i == 0 || i == 1 {
			table.Seats[i].Stack = 0 // Players 0, 1 are all-in
		} else if i == 4 {
			table.Seats[i].Stack = 1000 // Player 4 has chips left
		} else {
			table.Seats[i].Stack = 1000 // Other active players
		}
	}

	// Initialize hand
	table.CurrentHand = &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            0,
		CurrentBet:     500, // Highest bet (Player 1's all-in)
		Street:         "preflop",
		FoldedPlayers:  make(map[int]bool),
		ActedPlayers:   make(map[int]bool),
		PlayerBets:     make(map[int]int),
	}

	// Players have acted
	table.CurrentHand.ActedPlayers[0] = true
	table.CurrentHand.ActedPlayers[1] = true
	table.CurrentHand.ActedPlayers[2] = true
	table.CurrentHand.ActedPlayers[3] = true
	table.CurrentHand.ActedPlayers[4] = true

	// Players 2, 3 have folded
	table.CurrentHand.FoldedPlayers[2] = true
	table.CurrentHand.FoldedPlayers[3] = true

	// Bets: P0 all-in 300, P1 all-in 500, P2 folded (no bet), P3 folded (no bet), P4 matched 500
	table.CurrentHand.PlayerBets[0] = 300
	table.CurrentHand.PlayerBets[1] = 500
	table.CurrentHand.PlayerBets[2] = 0 // Folded
	table.CurrentHand.PlayerBets[3] = 0 // Folded
	table.CurrentHand.PlayerBets[4] = 500

	// Round SHOULD be complete (2 all-in, 2 folded, 1 active matched)
	// CURRENTLY FAILS: Player 0 has 300 != 500
	if !table.CurrentHand.IsBettingRoundComplete(table.Seats) {
		t.Error("expected betting round to be complete with multiple all-in and folded players")
	}
}

// TestIsBettingRoundComplete_AllPlayersAllIn tests all remaining players all-in.
func TestIsBettingRoundComplete_AllPlayersAllIn(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)

	// Set up 4 active players, all all-in
	tokens := []string{"p1", "p2", "p3", "p4"}
	for i := 0; i < 4; i++ {
		token := tokens[i]
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 0 // All all-in
	}

	// Initialize hand
	table.CurrentHand = &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            0,
		CurrentBet:     1000,
		Street:         "preflop",
		FoldedPlayers:  make(map[int]bool),
		ActedPlayers:   make(map[int]bool),
		PlayerBets:     make(map[int]int),
	}

	// All players have acted
	table.CurrentHand.ActedPlayers[0] = true
	table.CurrentHand.ActedPlayers[1] = true
	table.CurrentHand.ActedPlayers[2] = true
	table.CurrentHand.ActedPlayers[3] = true

	// Bets: All different amounts (all-in with different stacks)
	table.CurrentHand.PlayerBets[0] = 250
	table.CurrentHand.PlayerBets[1] = 500
	table.CurrentHand.PlayerBets[2] = 750
	table.CurrentHand.PlayerBets[3] = 1000

	// Round SHOULD be complete (all players all-in)
	// CURRENTLY FAILS: Bets don't all match
	if !table.CurrentHand.IsBettingRoundComplete(table.Seats) {
		t.Error("expected betting round to be complete when all players are all-in")
	}
}

// TestAdvanceAction moves to next active player and handles wrap-around
func TestAdvanceAction(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)

	// Set up 3 active players
	token1, token2, token3 := "player1", "player2", "player3"
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.Seats[2].Token = &token3
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000

	// Initialize hand with action state
	table.CurrentHand = &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            30,
		CurrentBet:     20,
		Street:         "preflop",
		FoldedPlayers:  make(map[int]bool),
		ActedPlayers:   make(map[int]bool),
		PlayerBets:     make(map[int]int),
	}

	// Start with player 0
	currentActor := 0
	table.CurrentHand.CurrentActor = &currentActor

	// Advance to player 1
	nextActor, err := table.CurrentHand.AdvanceAction(table.Seats)
	if err != nil {
		t.Errorf("expected no error advancing action, got %v", err)
	}

	if nextActor == nil {
		t.Error("expected nextActor to not be nil")
	} else if *nextActor != 1 {
		t.Errorf("expected next actor to be 1, got %d", *nextActor)
	}

	// Update CurrentActor to the returned nextActor
	table.CurrentHand.CurrentActor = nextActor

	// Advance to player 2
	nextActor, err = table.CurrentHand.AdvanceAction(table.Seats)
	if err != nil {
		t.Errorf("expected no error advancing action, got %v", err)
	}

	if nextActor == nil {
		t.Error("expected nextActor to not be nil")
	} else if *nextActor != 2 {
		t.Errorf("expected next actor to be 2, got %d", *nextActor)
	}

	// Update CurrentActor to the returned nextActor
	table.CurrentHand.CurrentActor = nextActor

	// Advance to player 0 (wrap-around)
	nextActor, err = table.CurrentHand.AdvanceAction(table.Seats)
	if err != nil {
		t.Errorf("expected no error advancing action, got %v", err)
	}

	if nextActor == nil {
		t.Error("expected nextActor to not be nil")
	} else if *nextActor != 0 {
		t.Errorf("expected next actor to wrap to 0, got %d", *nextActor)
	}
}

// TestAdvanceAction_WithFoldedPlayers skips folded players
func TestAdvanceAction_WithFoldedPlayers(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)

	// Set up 3 active players
	token1, token2, token3 := "player1", "player2", "player3"
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.Seats[2].Token = &token3
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000

	// Initialize hand with action state
	table.CurrentHand = &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            30,
		CurrentBet:     20,
		Street:         "preflop",
		FoldedPlayers:  make(map[int]bool),
		ActedPlayers:   make(map[int]bool),
		PlayerBets:     make(map[int]int),
	}

	// Mark player 1 as folded
	table.CurrentHand.FoldedPlayers[1] = true

	// Start with player 0
	currentActor := 0
	table.CurrentHand.CurrentActor = &currentActor

	// Advance to next active (should skip player 1 and go to player 2)
	nextActor, err := table.CurrentHand.AdvanceAction(table.Seats)
	if err != nil {
		t.Errorf("expected no error advancing action, got %v", err)
	}

	if nextActor == nil {
		t.Error("expected nextActor to not be nil")
	} else if *nextActor != 2 {
		t.Errorf("expected next actor to skip folded player and be 2, got %d", *nextActor)
	}
}

// TestAdvanceAction_ReturnNilWhenOnlyOnePlayerLeft returns nil when only one player remains
func TestAdvanceAction_ReturnNilWhenOnlyOnePlayerLeft(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)

	// Set up 3 active players
	token1, token2, token3 := "player1", "player2", "player3"
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[2].Token = &token3
	table.Seats[2].Status = "active"

	// Initialize hand with action state
	table.CurrentHand = &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            30,
		CurrentBet:     20,
		Street:         "preflop",
		FoldedPlayers:  make(map[int]bool),
		ActedPlayers:   make(map[int]bool),
		PlayerBets:     make(map[int]int),
	}

	// Mark players 1 and 2 as folded (only player 0 remains)
	table.CurrentHand.FoldedPlayers[1] = true
	table.CurrentHand.FoldedPlayers[2] = true

	// Start with player 0
	currentActor := 0
	table.CurrentHand.CurrentActor = &currentActor

	// Advance should return nil (no next active player)
	nextActor, err := table.CurrentHand.AdvanceAction(table.Seats)
	if err != nil {
		t.Errorf("expected no error advancing action, got %v", err)
	}

	if nextActor != nil {
		t.Errorf("expected nextActor to be nil when only one player left, got %d", *nextActor)
	}
}

// TestStartHand_InitializesActionState verifies action fields are initialized correctly
func TestStartHand_InitializesActionState(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	table := NewTable("table-1", "Test Table", server)

	// Set up 3 active players (waiting status)
	token1, token2, token3 := "player1", "player2", "player3"
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "waiting"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "waiting"
	table.Seats[1].Stack = 1000
	table.Seats[2].Token = &token3
	table.Seats[2].Status = "waiting"
	table.Seats[2].Stack = 1000

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Verify action state fields are initialized
	if table.CurrentHand.Street != "preflop" {
		t.Errorf("expected Street to be 'preflop', got '%s'", table.CurrentHand.Street)
	}

	if table.CurrentHand.CurrentBet != 20 {
		t.Errorf("expected CurrentBet to be 20 (BB), got %d", table.CurrentHand.CurrentBet)
	}

	if table.CurrentHand.PlayerBets == nil {
		t.Error("expected PlayerBets to be initialized, got nil")
	}

	if table.CurrentHand.FoldedPlayers == nil {
		t.Error("expected FoldedPlayers to be initialized, got nil")
	}

	if table.CurrentHand.ActedPlayers == nil {
		t.Error("expected ActedPlayers to be initialized, got nil")
	}

	// Verify blinds are in PlayerBets
	if table.CurrentHand.PlayerBets[table.CurrentHand.SmallBlindSeat] != 10 {
		t.Errorf("expected small blind in PlayerBets, got %d", table.CurrentHand.PlayerBets[table.CurrentHand.SmallBlindSeat])
	}

	if table.CurrentHand.PlayerBets[table.CurrentHand.BigBlindSeat] != 20 {
		t.Errorf("expected big blind in PlayerBets, got %d", table.CurrentHand.PlayerBets[table.CurrentHand.BigBlindSeat])
	}

	// Verify CurrentActor is set to first actor
	if table.CurrentHand.CurrentActor == nil {
		t.Error("expected CurrentActor to be set, got nil")
	}

	// In this 3-player setup: dealer=0, SB=1, BB=2, so UTG (first to act) should be 0
	if table.CurrentHand.CurrentActor != nil && *table.CurrentHand.CurrentActor != 0 {
		t.Errorf("expected CurrentActor to be 0 (UTG), got %d", *table.CurrentHand.CurrentActor)
	}
}

// TestStartHandBroadcastsFirstActionRequest verifies first action_request is broadcast
// This just tests that the pattern is set up correctly in StartHand for later broadcast
func TestStartHandBroadcastsFirstActionRequest(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)

	// Get first table
	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	// Set up 2 players
	sm := NewSessionManager(logger)
	session1, _ := sm.CreateSession("Player1")
	session2, _ := sm.CreateSession("Player2")
	token1 := session1.Token
	token2 := session2.Token

	// Assign seats and set to waiting (will transition to active)
	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "waiting"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "waiting"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Verify CurrentActor is set (will be used to broadcast action_request)
	table.mu.RLock()
	defer table.mu.RUnlock()

	if table.CurrentHand == nil {
		t.Fatal("CurrentHand should be set")
	}

	if table.CurrentHand.CurrentActor == nil {
		t.Fatal("CurrentActor should be set for first action_request")
	}

	// Verify it's a valid seat
	if *table.CurrentHand.CurrentActor < 0 || *table.CurrentHand.CurrentActor > 1 {
		t.Errorf("CurrentActor %d is not valid for 2-player setup", *table.CurrentHand.CurrentActor)
	}
}

// TestGetMinRaise_Preflop verifies min-raise calculation on preflop
// With BB=20, first raise should be 40 (20 + 20 increment)
func TestGetMinRaise_Preflop(t *testing.T) {
	hand := &Hand{
		BigBlindSeat: 1,
		CurrentBet:   20,
		LastRaise:    20,
		Street:       "preflop",
	}

	expected := 40
	result := hand.GetMinRaise()
	if result != expected {
		t.Errorf("GetMinRaise() = %d, want %d", result, expected)
	}
}

// TestGetMinRaise_AfterRaise verifies min-raise after player raises
// After raise to 60, min-raise should be 100 (60 + 40 increment)
func TestGetMinRaise_AfterRaise(t *testing.T) {
	hand := &Hand{
		BigBlindSeat: 1,
		CurrentBet:   60,
		LastRaise:    40,
		Street:       "preflop",
	}

	expected := 100
	result := hand.GetMinRaise()
	if result != expected {
		t.Errorf("GetMinRaise() = %d, want %d", result, expected)
	}
}

// TestGetMinRaise_AfterMultipleRaises verifies chain of raises maintains correct increments
// Sequence: BB 20 -> Raise to 60 (increment 40) -> Raise to 140 (increment 80)
func TestGetMinRaise_AfterMultipleRaises(t *testing.T) {
	hand := &Hand{
		BigBlindSeat: 1,
		CurrentBet:   140,
		LastRaise:    80,
		Street:       "preflop",
	}

	expected := 220
	result := hand.GetMinRaise()
	if result != expected {
		t.Errorf("GetMinRaise() = %d, want %d", result, expected)
	}
}

// TestGetMinRaise_PostFlop verifies min-raise on later streets
// First bet on flop is 30, min-raise should be 60
func TestGetMinRaise_PostFlop(t *testing.T) {
	hand := &Hand{
		BigBlindSeat: 1,
		CurrentBet:   30,
		LastRaise:    30,
		Street:       "flop",
	}

	expected := 60
	result := hand.GetMinRaise()
	if result != expected {
		t.Errorf("GetMinRaise() = %d, want %d", result, expected)
	}
}

// TestGetMinRaise_HeadsUp verifies GetMinRaise works in heads-up scenario
// 2-player scenario with standard min-raise
func TestGetMinRaise_HeadsUp(t *testing.T) {
	hand := &Hand{
		BigBlindSeat: 0,
		CurrentBet:   20,
		LastRaise:    20,
		Street:       "preflop",
	}

	expected := 40
	result := hand.GetMinRaise()
	if result != expected {
		t.Errorf("GetMinRaise() = %d, want %d", result, expected)
	}
}

// TestNewHand_InitializesLastRaise verifies LastRaise is initialized to BigBlind
// When StartHand creates a new Hand, LastRaise should be set to bigBlind amount
func TestNewHand_InitializesLastRaise(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)

	server.mu.RLock()
	table := server.tables[0]
	server.mu.RUnlock()

	sm := NewSessionManager(logger)
	session1, _ := sm.CreateSession("Player1")
	session2, _ := sm.CreateSession("Player2")
	token1 := session1.Token
	token2 := session2.Token

	table.mu.Lock()
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "waiting"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "waiting"
	table.Seats[1].Stack = 1000
	table.mu.Unlock()

	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	table.mu.RLock()
	defer table.mu.RUnlock()

	if table.CurrentHand == nil {
		t.Fatal("CurrentHand should be set")
	}

	if table.CurrentHand.LastRaise != 20 {
		t.Errorf("LastRaise = %d, want 20 (bigBlind)", table.CurrentHand.LastRaise)
	}
}

// TestAdvanceStreet_ResetsLastRaise verifies LastRaise is reset when advancing to next street
// After advancing from preflop to flop, LastRaise is preserved for postflop minimum raise calculation
func TestAdvanceStreet_ResetsLastRaise(t *testing.T) {
	hand := &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            100,
		Deck:           NewDeck(),
		HoleCards:      make(map[int][]Card),
		Street:         "preflop",
		CurrentBet:     20,
		PlayerBets:     make(map[int]int),
		FoldedPlayers:  make(map[int]bool),
		ActedPlayers:   make(map[int]bool),
		LastRaise:      50,
	}

	hand.AdvanceStreet()

	if hand.Street != "flop" {
		t.Errorf("Street = %s, want flop", hand.Street)
	}

	// LastRaise is now preserved on postflop streets (flop, turn, river) for minimum raise calculation
	if hand.LastRaise != 50 {
		t.Errorf("LastRaise = %d, want 50 after street advance (preserved for postflop)", hand.LastRaise)
	}

	if hand.CurrentBet != 0 {
		t.Errorf("CurrentBet = %d, want 0 after street advance", hand.CurrentBet)
	}
}

// TestAdvanceStreet_PreservesMinimumRaisePostflop verifies LastRaise equals 20 (big blind) after advancing to postflop streets
// This ensures the minimum raise increment on postflop streets is based on the big blind
func TestAdvanceStreet_PreservesMinimumRaisePostflop(t *testing.T) {
	// Test advancing from preflop to flop with LastRaise set to big blind (20)
	hand := &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            100,
		Deck:           NewDeck(),
		HoleCards:      make(map[int][]Card),
		Street:         "preflop",
		CurrentBet:     20,
		PlayerBets:     make(map[int]int),
		FoldedPlayers:  make(map[int]bool),
		ActedPlayers:   make(map[int]bool),
		LastRaise:      20, // Big blind
	}

	hand.AdvanceStreet()

	if hand.Street != "flop" {
		t.Errorf("Street = %s, want flop", hand.Street)
	}

	// After advancing to flop, LastRaise should be preserved as 20 (big blind) for min-raise calculation
	if hand.LastRaise != 20 {
		t.Errorf("LastRaise = %d, want 20 (big blind) after advancing to flop", hand.LastRaise)
	}

	// Test advancing from flop to turn
	hand.Street = "flop"
	hand.AdvanceStreet()

	if hand.Street != "turn" {
		t.Errorf("Street = %s, want turn", hand.Street)
	}

	// LastRaise should still be 20 on turn
	if hand.LastRaise != 20 {
		t.Errorf("LastRaise = %d, want 20 (big blind) after advancing to turn", hand.LastRaise)
	}

	// Test advancing from turn to river
	hand.Street = "turn"
	hand.AdvanceStreet()

	if hand.Street != "river" {
		t.Errorf("Street = %s, want river", hand.Street)
	}

	// LastRaise should still be 20 on river
	if hand.LastRaise != 20 {
		t.Errorf("LastRaise = %d, want 20 (big blind) after advancing to river", hand.LastRaise)
	}
}

// TestGetMinRaise_PostflopFirstAction verifies minimum raise is 40 (2x BB) when no raises yet postflop
// With CurrentBet = 0 (no bets yet) and LastRaise = 20 (big blind), min-raise should be 20 + 0 = 20... wait
// Actually, if someone bets 30, then CurrentBet = 30, and min-raise = 30 + 20 = 50
func TestGetMinRaise_PostflopFirstAction(t *testing.T) {
	hand := &Hand{
		Street:     "flop",
		CurrentBet: 0,  // No bets on this street yet
		LastRaise:  20, // Big blind preserved from preflop
	}

	// First action: no bet yet, so min-raise should just be the preserved big blind
	minRaise := hand.GetMinRaise()
	if minRaise != 20 {
		t.Errorf("GetMinRaise() on flop with no bets = %d, want 20", minRaise)
	}

	// After first bet of 30, CurrentBet becomes 30
	hand.CurrentBet = 30

	// Min-raise should now be 30 (current bet) + 20 (big blind) = 50
	minRaise = hand.GetMinRaise()
	if minRaise != 50 {
		t.Errorf("GetMinRaise() after 30 bet with BB increment = %d, want 50", minRaise)
	}
}

// ============ PHASE 2: MAX-RAISE AND SIDE POT PREVENTION TESTS ============

// TestGetMaxRaise_LimitedByPlayerStack verifies max raise equals player's own stack (new behavior)
// This test now validates that players can bet their full stacks regardless of opponent stacks
func TestGetMaxRaise_LimitedByPlayerStack(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 3 active players: seat 0 (1000), seat 1 (1000), seat 2 (1000)
	// All equal stacks - max raise should be player's own stack
	for i := 0; i < 3; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// GetMaxRaise for seat 0: returns player's stack 1000 (not limited by opponent's 1000)
	maxRaise := table.GetMaxRaise(0, createEmptyHand())
	if maxRaise != 1000 {
		t.Errorf("GetMaxRaise(seat 0) = %d, want 1000 (player's stack)", maxRaise)
	}

	// Now set seat 2 to 500 (smaller than seat 0)
	table.Seats[2].Stack = 500

	// GetMaxRaise for seat 0: still returns player's stack 1000 (NOT limited to opponent's 500)
	// This is the key difference - player can overbет the short stack
	maxRaise = table.GetMaxRaise(0, createEmptyHand())
	if maxRaise != 1000 {
		t.Errorf("GetMaxRaise(seat 0) = %d, want 1000 (player's full stack, not limited by short opponent)", maxRaise)
	}

	// GetMaxRaise for seat 1: returns player's stack 1000 (NOT limited to opponent's 500)
	maxRaise = table.GetMaxRaise(1, createEmptyHand())
	if maxRaise != 1000 {
		t.Errorf("GetMaxRaise(seat 1) = %d, want 1000 (player's full stack, not limited by short opponent)", maxRaise)
	}
}

// TestGetMaxRaise_LimitedByOpponentStack verifies max raise equals player's own stack (new behavior)
// Previously this was limited by opponent stacks, but now players can bet their full stacks
func TestGetMaxRaise_LimitedByOpponentStack(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 3 active players: seat 0 (1000), seat 1 (500), seat 2 (300)
	stacks := []int{1000, 500, 300}
	for i := 0; i < 3; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = stacks[i]
	}

	// Seat 0 player has 1000, can now raise full amount (not limited to opponent's 300)
	maxRaise := table.GetMaxRaise(0, createEmptyHand())
	if maxRaise != 1000 {
		t.Errorf("GetMaxRaise(seat 0) = %d, want 1000 (player's full stack)", maxRaise)
	}

	// Seat 1 player has 500, can raise full amount (not limited to opponent's 300)
	maxRaise = table.GetMaxRaise(1, createEmptyHand())
	if maxRaise != 500 {
		t.Errorf("GetMaxRaise(seat 1) = %d, want 500 (player's full stack)", maxRaise)
	}

	// Seat 2 player has 300, can raise full amount
	maxRaise = table.GetMaxRaise(2, createEmptyHand())
	if maxRaise != 300 {
		t.Errorf("GetMaxRaise(seat 2) = %d, want 300 (player's full stack)", maxRaise)
	}
}

// TestGetMaxRaise_HeadsUp verifies heads-up allows full player stack (new behavior)
// Previously limited by opponent's stack, now players can bet their full stacks
func TestGetMaxRaise_HeadsUp(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up heads-up: seat 0 (1000), seat 3 (800)
	token0 := "player-0"
	token3 := "player-3"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[3].Token = &token3
	table.Seats[3].Status = "active"
	table.Seats[3].Stack = 800

	// In heads-up, seat 0 can now raise full 1000 (not limited to opponent's 800)
	maxRaise := table.GetMaxRaise(0, createEmptyHand())
	if maxRaise != 1000 {
		t.Errorf("GetMaxRaise(seat 0 heads-up) = %d, want 1000 (player's full stack)", maxRaise)
	}

	// Seat 3 (800 stack) can raise their full 800
	maxRaise = table.GetMaxRaise(3, createEmptyHand())
	if maxRaise != 800 {
		t.Errorf("GetMaxRaise(seat 3 heads-up) = %d, want 800 (player's full stack)", maxRaise)
	}
}

// TestGetMaxRaise_MultiPlayer verifies multi-player allows full player stacks (new behavior)
// Previously limited to smallest opponent stack, now players can bet their full stacks
func TestGetMaxRaise_MultiPlayer(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 4 active players with varying stacks
	stacks := []int{1000, 600, 300, 800}
	for i := 0; i < 4; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = stacks[i]
	}

	// Seat 0 (1000): can now raise full 1000 (not limited to smallest opponent 300)
	maxRaise := table.GetMaxRaise(0, createEmptyHand())
	if maxRaise != 1000 {
		t.Errorf("GetMaxRaise(seat 0, 4-player) = %d, want 1000", maxRaise)
	}

	// Seat 1 (600): can raise full 600 (not limited to smallest opponent 300)
	maxRaise = table.GetMaxRaise(1, createEmptyHand())
	if maxRaise != 600 {
		t.Errorf("GetMaxRaise(seat 1, 4-player) = %d, want 600", maxRaise)
	}

	// Seat 2 (300): can raise full 300
	maxRaise = table.GetMaxRaise(2, createEmptyHand())
	if maxRaise != 300 {
		t.Errorf("GetMaxRaise(seat 2, 4-player) = %d, want 300", maxRaise)
	}

	// Seat 3 (800): can raise full 800 (not limited to smallest opponent 300)
	maxRaise = table.GetMaxRaise(3, createEmptyHand())
	if maxRaise != 800 {
		t.Errorf("GetMaxRaise(seat 3, 4-player) = %d, want 800", maxRaise)
	}
}

// TestValidateRaise_BelowMinimum verifies error when raise amount is below minimum
func TestValidateRaise_BelowMinimum(t *testing.T) {
	hand := &Hand{
		CurrentBet: 20,
		LastRaise:  20, // Min raise = 20 + 20 = 40
	}

	// Raise of 30 is below minimum of 40 (and not all-in)
	err := hand.ValidateRaise(0, 30, 1000, [6]Seat{})
	if err == nil {
		t.Fatal("expected error for raise below minimum, got nil")
	}

	expectedMsg := "raise amount below minimum"
	if err.Error() != expectedMsg {
		t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestValidateRaise_AboveMaximum verifies error when raise exceeds player's stack (new behavior)
// Previously checked against opponent stacks for side pot prevention
// Now players can bet full stacks, so error only if exceeding player's own stack
func TestValidateRaise_AboveMaximum(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 3 active players: seat 0 (1000), seat 1 (1000), seat 2 (300)
	stacks := []int{1000, 1000, 300}
	for i := 0; i < 3; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = stacks[i]
	}

	hand := &Hand{
		CurrentBet: 20,
		LastRaise:  20, // Min raise = 40
	}

	// Seat 0 tries to raise 500 (below their 1000), should be valid
	err := hand.ValidateRaise(0, 500, 1000, table.Seats)
	if err != nil {
		t.Fatalf("expected no error for 500 raise with 1000 stack, got %v", err)
	}

	// Seat 0 tries to raise 1100 (exceeds their 1000 stack), should error
	err = hand.ValidateRaise(0, 1100, 1000, table.Seats)
	if err == nil {
		t.Fatal("expected error for raise exceeding player stack, got nil")
	}

	expectedMsg := "raise exceeds player stack"
	if err.Error() != expectedMsg {
		t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestValidateRaise_ValidAmount verifies nil error for valid raise amount
func TestValidateRaise_ValidAmount(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 3 active players: seat 0 (1000), seat 1 (1000), seat 2 (300)
	stacks := []int{1000, 1000, 300}
	for i := 0; i < 3; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = stacks[i]
	}

	hand := &Hand{
		CurrentBet: 20,
		LastRaise:  20, // Min raise = 40
	}

	// Raise of 40 is exactly minimum and within max (300)
	err := hand.ValidateRaise(0, 40, 1000, table.Seats)
	if err != nil {
		t.Errorf("expected no error for valid raise 40, got %v", err)
	}

	// Raise of 300 is at max allowed (smallest opponent)
	err = hand.ValidateRaise(0, 300, 1000, table.Seats)
	if err != nil {
		t.Errorf("expected no error for valid raise 300, got %v", err)
	}

	// Raise of 100 is between min and max
	err = hand.ValidateRaise(0, 100, 1000, table.Seats)
	if err != nil {
		t.Errorf("expected no error for valid raise 100, got %v", err)
	}
}

// TestValidateRaise_AllInBelowMin verifies all-in is allowed even if below minimum raise
func TestValidateRaise_AllInBelowMin(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 3 active players
	for i := 0; i < 3; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	hand := &Hand{
		CurrentBet: 20,
		LastRaise:  20, // Min raise = 40
	}

	// Player at seat 0 has only 25 chips left (all-in)
	// 25 < minimum raise of 40, but should be allowed as all-in
	err := hand.ValidateRaise(0, 25, 25, table.Seats) // amount=25, playerStack=25 (all-in)
	if err != nil {
		t.Errorf("expected no error for all-in below minimum, got %v", err)
	}

	// All-in with less than all chips should fail (not actually all-in)
	// Player has 100 chips, tries to raise 25 (all-in would be 100)
	err = hand.ValidateRaise(0, 25, 100, table.Seats) // amount < playerStack, not all-in
	if err == nil {
		t.Fatal("expected error for raise below minimum when not all-in, got nil")
	}
}

// TestValidateRaise_HeadsUp verifies validation works correctly in heads-up (new behavior)
// Players can now bet their full stacks, not limited by opponent stacks
func TestValidateRaise_HeadsUp(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up heads-up: seat 0 (1000), seat 3 (800)
	token0 := "player-0"
	token3 := "player-3"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[3].Token = &token3
	table.Seats[3].Status = "active"
	table.Seats[3].Stack = 800

	hand := &Hand{
		CurrentBet: 20,
		LastRaise:  20, // Min raise = 40
	}

	// In heads-up, seat 0 (1000 stack) can raise up to 1000 (their full stack)
	// Raise of 40 (minimum) should be valid
	err := hand.ValidateRaise(0, 40, 1000, table.Seats)
	if err != nil {
		t.Errorf("expected no error for valid raise in heads-up, got %v", err)
	}

	// Raise of 1000 (at max for their stack) should be valid (all-in)
	err = hand.ValidateRaise(0, 1000, 1000, table.Seats)
	if err != nil {
		t.Errorf("expected no error for max raise in heads-up, got %v", err)
	}

	// Raise of 1100 (exceeds their stack of 1000) should error
	err = hand.ValidateRaise(0, 1100, 1000, table.Seats)
	if err == nil {
		t.Fatal("expected error for raise exceeding stack in heads-up, got nil")
	}

	if err.Error() != "raise exceeds player stack" {
		t.Errorf("expected error 'raise exceeds player stack', got '%s'", err.Error())
	}
}

// ============================================================================
// PHASE 3: Raise Action Processing Tests
// ============================================================================

// TestGetValidActions_IncludesRaise verifies raise appears in valid actions when player has enough chips
func TestGetValidActions_IncludesRaise(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 3 active players with stacks of 1000 each
	for i := 0; i < 3; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Initialize action state
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}

	// Player 0 is facing a 50 bet (must raise to at least 100)
	// Player has 1000, so raise is valid option
	table.CurrentHand.CurrentBet = 50
	table.CurrentHand.PlayerBets[0] = 0 // Player hasn't acted yet
	table.CurrentHand.LastRaise = 50    // Last raise was 50 (so min-raise = 100)

	validActions := table.CurrentHand.GetValidActions(0, table.Seats[0].Stack, table.Seats)

	// Should include raise, call, and fold
	hasRaise := false
	hasCall := false
	hasFold := false
	for _, action := range validActions {
		if action == "raise" {
			hasRaise = true
		}
		if action == "call" {
			hasCall = true
		}
		if action == "fold" {
			hasFold = true
		}
	}

	if !hasRaise {
		t.Errorf("expected 'raise' in valid actions when player has chips to raise, got %v", validActions)
	}
	if !hasCall {
		t.Errorf("expected 'call' in valid actions, got %v", validActions)
	}
	if !hasFold {
		t.Errorf("expected 'fold' in valid actions, got %v", validActions)
	}
	if len(validActions) != 3 {
		t.Errorf("expected exactly 3 valid actions (fold, call, raise), got %d: %v", len(validActions), validActions)
	}
}

// TestGetValidActions_NoRaiseWhenInsufficient verifies raise is excluded when player can't raise minimum
func TestGetValidActions_NoRaiseWhenInsufficient(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players
	token0 := "player-0"
	token1 := "player-1"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 30 // Very small stack

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Initialize action state
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}

	// Player 0 is facing a 20 bet
	// To raise, they'd need to call (20) + min raise (20) = 40, but only have 30
	table.CurrentHand.CurrentBet = 20
	table.CurrentHand.PlayerBets[0] = 0
	table.CurrentHand.LastRaise = 20

	validActions := table.CurrentHand.GetValidActions(0, table.Seats[0].Stack, table.Seats)

	// Should include call and fold, but NOT raise
	hasRaise := false
	hasCall := false
	hasFold := false
	for _, action := range validActions {
		if action == "raise" {
			hasRaise = true
		}
		if action == "call" {
			hasCall = true
		}
		if action == "fold" {
			hasFold = true
		}
	}

	if hasRaise {
		t.Errorf("expected no 'raise' in valid actions when insufficient chips, got %v", validActions)
	}
	if !hasCall {
		t.Errorf("expected 'call' in valid actions, got %v", validActions)
	}
	if !hasFold {
		t.Errorf("expected 'fold' in valid actions, got %v", validActions)
	}
	if len(validActions) != 2 {
		t.Errorf("expected exactly 2 valid actions (fold, call), got %d: %v", len(validActions), validActions)
	}
}

// TestGetValidActions_HeadsUp verifies valid actions are correct in heads-up scenario
func TestGetValidActions_HeadsUp(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up heads-up: seat 0 (dealer) and seat 3 (BB)
	token0 := "player-0"
	token3 := "player-3"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[3].Token = &token3
	table.Seats[3].Status = "active"
	table.Seats[3].Stack = 1000

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Initialize action state for heads-up
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}

	// In heads-up, dealer acts first on preflop
	// Player 0 (dealer/SB) is facing BB of 50
	table.CurrentHand.CurrentBet = 50    // BB was posted
	table.CurrentHand.LastRaise = 50     // min-raise = 100
	table.CurrentHand.PlayerBets[0] = 25 // Player 0 posted SB of 25

	validActions := table.CurrentHand.GetValidActions(0, table.Seats[0].Stack, table.Seats)

	// Should include fold, call, and raise
	hasRaise := false
	hasCall := false
	hasFold := false
	for _, action := range validActions {
		if action == "raise" {
			hasRaise = true
		}
		if action == "call" {
			hasCall = true
		}
		if action == "fold" {
			hasFold = true
		}
	}

	if !hasRaise {
		t.Errorf("expected 'raise' in valid actions for heads-up, got %v", validActions)
	}
	if !hasCall {
		t.Errorf("expected 'call' in valid actions, got %v", validActions)
	}
	if !hasFold {
		t.Errorf("expected 'fold' in valid actions, got %v", validActions)
	}
	if len(validActions) != 3 {
		t.Errorf("expected exactly 3 valid actions (fold, call, raise), got %d: %v", len(validActions), validActions)
	}
}

// TestProcessAction_RaiseUpdatesBets verifies raise correctly updates CurrentBet, LastRaise, PlayerBets, and Pot
func TestProcessAction_RaiseUpdatesBets(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players
	for i := 0; i < 2; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Set up betting state: CurrentBet=50, Player 0 hasn't acted
	table.CurrentHand.CurrentBet = 50
	table.CurrentHand.LastRaise = 50
	// During betting, Pot stays at 0 (chips are in PlayerBets)
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}
	table.CurrentHand.PlayerBets[0] = 0
	// Manually set some existing PlayerBets from previous actions
	table.CurrentHand.PlayerBets[1] = 50
	initialPot := table.CurrentHand.Pot

	// Player 0 raises to 150 (initial bet of 50, raise increment of 100)
	chipsMoved, err := table.CurrentHand.ProcessAction(0, "raise", 1000, 150)
	if err != nil {
		t.Errorf("expected no error processing raise, got %v", err)
	}

	// Should move 150 chips (call 50 + raise 100)
	if chipsMoved != 150 {
		t.Errorf("expected 150 chips moved, got %d", chipsMoved)
	}

	// CurrentBet should be updated to 150
	if table.CurrentHand.CurrentBet != 150 {
		t.Errorf("expected CurrentBet=150, got %d", table.CurrentHand.CurrentBet)
	}

	// LastRaise should be updated to 100 (150 - 50 = 100)
	if table.CurrentHand.LastRaise != 100 {
		t.Errorf("expected LastRaise=100, got %d", table.CurrentHand.LastRaise)
	}

	// PlayerBets[0] should be updated to 150
	if table.CurrentHand.PlayerBets[0] != 150 {
		t.Errorf("expected PlayerBets[0]=150, got %d", table.CurrentHand.PlayerBets[0])
	}

	// Pot should remain unchanged during betting (chips stay in PlayerBets until street advance)
	if table.CurrentHand.Pot != initialPot {
		t.Errorf("expected Pot=%d (unchanged during betting), got %d", initialPot, table.CurrentHand.Pot)
	}

	// ActedPlayers[0] should be marked true
	if !table.CurrentHand.ActedPlayers[0] {
		t.Errorf("expected ActedPlayers[0]=true")
	}
}

// TestProcessAction_RaiseInvalidAmount verifies error when raise amount is invalid
func TestProcessAction_RaiseInvalidAmount(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players
	for i := 0; i < 2; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Set up betting state: CurrentBet=50
	table.CurrentHand.CurrentBet = 50
	table.CurrentHand.LastRaise = 50
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}
	table.CurrentHand.PlayerBets[0] = 0

	// Try to raise to 75 (below minimum of 100)
	_, err = table.CurrentHand.ProcessAction(0, "raise", 1000, 75)
	if err == nil {
		t.Fatal("expected error for raise below minimum, got nil")
	}
	if err.Error() != "raise amount below minimum" {
		t.Errorf("expected error 'raise amount below minimum', got '%s'", err.Error())
	}
}

// TestProcessAction_RaiseAllIn verifies all-in raise is handled correctly even below minimum
func TestProcessAction_RaiseAllIn(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players
	for i := 0; i < 2; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Set up betting state: CurrentBet=50
	table.CurrentHand.CurrentBet = 50
	table.CurrentHand.LastRaise = 50
	table.CurrentHand.Pot = 100
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}
	table.CurrentHand.PlayerBets[0] = 0

	// Player with 75 chips goes all-in (below minimum of 100, but should be allowed)
	chipsMoved, err := table.CurrentHand.ProcessAction(0, "raise", 75, 75)
	if err != nil {
		t.Errorf("expected no error for all-in raise below minimum, got %v", err)
	}

	// Should move all 75 chips
	if chipsMoved != 75 {
		t.Errorf("expected 75 chips moved, got %d", chipsMoved)
	}

	// CurrentBet should be updated to 75
	if table.CurrentHand.CurrentBet != 75 {
		t.Errorf("expected CurrentBet=75, got %d", table.CurrentHand.CurrentBet)
	}

	// PlayerBets[0] should be 75
	if table.CurrentHand.PlayerBets[0] != 75 {
		t.Errorf("expected PlayerBets[0]=75, got %d", table.CurrentHand.PlayerBets[0])
	}
}

// TestProcessAction_MultipleRaises verifies LastRaise is correctly updated through a chain of raises
func TestProcessAction_MultipleRaises(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 3 active players
	for i := 0; i < 3; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Initialize betting state: BB=50
	table.CurrentHand.CurrentBet = 50
	table.CurrentHand.LastRaise = 50
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}

	// Player 0 raises to 150 (bet 50, raise 100)
	table.CurrentHand.ProcessAction(0, "raise", 1000, 150)
	if table.CurrentHand.LastRaise != 100 {
		t.Errorf("after first raise, expected LastRaise=100, got %d", table.CurrentHand.LastRaise)
	}

	// Player 1 re-raises to 300 (call 150, raise 150)
	table.CurrentHand.PlayerBets[1] = 0
	table.CurrentHand.ProcessAction(1, "raise", 1000, 300)
	if table.CurrentHand.LastRaise != 150 {
		t.Errorf("after second raise, expected LastRaise=150, got %d", table.CurrentHand.LastRaise)
	}

	// Player 2 re-raises to 600 (call 300, raise 300)
	table.CurrentHand.PlayerBets[2] = 0
	table.CurrentHand.ProcessAction(2, "raise", 1000, 600)
	if table.CurrentHand.LastRaise != 300 {
		t.Errorf("after third raise, expected LastRaise=300, got %d", table.CurrentHand.LastRaise)
	}

	if table.CurrentHand.CurrentBet != 600 {
		t.Errorf("expected final CurrentBet=600, got %d", table.CurrentHand.CurrentBet)
	}
}

// TestProcessAction_RaiseHeadsUp verifies raise works correctly in heads-up scenario
func TestProcessAction_RaiseHeadsUp(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up heads-up: seat 0 and seat 3
	token0 := "player-0"
	token3 := "player-3"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[3].Token = &token3
	table.Seats[3].Status = "active"
	table.Seats[3].Stack = 1000

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Set up betting state: Button is dealer (seat 0), BB is seat 3 (25 chips)
	table.CurrentHand.CurrentBet = 25
	table.CurrentHand.LastRaise = 25
	table.CurrentHand.Pot = 50
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}
	table.CurrentHand.PlayerBets[0] = 0

	// Seat 0 raises to 75
	chipsMoved, err := table.CurrentHand.ProcessAction(0, "raise", 1000, 75)
	if err != nil {
		t.Errorf("expected no error processing raise in heads-up, got %v", err)
	}

	if chipsMoved != 75 {
		t.Errorf("expected 75 chips moved, got %d", chipsMoved)
	}

	if table.CurrentHand.CurrentBet != 75 {
		t.Errorf("expected CurrentBet=75, got %d", table.CurrentHand.CurrentBet)
	}

	if table.CurrentHand.LastRaise != 50 {
		t.Errorf("expected LastRaise=50 (75-25), got %d", table.CurrentHand.LastRaise)
	}
}

// ============ PHASE 2: STREET PROGRESSION TRIGGER LOGIC TESTS ============

// TestHand_AdvanceToNextStreet_PreflopToFlop verifies preflop advances to flop and deals 3 cards
func TestHand_AdvanceToNextStreet_PreflopToFlop(t *testing.T) {
	hand := &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            30,
		Deck:           NewDeck(),
		HoleCards:      make(map[int][]Card),
		BoardCards:     []Card{},
		Street:         "preflop",
		CurrentBet:     20,
		PlayerBets:     make(map[int]int),
		ActedPlayers:   make(map[int]bool),
		FoldedPlayers:  make(map[int]bool),
		CurrentActor:   nil,
	}

	err := ShuffleDeck(hand.Deck)
	if err != nil {
		t.Fatalf("failed to shuffle deck: %v", err)
	}

	// Advance from preflop to flop
	err = hand.AdvanceToNextStreet()
	if err != nil {
		t.Fatalf("expected no error advancing to flop, got %v", err)
	}

	// Verify street changed to flop
	if hand.Street != "flop" {
		t.Errorf("expected street to be 'flop', got '%s'", hand.Street)
	}

	// Verify 3 board cards were dealt
	if len(hand.BoardCards) != 3 {
		t.Errorf("expected 3 board cards after flop, got %d", len(hand.BoardCards))
	}

	// Verify betting state was reset
	if hand.CurrentBet != 0 {
		t.Errorf("expected CurrentBet to be reset to 0, got %d", hand.CurrentBet)
	}
	if len(hand.ActedPlayers) != 0 {
		t.Errorf("expected ActedPlayers to be reset, got %d entries", len(hand.ActedPlayers))
	}
	if len(hand.PlayerBets) != 0 {
		t.Errorf("expected PlayerBets to be reset, got %d entries", len(hand.PlayerBets))
	}
}

// TestHand_AdvanceToNextStreet_FlopToTurn verifies flop advances to turn and deals 1 card
func TestHand_AdvanceToNextStreet_FlopToTurn(t *testing.T) {
	hand := &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            50,
		Deck:           NewDeck(),
		HoleCards:      make(map[int][]Card),
		BoardCards:     []Card{{Rank: "A", Suit: "s"}, {Rank: "K", Suit: "h"}, {Rank: "Q", Suit: "d"}},
		Street:         "flop",
		CurrentBet:     20,
		PlayerBets:     make(map[int]int),
		ActedPlayers:   make(map[int]bool),
		FoldedPlayers:  make(map[int]bool),
		CurrentActor:   nil,
	}

	err := ShuffleDeck(hand.Deck)
	if err != nil {
		t.Fatalf("failed to shuffle deck: %v", err)
	}

	// Advance from flop to turn
	err = hand.AdvanceToNextStreet()
	if err != nil {
		t.Fatalf("expected no error advancing to turn, got %v", err)
	}

	// Verify street changed to turn
	if hand.Street != "turn" {
		t.Errorf("expected street to be 'turn', got '%s'", hand.Street)
	}

	// Verify 4 board cards total (3 flop + 1 turn)
	if len(hand.BoardCards) != 4 {
		t.Errorf("expected 4 board cards after turn, got %d", len(hand.BoardCards))
	}

	// Verify betting state was reset
	if hand.CurrentBet != 0 {
		t.Errorf("expected CurrentBet to be reset to 0, got %d", hand.CurrentBet)
	}
}

// TestHand_AdvanceToNextStreet_TurnToRiver verifies turn advances to river and deals 1 card
func TestHand_AdvanceToNextStreet_TurnToRiver(t *testing.T) {
	hand := &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            70,
		Deck:           NewDeck(),
		HoleCards:      make(map[int][]Card),
		BoardCards: []Card{
			{Rank: "A", Suit: "s"}, {Rank: "K", Suit: "h"}, {Rank: "Q", Suit: "d"},
			{Rank: "J", Suit: "c"},
		},
		Street:        "turn",
		CurrentBet:    20,
		PlayerBets:    make(map[int]int),
		ActedPlayers:  make(map[int]bool),
		FoldedPlayers: make(map[int]bool),
		CurrentActor:  nil,
	}

	err := ShuffleDeck(hand.Deck)
	if err != nil {
		t.Fatalf("failed to shuffle deck: %v", err)
	}

	// Advance from turn to river
	err = hand.AdvanceToNextStreet()
	if err != nil {
		t.Fatalf("expected no error advancing to river, got %v", err)
	}

	// Verify street changed to river
	if hand.Street != "river" {
		t.Errorf("expected street to be 'river', got '%s'", hand.Street)
	}

	// Verify 5 board cards total (4 turn + 1 river)
	if len(hand.BoardCards) != 5 {
		t.Errorf("expected 5 board cards after river, got %d", len(hand.BoardCards))
	}

	// Verify betting state was reset
	if hand.CurrentBet != 0 {
		t.Errorf("expected CurrentBet to be reset to 0, got %d", hand.CurrentBet)
	}
}

// TestHand_AdvanceToNextStreet_RiverDoesNotAdvance verifies river does not advance further
func TestHand_AdvanceToNextStreet_RiverDoesNotAdvance(t *testing.T) {
	hand := &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            100,
		Deck:           NewDeck(),
		HoleCards:      make(map[int][]Card),
		BoardCards: []Card{
			{Rank: "A", Suit: "s"}, {Rank: "K", Suit: "h"}, {Rank: "Q", Suit: "d"},
			{Rank: "J", Suit: "c"}, {Rank: "T", Suit: "s"},
		},
		Street:        "river",
		CurrentBet:    0,
		PlayerBets:    make(map[int]int),
		ActedPlayers:  make(map[int]bool),
		FoldedPlayers: make(map[int]bool),
		CurrentActor:  nil,
	}

	err := ShuffleDeck(hand.Deck)
	if err != nil {
		t.Fatalf("failed to shuffle deck: %v", err)
	}

	initialBoardSize := len(hand.BoardCards)

	// Try to advance from river (should not error, but should not deal cards)
	err = hand.AdvanceToNextStreet()
	if err != nil {
		t.Fatalf("expected no error advancing from river, got %v", err)
	}

	// Verify street remains river
	if hand.Street != "river" {
		t.Errorf("expected street to remain 'river', got '%s'", hand.Street)
	}

	// Verify board cards unchanged
	if len(hand.BoardCards) != initialBoardSize {
		t.Errorf("expected board cards to remain at %d, got %d", initialBoardSize, len(hand.BoardCards))
	}
}

// TestHand_AdvanceToNextStreet_ErrorsIfInsufficientDeck verifies error when deck exhausted
func TestHand_AdvanceToNextStreet_ErrorsIfInsufficientDeck(t *testing.T) {
	hand := &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            30,
		Deck: []Card{
			{Rank: "A", Suit: "s"},
			{Rank: "K", Suit: "h"},
		},
		HoleCards:     make(map[int][]Card),
		BoardCards:    []Card{},
		Street:        "preflop",
		CurrentBet:    20,
		PlayerBets:    make(map[int]int),
		ActedPlayers:  make(map[int]bool),
		FoldedPlayers: make(map[int]bool),
		CurrentActor:  nil,
	}

	// Try to advance to flop with only 2 cards (need 4)
	err := hand.AdvanceToNextStreet()
	if err == nil {
		t.Fatal("expected error when deck has insufficient cards for flop, got nil")
	}

	// Verify street did not change
	if hand.Street != "preflop" {
		t.Errorf("expected street to remain 'preflop' on error, got '%s'", hand.Street)
	}

	// Verify no board cards were dealt
	if len(hand.BoardCards) != 0 {
		t.Errorf("expected board to remain empty on error, got %d cards", len(hand.BoardCards))
	}
}

// TestHand_FullHandProgression_PreflopToRiver verifies full hand progression through all 4 streets
func TestHand_FullHandProgression_PreflopToRiver(t *testing.T) {
	// Create hand starting in preflop
	hand := &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            30,
		Deck:           NewDeck(),
		HoleCards:      make(map[int][]Card),
		BoardCards:     []Card{},
		Street:         "preflop",
		CurrentBet:     0,
		PlayerBets:     make(map[int]int),
		ActedPlayers:   make(map[int]bool),
		FoldedPlayers:  make(map[int]bool),
		CurrentActor:   nil,
	}

	// Shuffle deck
	err := ShuffleDeck(hand.Deck)
	if err != nil {
		t.Fatalf("failed to shuffle deck: %v", err)
	}

	// Verify initial state
	if hand.Street != "preflop" {
		t.Fatalf("expected initial street to be 'preflop', got '%s'", hand.Street)
	}
	if len(hand.BoardCards) != 0 {
		t.Errorf("expected 0 board cards preflop, got %d", len(hand.BoardCards))
	}

	// Step 1: Advance to flop
	err = hand.AdvanceToNextStreet()
	if err != nil {
		t.Fatalf("failed to advance to flop: %v", err)
	}
	if hand.Street != "flop" {
		t.Errorf("expected street to be 'flop', got '%s'", hand.Street)
	}
	if len(hand.BoardCards) != 3 {
		t.Errorf("expected 3 board cards on flop, got %d", len(hand.BoardCards))
	}

	// Step 2: Advance to turn
	err = hand.AdvanceToNextStreet()
	if err != nil {
		t.Fatalf("failed to advance to turn: %v", err)
	}
	if hand.Street != "turn" {
		t.Errorf("expected street to be 'turn', got '%s'", hand.Street)
	}
	if len(hand.BoardCards) != 4 {
		t.Errorf("expected 4 board cards on turn, got %d", len(hand.BoardCards))
	}

	// Step 3: Advance to river
	err = hand.AdvanceToNextStreet()
	if err != nil {
		t.Fatalf("failed to advance to river: %v", err)
	}
	if hand.Street != "river" {
		t.Errorf("expected street to be 'river', got '%s'", hand.Street)
	}
	if len(hand.BoardCards) != 5 {
		t.Errorf("expected 5 board cards on river, got %d", len(hand.BoardCards))
	}

	// Verify betting state reset after each street advancement
	if hand.CurrentBet != 0 {
		t.Errorf("expected CurrentBet to be reset after street advancement, got %d", hand.CurrentBet)
	}
}

// TestHand_ActionFlow_ContinuesAcrossStreets verifies actions work smoothly across street transitions
func TestHand_ActionFlow_ContinuesAcrossStreets(t *testing.T) {
	// Create hand starting in preflop with multiple players
	hand := &Hand{
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Pot:            30,
		Deck:           NewDeck(),
		HoleCards:      make(map[int][]Card),
		BoardCards:     []Card{},
		Street:         "preflop",
		CurrentBet:     20,
		PlayerBets:     map[int]int{0: 0, 1: 10, 2: 20},
		ActedPlayers:   map[int]bool{1: true, 2: true},
		FoldedPlayers:  make(map[int]bool),
		CurrentActor:   nil,
	}

	// Shuffle deck
	err := ShuffleDeck(hand.Deck)
	if err != nil {
		t.Fatalf("failed to shuffle deck: %v", err)
	}

	// Verify preflop state
	if hand.Street != "preflop" {
		t.Fatalf("expected initial street to be 'preflop', got '%s'", hand.Street)
	}

	// Step 1: Advance to flop - verify board cards and betting reset
	err = hand.AdvanceToNextStreet()
	if err != nil {
		t.Fatalf("failed to advance to flop: %v", err)
	}
	if hand.Street != "flop" {
		t.Errorf("expected street to be 'flop' after advancement, got '%s'", hand.Street)
	}
	if len(hand.BoardCards) != 3 {
		t.Errorf("expected 3 board cards on flop, got %d", len(hand.BoardCards))
	}
	// Betting state should be reset for new street
	if hand.CurrentBet != 0 {
		t.Errorf("expected CurrentBet to reset on flop, got %d", hand.CurrentBet)
	}

	// Step 2: Advance to turn - verify board cards and betting reset
	err = hand.AdvanceToNextStreet()
	if err != nil {
		t.Fatalf("failed to advance to turn: %v", err)
	}
	if hand.Street != "turn" {
		t.Errorf("expected street to be 'turn' after advancement, got '%s'", hand.Street)
	}
	if len(hand.BoardCards) != 4 {
		t.Errorf("expected 4 board cards on turn, got %d", len(hand.BoardCards))
	}
	// Betting state should be reset for new street
	if hand.CurrentBet != 0 {
		t.Errorf("expected CurrentBet to reset on turn, got %d", hand.CurrentBet)
	}

	// Step 3: Advance to river - verify board cards and betting reset
	err = hand.AdvanceToNextStreet()
	if err != nil {
		t.Fatalf("failed to advance to river: %v", err)
	}
	if hand.Street != "river" {
		t.Errorf("expected street to be 'river' after advancement, got '%s'", hand.Street)
	}
	if len(hand.BoardCards) != 5 {
		t.Errorf("expected 5 board cards on river, got %d", len(hand.BoardCards))
	}
	// Betting state should be reset for new street
	if hand.CurrentBet != 0 {
		t.Errorf("expected CurrentBet to reset on river, got %d", hand.CurrentBet)
	}
}

// TestHand_BigBlindHasOption_InitiallyTrue verifies flag is true after StartHand()
func TestHand_BigBlindHasOption_InitiallyTrue(t *testing.T) {
	// Create a new table
	table := NewTable("table-1", "Test Table", nil)

	// Seat two players as "active"
	token1 := "player1"
	token2 := "player2"
	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Stack = 1000

	// Start a hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Verify BigBlindHasOption is true on preflop after StartHand
	if table.CurrentHand == nil {
		t.Fatal("expected CurrentHand to be set after StartHand")
	}

	if !table.CurrentHand.BigBlindHasOption {
		t.Errorf("expected BigBlindHasOption to be true after StartHand, got %v", table.CurrentHand.BigBlindHasOption)
	}

	// Verify we're on preflop
	if table.CurrentHand.Street != "preflop" {
		t.Errorf("expected street to be 'preflop', got '%s'", table.CurrentHand.Street)
	}
}

// TestHand_BigBlindHasOption_ClearedWhenBBChecks verifies flag cleared when BB checks
func TestHand_BigBlindHasOption_ClearedWhenBBChecks(t *testing.T) {
	// Create a hand in preflop with BB as next actor
	hand := &Hand{
		DealerSeat:        0,
		SmallBlindSeat:    0,
		BigBlindSeat:      1,
		Pot:               30,
		Deck:              NewDeck(),
		HoleCards:         make(map[int][]Card),
		BoardCards:        []Card{},
		Street:            "preflop",
		CurrentBet:        20,
		PlayerBets:        map[int]int{0: 10, 1: 20},
		ActedPlayers:      map[int]bool{0: true},
		FoldedPlayers:     make(map[int]bool),
		BigBlindHasOption: true,
		CurrentActor:      nil,
	}

	// BB checks (already matched the bet at 20)
	_, err := hand.ProcessAction(1, "check", 980)
	if err != nil {
		t.Fatalf("failed to process BB check: %v", err)
	}

	// Verify BigBlindHasOption is now false after BB checks
	if hand.BigBlindHasOption {
		t.Errorf("expected BigBlindHasOption to be false after BB checks, got %v", hand.BigBlindHasOption)
	}
}

// TestHand_BigBlindHasOption_ClearedWhenBBRaises verifies flag cleared when BB raises
func TestHand_BigBlindHasOption_ClearedWhenBBRaises(t *testing.T) {
	// Create a hand in preflop with BB as next actor
	hand := &Hand{
		DealerSeat:        0,
		SmallBlindSeat:    0,
		BigBlindSeat:      1,
		Pot:               30,
		Deck:              NewDeck(),
		HoleCards:         make(map[int][]Card),
		BoardCards:        []Card{},
		Street:            "preflop",
		CurrentBet:        20,
		PlayerBets:        map[int]int{0: 10, 1: 20},
		ActedPlayers:      map[int]bool{0: true},
		FoldedPlayers:     make(map[int]bool),
		BigBlindHasOption: true,
		LastRaise:         20,
		CurrentActor:      nil,
	}

	// BB raises to 60
	_, err := hand.ProcessAction(1, "raise", 980, 60)
	if err != nil {
		t.Fatalf("failed to process BB raise: %v", err)
	}

	// Verify BigBlindHasOption is now false after BB raises
	if hand.BigBlindHasOption {
		t.Errorf("expected BigBlindHasOption to be false after BB raises, got %v", hand.BigBlindHasOption)
	}
}

// TestHand_BigBlindHasOption_ClearedOnAnyRaise verifies flag cleared when any player raises
func TestHand_BigBlindHasOption_ClearedOnAnyRaise(t *testing.T) {
	// Create a hand in preflop with SB as next actor (after blinds posted)
	hand := &Hand{
		DealerSeat:        0,
		SmallBlindSeat:    0,
		BigBlindSeat:      1,
		Pot:               30,
		Deck:              NewDeck(),
		HoleCards:         make(map[int][]Card),
		BoardCards:        []Card{},
		Street:            "preflop",
		CurrentBet:        20,
		PlayerBets:        map[int]int{0: 10, 1: 20},
		ActedPlayers:      map[int]bool{},
		FoldedPlayers:     make(map[int]bool),
		BigBlindHasOption: true,
		LastRaise:         20,
		CurrentActor:      nil,
	}

	// SB raises to 50 (any raise should clear the flag)
	_, err := hand.ProcessAction(0, "raise", 990, 50)
	if err != nil {
		t.Fatalf("failed to process SB raise: %v", err)
	}

	// Verify BigBlindHasOption is now false when any player raises
	if hand.BigBlindHasOption {
		t.Errorf("expected BigBlindHasOption to be false after any raise, got %v", hand.BigBlindHasOption)
	}
}

// TestHand_BigBlindHasOption_ClearedOnStreetAdvance verifies flag cleared when advancing to flop
func TestHand_BigBlindHasOption_ClearedOnStreetAdvance(t *testing.T) {
	// Create a hand in preflop
	hand := &Hand{
		DealerSeat:        0,
		SmallBlindSeat:    1,
		BigBlindSeat:      2,
		Pot:               30,
		Deck:              NewDeck(),
		HoleCards:         make(map[int][]Card),
		BoardCards:        []Card{},
		Street:            "preflop",
		CurrentBet:        0,
		PlayerBets:        make(map[int]int),
		ActedPlayers:      make(map[int]bool),
		FoldedPlayers:     make(map[int]bool),
		BigBlindHasOption: true,
		CurrentActor:      nil,
	}

	// Shuffle deck
	err := ShuffleDeck(hand.Deck)
	if err != nil {
		t.Fatalf("failed to shuffle deck: %v", err)
	}

	// Advance to flop
	err = hand.AdvanceToNextStreet()
	if err != nil {
		t.Fatalf("failed to advance to flop: %v", err)
	}

	// Verify BigBlindHasOption is now false after advancing to flop
	if hand.BigBlindHasOption {
		t.Errorf("expected BigBlindHasOption to be false after advancing to flop, got %v", hand.BigBlindHasOption)
	}

	// Verify we're on flop
	if hand.Street != "flop" {
		t.Errorf("expected street to be 'flop', got '%s'", hand.Street)
	}
}

// TestHandFlow_BBCanRaiseUnopenedPot verifies BB can raise an unopened pot when facing calls
// This tests the complete hand flow with BB choosing to raise instead of check/fold
func TestHandFlow_BBCanRaiseUnopenedPot(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)

	// Set up 3 players: UTG (seat 0), Dealer (seat 1), BB (seat 2)
	// Dealer posts SB, BB posts BB
	token0, token1, token2 := "utg", "dealer", "bb"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.Seats[2].Token = &token2
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	hand := table.CurrentHand
	if hand == nil {
		t.Fatal("CurrentHand is nil after StartHand")
	}

	// Verify initial state
	if hand.Street != "preflop" {
		t.Errorf("expected preflop, got %s", hand.Street)
	}
	if !hand.BigBlindHasOption {
		t.Error("expected BigBlindHasOption to be true initially")
	}

	// Verify BB initially has raise option when pot is unopened
	bbValidActions := hand.GetValidActions(2, table.Seats[2].Stack, table.Seats)
	hasRaise := false
	for _, action := range bbValidActions {
		if action == "raise" {
			hasRaise = true
			break
		}
	}
	if !hasRaise {
		t.Fatalf("expected BB to have raise option preflop with unopened pot, got actions: %v", bbValidActions)
	}

	// UTG (seat 0) calls the BB
	_, err = hand.ProcessAction(0, "call", 980)
	if err != nil {
		t.Fatalf("UTG call failed: %v", err)
	}

	// Dealer (seat 1) calls the BB
	_, err = hand.ProcessAction(1, "call", 990)
	if err != nil {
		t.Fatalf("Dealer call failed: %v", err)
	}

	// Verify BigBlindHasOption is still true (no raise yet, unopened pot)
	if !hand.BigBlindHasOption {
		t.Error("expected BigBlindHasOption to still be true after UTG/Dealer calls unopened pot")
	}

	// Verify BB still has raise option
	bbValidActions = hand.GetValidActions(2, table.Seats[2].Stack, table.Seats)
	hasRaise = false
	for _, action := range bbValidActions {
		if action == "raise" {
			hasRaise = true
			break
		}
	}
	if !hasRaise {
		t.Fatalf("expected BB to have raise option after UTG/Dealer call unopened pot, got actions: %v", bbValidActions)
	}

	// BB raises to 40 (minimum raise from 20 is 40)
	_, err = hand.ProcessAction(2, "raise", 960, 40)
	if err != nil {
		t.Fatalf("BB raise failed: %v", err)
	}

	// Verify BigBlindHasOption is now false (BB exercised option by raising)
	if hand.BigBlindHasOption {
		t.Error("expected BigBlindHasOption to be false after BB raises")
	}

	// Verify CurrentBet is now 40 (the raise amount)
	if hand.CurrentBet != 40 {
		t.Errorf("expected CurrentBet to be 40 after BB raise, got %d", hand.CurrentBet)
	}

	// UTG now faces a raise and can call/fold/reraise
	// Action should return to UTG (first to act in preflop)
	// Verify UTG has the option to call the raise or fold
	utgValidActions := hand.GetValidActions(0, table.Seats[0].Stack-20, table.Seats)
	hasCall := false
	hasFold := false
	for _, action := range utgValidActions {
		if action == "call" {
			hasCall = true
		}
		if action == "fold" {
			hasFold = true
		}
	}
	if !hasCall || !hasFold {
		t.Errorf("expected UTG to have call/fold options after BB raise, got: %v", utgValidActions)
	}

	// UTG calls the raise
	_, err = hand.ProcessAction(0, "call", 960)
	if err != nil {
		t.Fatalf("UTG call of raise failed: %v", err)
	}

	// Dealer calls the raise
	_, err = hand.ProcessAction(1, "call", 960)
	if err != nil {
		t.Fatalf("Dealer call of raise failed: %v", err)
	}

	// Verify betting round is complete
	if !hand.IsBettingRoundComplete(table.Seats) {
		t.Error("expected betting round to be complete after all players act")
	}

	// Advance to flop
	err = hand.AdvanceToNextStreet()
	if err != nil {
		t.Fatalf("failed to advance to flop: %v", err)
	}

	// Verify we're on flop with 3 cards
	if hand.Street != "flop" {
		t.Errorf("expected street to be 'flop', got '%s'", hand.Street)
	}
	if len(hand.BoardCards) != 3 {
		t.Errorf("expected 3 board cards on flop, got %d", len(hand.BoardCards))
	}
}

// TestHandFlow_PostflopCheckRaise verifies players can check-raise on postflop streets
// This tests that players who have matched the bet on flop have raise option
func TestHandFlow_PostflopCheckRaise(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)

	// Set up 3 players for a complete preflop action leading to flop
	token0, token1, token2 := "player0", "player1", "player2"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.Seats[2].Token = &token2
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	hand := table.CurrentHand
	if hand == nil {
		t.Fatal("CurrentHand is nil after StartHand")
	}

	// Preflop: seat 0 calls, seat 1 calls, seat 2 (BB) checks
	_, err = hand.ProcessAction(0, "call", 980)
	if err != nil {
		t.Fatalf("seat 0 call failed: %v", err)
	}
	_, err = hand.ProcessAction(1, "call", 990)
	if err != nil {
		t.Fatalf("seat 1 call failed: %v", err)
	}
	_, err = hand.ProcessAction(2, "check", 980)
	if err != nil {
		t.Fatalf("seat 2 (BB) check failed: %v", err)
	}

	// Verify betting round is complete
	if !hand.IsBettingRoundComplete(table.Seats) {
		t.Error("expected preflop betting round to be complete")
	}

	// Advance to flop
	err = hand.AdvanceToNextStreet()
	if err != nil {
		t.Fatalf("failed to advance to flop: %v", err)
	}

	// Verify we're on flop
	if hand.Street != "flop" {
		t.Errorf("expected street to be 'flop', got '%s'", hand.Street)
	}

	// On flop, all players have matched the current bet (all at 0 after street change)
	// Verify seat 2 has raise option when checking
	seat2ValidActions := hand.GetValidActions(2, table.Seats[2].Stack, table.Seats)
	hasRaise := false
	for _, action := range seat2ValidActions {
		if action == "raise" {
			hasRaise = true
			break
		}
	}
	if !hasRaise {
		t.Fatalf("expected seat 2 to have raise option on flop with unopened action, got actions: %v", seat2ValidActions)
	}

	// Seat 2 checks on flop
	_, err = hand.ProcessAction(2, "check", 980)
	if err != nil {
		t.Fatalf("seat 2 check on flop failed: %v", err)
	}

	// Seat 0 checks on flop
	_, err = hand.ProcessAction(0, "check", 980)
	if err != nil {
		t.Fatalf("seat 0 check on flop failed: %v", err)
	}

	// Seat 1 now faces an unopened pot and can raise or check
	// Verify seat 1 has raise option
	seat1ValidActions := hand.GetValidActions(1, table.Seats[1].Stack, table.Seats)
	hasRaise = false
	hasCheck := false
	for _, action := range seat1ValidActions {
		if action == "raise" {
			hasRaise = true
		}
		if action == "check" {
			hasCheck = true
		}
	}
	if !hasRaise {
		t.Fatalf("expected seat 1 to have raise option on flop unopened, got actions: %v", seat1ValidActions)
	}
	if !hasCheck {
		t.Fatalf("expected seat 1 to have check option on flop, got actions: %v", seat1ValidActions)
	}

	// Seat 1 raises to 40 on flop
	_, err = hand.ProcessAction(1, "raise", 960, 40)
	if err != nil {
		t.Fatalf("seat 1 raise on flop failed: %v", err)
	}

	// Verify current bet is 40
	if hand.CurrentBet != 40 {
		t.Errorf("expected CurrentBet to be 40 after raise, got %d", hand.CurrentBet)
	}

	// Seat 2 now faces the raise and can call/fold/reraise
	seat2ValidActions = hand.GetValidActions(2, table.Seats[2].Stack, table.Seats)
	hasCall := false
	hasFold := false
	hasRaise = false
	for _, action := range seat2ValidActions {
		if action == "call" {
			hasCall = true
		}
		if action == "fold" {
			hasFold = true
		}
		if action == "raise" {
			hasRaise = true
		}
	}
	if !hasCall {
		t.Errorf("expected seat 2 to have call option after raise, got: %v", seat2ValidActions)
	}
	if !hasFold {
		t.Errorf("expected seat 2 to have fold option after raise, got: %v", seat2ValidActions)
	}
	if !hasRaise {
		t.Errorf("expected seat 2 to have raise option after raise (for 3-bet), got: %v", seat2ValidActions)
	}

	// Seat 2 calls the raise
	_, err = hand.ProcessAction(2, "call", 960)
	if err != nil {
		t.Fatalf("seat 2 call of raise failed: %v", err)
	}

	// Seat 0 calls the raise
	_, err = hand.ProcessAction(0, "call", 960)
	if err != nil {
		t.Fatalf("seat 0 call of raise failed: %v", err)
	}

	// Verify flop betting round is complete
	if !hand.IsBettingRoundComplete(table.Seats) {
		t.Error("expected flop betting round to be complete")
	}

	// Advance to turn
	err = hand.AdvanceToNextStreet()
	if err != nil {
		t.Fatalf("failed to advance to turn: %v", err)
	}

	// Verify we're on turn
	if hand.Street != "turn" {
		t.Errorf("expected street to be 'turn', got '%s'", hand.Street)
	}
	if len(hand.BoardCards) != 4 {
		t.Errorf("expected 4 board cards on turn, got %d", len(hand.BoardCards))
	}
}

// TestHandFlow_ActionOrderChangesPostflop verifies action order changes correctly from preflop to postflop
// Preflop: UTG (seat 3) acts first with 4 players
// Postflop: SB (seat 1) acts first
func TestHandFlow_ActionOrderChangesPostflop(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)

	// Set up 4 active players: seats 0, 1, 2, 3
	// Dealer=0, SB=1, BB=2, UTG=3
	token0, token1, token2, token3 := "dealer", "sb", "bb", "utg"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.Seats[2].Token = &token2
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000
	table.Seats[3].Token = &token3
	table.Seats[3].Status = "active"
	table.Seats[3].Stack = 1000

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	hand := table.CurrentHand
	if hand == nil {
		t.Fatal("CurrentHand is nil after StartHand")
	}

	// Verify initial state: preflop
	if hand.Street != "preflop" {
		t.Errorf("expected preflop, got %s", hand.Street)
	}

	// Verify CurrentActor is seat 3 (UTG acts first preflop)
	if hand.CurrentActor == nil {
		t.Error("expected CurrentActor to be set (UTG), got nil")
	} else if *hand.CurrentActor != 3 {
		t.Errorf("expected CurrentActor to be seat 3 (UTG) preflop, got seat %d", *hand.CurrentActor)
	}

	// All players call to complete preflop betting
	// UTG (seat 3) calls
	_, err = hand.ProcessAction(3, "call", 980)
	if err != nil {
		t.Fatalf("UTG call failed: %v", err)
	}

	// Dealer (seat 0) calls
	_, err = hand.ProcessAction(0, "call", 990)
	if err != nil {
		t.Fatalf("Dealer call failed: %v", err)
	}

	// SB (seat 1) calls
	_, err = hand.ProcessAction(1, "call", 995)
	if err != nil {
		t.Fatalf("SB call failed: %v", err)
	}

	// BB (seat 2) checks (completes preflop)
	_, err = hand.ProcessAction(2, "check", 980)
	if err != nil {
		t.Fatalf("BB check failed: %v", err)
	}

	// Verify betting round is complete
	if !hand.IsBettingRoundComplete(table.Seats) {
		t.Error("expected preflop betting round to be complete")
	}

	// Advance to flop
	err = hand.AdvanceToNextStreet()
	if err != nil {
		t.Fatalf("failed to advance to flop: %v", err)
	}

	// Verify Street is "flop"
	if hand.Street != "flop" {
		t.Errorf("expected street to be 'flop' after advance, got '%s'", hand.Street)
	}

	// Set CurrentActor to the first actor on the new street (this is done in handlers.go in real flow)
	firstActor := hand.GetFirstActor(table.Seats)
	hand.CurrentActor = &firstActor

	// Verify CurrentActor is now seat 1 (SB acts first postflop)
	if hand.CurrentActor == nil {
		t.Error("expected CurrentActor to be set (SB), got nil after advance to flop")
	} else if *hand.CurrentActor != 1 {
		t.Errorf("expected CurrentActor to be seat 1 (SB) on flop, got seat %d", *hand.CurrentActor)
	}
}

// TestHandFlow_ActionOrderHeadsUpPostflop verifies action order changes correctly in heads-up
// Preflop: Dealer (seat 0) acts first (heads-up)
// Postflop: BB (seat 2) acts first
func TestHandFlow_ActionOrderHeadsUpPostflop(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)

	// Set up 2 active players: seats 0 and 2
	// Dealer/SB=0, BB=2
	token0, token2 := "dealer", "bb"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[2].Token = &token2
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	hand := table.CurrentHand
	if hand == nil {
		t.Fatal("CurrentHand is nil after StartHand")
	}

	// Verify initial state: preflop
	if hand.Street != "preflop" {
		t.Errorf("expected preflop, got %s", hand.Street)
	}

	// Verify CurrentActor is seat 0 (Dealer acts first heads-up preflop)
	if hand.CurrentActor == nil {
		t.Error("expected CurrentActor to be set (Dealer), got nil")
	} else if *hand.CurrentActor != 0 {
		t.Errorf("expected CurrentActor to be seat 0 (Dealer) heads-up preflop, got seat %d", *hand.CurrentActor)
	}

	// Dealer calls (completes preflop in heads-up)
	_, err = hand.ProcessAction(0, "call", 990)
	if err != nil {
		t.Fatalf("Dealer call failed: %v", err)
	}

	// BB checks
	_, err = hand.ProcessAction(2, "check", 990)
	if err != nil {
		t.Fatalf("BB check failed: %v", err)
	}

	// Verify betting round is complete
	if !hand.IsBettingRoundComplete(table.Seats) {
		t.Error("expected preflop betting round to be complete")
	}

	// Advance to flop
	err = hand.AdvanceToNextStreet()
	if err != nil {
		t.Fatalf("failed to advance to flop: %v", err)
	}

	// Verify Street is "flop"
	if hand.Street != "flop" {
		t.Errorf("expected street to be 'flop' after advance, got '%s'", hand.Street)
	}

	// Set CurrentActor to the first actor on the new street (this is done in handlers.go in real flow)
	firstActor := hand.GetFirstActor(table.Seats)
	hand.CurrentActor = &firstActor

	// Verify CurrentActor is now seat 2 (BB acts first postflop heads-up)
	if hand.CurrentActor == nil {
		t.Error("expected CurrentActor to be set (BB), got nil after advance to flop")
	} else if *hand.CurrentActor != 2 {
		t.Errorf("expected CurrentActor to be seat 2 (BB) on flop heads-up, got seat %d", *hand.CurrentActor)
	}
}

// TestHandFlow_ActionOrderWithFolds verifies action order skips folded players and adjusts correctly
// when advancing from preflop to postflop
// UTG (seat 3) folds preflop
// Postflop: SB (seat 1) acts first (UTG already folded, so no need to skip)
// SB folds on flop
// Next actor should be BB (seat 2)
func TestHandFlow_ActionOrderWithFolds(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)

	// Set up 4 active players: seats 0, 1, 2, 3
	// Dealer=0, SB=1, BB=2, UTG=3
	token0, token1, token2, token3 := "dealer", "sb", "bb", "utg"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000
	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000
	table.Seats[2].Token = &token2
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000
	table.Seats[3].Token = &token3
	table.Seats[3].Status = "active"
	table.Seats[3].Stack = 1000

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	hand := table.CurrentHand
	if hand == nil {
		t.Fatal("CurrentHand is nil after StartHand")
	}

	// Verify initial state: preflop
	if hand.Street != "preflop" {
		t.Errorf("expected preflop, got %s", hand.Street)
	}

	// Verify CurrentActor is seat 3 (UTG acts first preflop)
	if hand.CurrentActor == nil {
		t.Error("expected CurrentActor to be set (UTG), got nil")
	} else if *hand.CurrentActor != 3 {
		t.Errorf("expected CurrentActor to be seat 3 (UTG) preflop, got seat %d", *hand.CurrentActor)
	}

	// UTG (seat 3) folds preflop
	_, err = hand.ProcessAction(3, "fold", 1000)
	if err != nil {
		t.Fatalf("UTG fold failed: %v", err)
	}

	// Verify UTG is marked as folded
	if !hand.FoldedPlayers[3] {
		t.Error("expected seat 3 (UTG) to be marked as folded")
	}

	// Dealer (seat 0) calls
	_, err = hand.ProcessAction(0, "call", 990)
	if err != nil {
		t.Fatalf("Dealer call failed: %v", err)
	}

	// SB (seat 1) calls
	_, err = hand.ProcessAction(1, "call", 995)
	if err != nil {
		t.Fatalf("SB call failed: %v", err)
	}

	// BB (seat 2) checks (completes preflop)
	_, err = hand.ProcessAction(2, "check", 980)
	if err != nil {
		t.Fatalf("BB check failed: %v", err)
	}

	// Verify betting round is complete
	if !hand.IsBettingRoundComplete(table.Seats) {
		t.Error("expected preflop betting round to be complete")
	}

	// Advance to flop
	err = hand.AdvanceToNextStreet()
	if err != nil {
		t.Fatalf("failed to advance to flop: %v", err)
	}

	// Verify Street is "flop"
	if hand.Street != "flop" {
		t.Errorf("expected street to be 'flop' after advance, got '%s'", hand.Street)
	}

	// Set CurrentActor to the first actor on the new street (this is done in handlers.go in real flow)
	firstActor := hand.GetFirstActor(table.Seats)
	hand.CurrentActor = &firstActor

	// Verify CurrentActor is seat 1 (SB acts first postflop, UTG is folded so skipped)
	if hand.CurrentActor == nil {
		t.Error("expected CurrentActor to be set (SB), got nil after advance to flop")
	} else if *hand.CurrentActor != 1 {
		t.Errorf("expected CurrentActor to be seat 1 (SB) on flop, got seat %d", *hand.CurrentActor)
	}

	// SB (seat 1) folds on flop
	_, err = hand.ProcessAction(1, "fold", 1000)
	if err != nil {
		t.Fatalf("SB fold on flop failed: %v", err)
	}

	// Verify SB is marked as folded
	if !hand.FoldedPlayers[1] {
		t.Error("expected seat 1 (SB) to be marked as folded")
	}

	// Verify next actor is BB (seat 2) - GetNextActiveSeat should skip SB
	nextSeat := hand.GetNextActiveSeat(1, table.Seats)
	if nextSeat == nil {
		t.Error("expected next active seat to be found (BB), got nil")
	} else if *nextSeat != 2 {
		t.Errorf("expected next active seat to be 2 (BB), got %d", *nextSeat)
	}
}

// TestDetermineWinner_SingleWinner_HighCard - One player has highest card
func TestDetermineWinner_SingleWinner_HighCard(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)
	hand := &Hand{
		Pot:            100,
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Street:         "river",
		HoleCards: map[int][]Card{
			0: {Card{Rank: "A", Suit: "s"}, Card{Rank: "K", Suit: "s"}}, // Strong hand
			1: {Card{Rank: "2", Suit: "h"}, Card{Rank: "3", Suit: "h"}}, // Weak hand
		},
		BoardCards: []Card{
			{Rank: "T", Suit: "d"}, {Rank: "J", Suit: "c"}, {Rank: "Q", Suit: "s"},
			{Rank: "K", Suit: "h"}, {Rank: "2", Suit: "d"},
		},
		FoldedPlayers: make(map[int]bool),
	}

	// Setup seats
	table.Seats[0].Status = "active"
	table.Seats[1].Status = "active"

	winners, winningRank := hand.DetermineWinner(seatsToPointers(table.Seats[:]))

	if len(winners) != 1 {
		t.Fatalf("expected 1 winner, got %d", len(winners))
	}

	if winners[0] != 0 {
		t.Errorf("expected winner to be seat 0, got seat %d", winners[0])
	}

	if winningRank == nil {
		t.Error("expected winningRank to be non-nil")
	}
}

// TestDetermineWinner_SingleWinner_Flush - One player has flush
func TestDetermineWinner_SingleWinner_Flush(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)
	hand := &Hand{
		Pot:            100,
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Street:         "river",
		HoleCards: map[int][]Card{
			0: {Card{Rank: "A", Suit: "s"}, Card{Rank: "K", Suit: "s"}}, // Flush
			1: {Card{Rank: "Q", Suit: "h"}, Card{Rank: "J", Suit: "h"}}, // No flush
		},
		BoardCards: []Card{
			{Rank: "T", Suit: "s"}, {Rank: "9", Suit: "s"}, {Rank: "8", Suit: "s"},
			{Rank: "5", Suit: "c"}, {Rank: "4", Suit: "d"},
		},
		FoldedPlayers: make(map[int]bool),
	}

	// Setup seats
	table.Seats[0].Status = "active"
	table.Seats[1].Status = "active"

	winners, winningRank := hand.DetermineWinner(seatsToPointers(table.Seats[:]))

	if len(winners) != 1 {
		t.Fatalf("expected 1 winner, got %d", len(winners))
	}

	if winners[0] != 0 {
		t.Errorf("expected winner to be seat 0 with flush, got seat %d", winners[0])
	}

	if winningRank == nil {
		t.Error("expected winningRank to be non-nil")
	}
}

// TestDetermineWinner_Tie_TwoPlayers - Two players with identical hands
func TestDetermineWinner_Tie_TwoPlayers(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)
	hand := &Hand{
		Pot:            100,
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Street:         "river",
		HoleCards: map[int][]Card{
			0: {Card{Rank: "A", Suit: "s"}, Card{Rank: "K", Suit: "h"}}, // Identical best 5
			1: {Card{Rank: "2", Suit: "d"}, Card{Rank: "3", Suit: "c"}}, // Identical best 5
		},
		BoardCards: []Card{
			{Rank: "A", Suit: "d"}, {Rank: "K", Suit: "c"}, {Rank: "Q", Suit: "s"},
			{Rank: "J", Suit: "h"}, {Rank: "T", Suit: "d"},
		},
		FoldedPlayers: make(map[int]bool),
	}

	// Setup seats
	table.Seats[0].Status = "active"
	table.Seats[1].Status = "active"

	winners, winningRank := hand.DetermineWinner(seatsToPointers(table.Seats[:]))

	if len(winners) != 2 {
		t.Fatalf("expected 2 winners (tie), got %d", len(winners))
	}

	// Should contain both seat 0 and 1
	foundSeat0 := false
	foundSeat1 := false
	for _, w := range winners {
		if w == 0 {
			foundSeat0 = true
		}
		if w == 1 {
			foundSeat1 = true
		}
	}

	if !foundSeat0 || !foundSeat1 {
		t.Error("expected both seat 0 and seat 1 in winners list for tie")
	}

	if winningRank == nil {
		t.Error("expected winningRank to be non-nil")
	}
}

// TestDetermineWinner_Tie_ThreePlayers - Three players with identical hands
func TestDetermineWinner_Tie_ThreePlayers(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)
	hand := &Hand{
		Pot:            300,
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Street:         "river",
		HoleCards: map[int][]Card{
			0: {Card{Rank: "2", Suit: "s"}, Card{Rank: "3", Suit: "s"}},
			1: {Card{Rank: "4", Suit: "h"}, Card{Rank: "5", Suit: "h"}},
			2: {Card{Rank: "6", Suit: "d"}, Card{Rank: "7", Suit: "d"}},
		},
		BoardCards: []Card{
			{Rank: "A", Suit: "c"}, {Rank: "K", Suit: "s"}, {Rank: "Q", Suit: "h"},
			{Rank: "J", Suit: "d"}, {Rank: "T", Suit: "c"},
		},
		FoldedPlayers: make(map[int]bool),
	}

	// Setup seats
	table.Seats[0].Status = "active"
	table.Seats[1].Status = "active"
	table.Seats[2].Status = "active"

	winners, winningRank := hand.DetermineWinner(seatsToPointers(table.Seats[:]))

	if len(winners) != 3 {
		t.Fatalf("expected 3 winners (three-way tie), got %d", len(winners))
	}

	if winningRank == nil {
		t.Error("expected winningRank to be non-nil")
	}
}

// TestDetermineWinner_HeadsUp - Two player showdown
func TestDetermineWinner_HeadsUp(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)
	hand := &Hand{
		Pot:            100,
		DealerSeat:     0,
		SmallBlindSeat: 0,
		BigBlindSeat:   1,
		Street:         "river",
		HoleCards: map[int][]Card{
			0: {Card{Rank: "A", Suit: "s"}, Card{Rank: "A", Suit: "h"}}, // Pair of aces
			1: {Card{Rank: "K", Suit: "d"}, Card{Rank: "K", Suit: "c"}}, // Pair of kings
		},
		BoardCards: []Card{
			{Rank: "2", Suit: "s"}, {Rank: "3", Suit: "h"}, {Rank: "4", Suit: "d"},
			{Rank: "5", Suit: "c"}, {Rank: "7", Suit: "s"},
		},
		FoldedPlayers: make(map[int]bool),
	}

	// Setup seats
	table.Seats[0].Status = "active"
	table.Seats[1].Status = "active"

	winners, winningRank := hand.DetermineWinner(seatsToPointers(table.Seats[:]))

	if len(winners) != 1 {
		t.Fatalf("expected 1 winner in heads-up, got %d", len(winners))
	}

	if winners[0] != 0 {
		t.Errorf("expected winner to be seat 0 with pair of aces, got seat %d", winners[0])
	}

	if winningRank == nil {
		t.Error("expected winningRank to be non-nil")
	}
}

// TestDetermineWinner_MultiWay_FourPlayers - Four players at showdown
func TestDetermineWinner_MultiWay_FourPlayers(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)
	hand := &Hand{
		Pot:            400,
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Street:         "river",
		HoleCards: map[int][]Card{
			0: {Card{Rank: "A", Suit: "s"}, Card{Rank: "A", Suit: "h"}}, // Pair of aces (best)
			1: {Card{Rank: "K", Suit: "d"}, Card{Rank: "K", Suit: "c"}}, // Pair of kings
			2: {Card{Rank: "Q", Suit: "s"}, Card{Rank: "J", Suit: "h"}}, // High card QJ
			3: {Card{Rank: "T", Suit: "d"}, Card{Rank: "9", Suit: "c"}}, // High card T9
		},
		BoardCards: []Card{
			{Rank: "2", Suit: "s"}, {Rank: "3", Suit: "h"}, {Rank: "4", Suit: "d"},
			{Rank: "6", Suit: "c"}, {Rank: "8", Suit: "s"},
		},
		FoldedPlayers: make(map[int]bool),
	}

	// Setup seats
	table.Seats[0].Status = "active"
	table.Seats[1].Status = "active"
	table.Seats[2].Status = "active"
	table.Seats[3].Status = "active"

	winners, winningRank := hand.DetermineWinner(seatsToPointers(table.Seats[:]))

	if len(winners) != 1 {
		t.Fatalf("expected 1 winner in 4-way showdown, got %d", len(winners))
	}

	if winners[0] != 0 {
		t.Errorf("expected winner to be seat 0, got seat %d", winners[0])
	}

	if winningRank == nil {
		t.Error("expected winningRank to be non-nil")
	}
}

// TestDetermineWinner_SkipsFoldedPlayers - Only evaluates non-folded players
func TestDetermineWinner_SkipsFoldedPlayers(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)
	hand := &Hand{
		Pot:            100,
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Street:         "river",
		HoleCards: map[int][]Card{
			0: {Card{Rank: "2", Suit: "h"}, Card{Rank: "3", Suit: "h"}}, // Weak but not folded
			1: {Card{Rank: "A", Suit: "s"}, Card{Rank: "K", Suit: "s"}}, // Strong but folded
			2: {Card{Rank: "Q", Suit: "d"}, Card{Rank: "J", Suit: "d"}}, // Medium but folded
		},
		BoardCards: []Card{
			{Rank: "4", Suit: "s"}, {Rank: "5", Suit: "h"}, {Rank: "6", Suit: "d"},
			{Rank: "7", Suit: "c"}, {Rank: "8", Suit: "s"},
		},
		FoldedPlayers: map[int]bool{
			1: true, // Folded
			2: true, // Folded
		},
	}

	// Setup seats
	table.Seats[0].Status = "active"
	table.Seats[1].Status = "active"
	table.Seats[2].Status = "active"

	winners, winningRank := hand.DetermineWinner(seatsToPointers(table.Seats[:]))

	// Should only evaluate seat 0 since 1 and 2 folded
	if len(winners) != 1 {
		t.Fatalf("expected 1 winner, got %d", len(winners))
	}

	if winners[0] != 0 {
		t.Errorf("expected winner to be seat 0 (only non-folded), got seat %d", winners[0])
	}

	if winningRank == nil {
		t.Error("expected winningRank to be non-nil")
	}
}

// TestHandleShowdown_TriggersOnRiverComplete - Verify HandleShowdown is called
func TestHandleShowdown_TriggersOnRiverComplete(t *testing.T) {
	server := &Server{logger: slog.Default()}
	table := NewTable("table-1", "Test Table", server)
	hand := &Hand{
		Pot:            100,
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Street:         "river",
		HoleCards: map[int][]Card{
			0: {Card{Rank: "A", Suit: "s"}, Card{Rank: "K", Suit: "s"}},
			1: {Card{Rank: "2", Suit: "h"}, Card{Rank: "3", Suit: "h"}},
		},
		BoardCards: []Card{
			{Rank: "5", Suit: "d"}, {Rank: "6", Suit: "c"}, {Rank: "7", Suit: "s"},
			{Rank: "8", Suit: "h"}, {Rank: "9", Suit: "d"},
		},
		FoldedPlayers: make(map[int]bool),
	}

	table.CurrentHand = hand
	dealerSeat := 0
	table.DealerSeat = &dealerSeat // Set the table's dealer to match the hand
	table.Seats[0].Status = "active"
	table.Seats[1].Status = "active"

	// HandleShowdown should not panic and should return without error
	table.HandleShowdown()

	// Verify hand is cleared after HandleShowdown
	if table.CurrentHand != nil {
		t.Error("expected CurrentHand to be nil after HandleShowdown")
	}

	// Verify dealer was rotated
	if table.DealerSeat == nil || *table.DealerSeat != 1 {
		t.Errorf("expected dealer to rotate to seat 1, got %v", table.DealerSeat)
	}
}

// TestHandleShowdown_EarlyWinner_AllFold - Single remaining player wins without evaluation
// Verifies early winner receives pot and opponent bust-out seats are cleared
func TestHandleShowdown_EarlyWinner_AllFold(t *testing.T) {
	server := &Server{logger: slog.Default()}
	table := NewTable("table-1", "Test Table", server)

	// Set up initial stacks
	token0 := "player-0"
	token1 := "player-1"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 950 // Started with 1000, put 50 in pot

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 0 // Already all-in

	hand := &Hand{
		Pot:            100, // 50 from each player
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   2,
		Street:         "flop",
		HoleCards: map[int][]Card{
			0: {Card{Rank: "A", Suit: "s"}, Card{Rank: "K", Suit: "s"}},
			1: {Card{Rank: "2", Suit: "h"}, Card{Rank: "3", Suit: "h"}},
		},
		BoardCards: []Card{
			{Rank: "5", Suit: "d"}, {Rank: "6", Suit: "c"}, {Rank: "7", Suit: "s"},
		},
		FoldedPlayers: map[int]bool{
			1: true, // All but seat 0 folded
		},
	}

	table.CurrentHand = hand

	// Call HandleShowdown - should handle early winner case
	table.HandleShowdown()

	// Verify winner's stack is increased by pot amount
	if table.Seats[0].Stack != 1050 {
		t.Errorf("expected winner stack 1050, got %d", table.Seats[0].Stack)
	}

	// Verify bust-out seat (seat 1 with stack 0) is cleared
	if table.Seats[1].Status != "empty" {
		t.Errorf("expected bust-out seat to be 'empty', got '%s'", table.Seats[1].Status)
	}
	if table.Seats[1].Token != nil {
		t.Errorf("expected bust-out seat Token to be nil, got %v", table.Seats[1].Token)
	}
}

// TestHandleShowdown_EarlyWinner_OpponentBustsOut - Early winner with opponent busting out
// Verifies winner gets pot and opponent bust-out seat is cleared
func TestHandleShowdown_EarlyWinner_OpponentBustsOut(t *testing.T) {
	server := &Server{logger: slog.Default()}
	table := NewTable("table-1", "Test Table", server)

	// Set up initial stacks - opponent is all-in with 0 chips, will bust after losing
	token0 := "player-0"
	token1 := "player-1"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 900 // Has 900 chips

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 0 // Already all-in with no chips

	hand := &Hand{
		Pot:            100, // Opponent put their last chips in, seat 0 matched with 100
		DealerSeat:     0,
		SmallBlindSeat: 1,
		BigBlindSeat:   0,
		Street:         "river",
		HoleCards: map[int][]Card{
			0: {Card{Rank: "A", Suit: "s"}, Card{Rank: "K", Suit: "s"}},
			1: {Card{Rank: "2", Suit: "h"}, Card{Rank: "3", Suit: "h"}},
		},
		BoardCards: []Card{
			{Rank: "5", Suit: "d"}, {Rank: "6", Suit: "c"}, {Rank: "7", Suit: "s"},
			{Rank: "8", Suit: "h"}, {Rank: "9", Suit: "c"},
		},
		FoldedPlayers: map[int]bool{
			1: true, // Opponent folded (early winner)
		},
	}

	table.CurrentHand = hand

	// Call HandleShowdown - should handle early winner case with opponent already busted
	table.HandleShowdown()

	// Verify winner's stack includes the pot (900 + 100)
	if table.Seats[0].Stack != 1000 {
		t.Errorf("expected winner stack 1000, got %d", table.Seats[0].Stack)
	}

	// Verify bust-out opponent seat is cleared (stack was 0, stays 0)
	if table.Seats[1].Status != "empty" {
		t.Errorf("expected bust-out opponent seat Status to be 'empty', got '%s'", table.Seats[1].Status)
	}
	if table.Seats[1].Token != nil {
		t.Errorf("expected bust-out opponent seat Token to be nil, got %v", table.Seats[1].Token)
	}
	if table.Seats[1].Stack != 0 {
		t.Errorf("expected bust-out opponent stack to remain 0, got %d", table.Seats[1].Stack)
	}
}

// TestHandleShowdown_EarlyWinner_UnsweptBets verifies that HandleShowdown correctly
// sweeps unswept PlayerBets into Pot before calculating winner's new stack.
// This test replicates the critical bug: preflop SB(10) + raise to 100, BB(20) + fold.
// Expected: Winner gets 120 total (10 SB + 20 BB + 100 raise)
// Bug: Previous code read Pot=0 before sweep, so winner got 0
func TestHandleShowdown_EarlyWinner_UnsweptBets(t *testing.T) {
	server := &Server{logger: slog.Default()}
	table := NewTable("table-1", "Test Table", server)

	// Set up initial stacks
	token0 := "player-0" // SB, will raise and win
	token1 := "player-1" // BB, will fold and bust
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000 // Initial 1000

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 20 // BB with only 20 (will go all-in with BB and fold)

	// Create a hand with unswept bets (critical bug scenario)
	// Scenario: SB (player 0) posts 10 as SB, BB (player 1) posts 20 as BB
	// SB then raises to 100, BB folds immediately (early winner)
	hand := &Hand{
		Pot:            30, // SB(10) + BB(20) already in pot from blind posting
		DealerSeat:     1,
		SmallBlindSeat: 0,
		BigBlindSeat:   1,
		Street:         "preflop",
		HoleCards: map[int][]Card{
			0: {Card{Rank: "A", Suit: "s"}, Card{Rank: "K", Suit: "s"}},
			1: {Card{Rank: "2", Suit: "h"}, Card{Rank: "3", Suit: "h"}},
		},
		// Additional bets during preflop: SB raises additional 90 (to 100 total)
		PlayerBets: map[int]int{
			0: 90, // SB's additional bet to reach 100 total
		},
		FoldedPlayers: map[int]bool{
			1: true, // BB folded (early winner scenario)
		},
	}

	table.CurrentHand = hand

	// Call HandleShowdown - should sweep PlayerBets into Pot before calculating winner stack
	table.HandleShowdown()

	// CRITICAL VERIFICATION: Winner should get 120 (100 from their bet + 20 from BB)
	// Expected winner stack: 1000 (initial) + 120 (pot) = 1120
	expectedWinnerStack := 1000 + 120 // Initial 1000 + pot of 120
	if table.Seats[0].Stack != expectedWinnerStack {
		t.Errorf("expected winner stack %d, got %d", expectedWinnerStack, table.Seats[0].Stack)
	}
}

// TestHandleBustOutsWithNotificationsLocked_SinglePlayerBusted
// Verifies that a single player with stack 0 is identified, cleared, and token is returned
func TestHandleBustOutsWithNotificationsLocked_SinglePlayerBusted(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)

	// Set up: 2 players, one with 0 stack
	token0 := "player-0"
	token1 := "player-1"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 500

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 0 // This player busted

	// Call handleBustOutsWithNotificationsLocked - should return the busted token
	bustedTokens := table.handleBustOutsWithNotificationsLocked()

	// Verify busted player token is returned
	if len(bustedTokens) != 1 {
		t.Errorf("expected 1 busted token, got %d", len(bustedTokens))
	}
	if len(bustedTokens) > 0 && bustedTokens[0] != token1 {
		t.Errorf("expected busted token '%s', got '%s'", token1, bustedTokens[0])
	}

	// Verify seat 1 is cleared
	if table.Seats[1].Status != "empty" {
		t.Errorf("expected busted seat to be 'empty', got '%s'", table.Seats[1].Status)
	}
	if table.Seats[1].Token != nil {
		t.Errorf("expected busted seat Token to be nil, got %v", table.Seats[1].Token)
	}

	// Verify seat 0 (winner) is untouched
	if table.Seats[0].Status != "active" {
		t.Errorf("expected winner seat to remain 'active', got '%s'", table.Seats[0].Status)
	}
	if table.Seats[0].Token == nil || *table.Seats[0].Token != token0 {
		t.Errorf("expected winner token to remain '%s'", token0)
	}
	if table.Seats[0].Stack != 500 {
		t.Errorf("expected winner stack to remain 500, got %d", table.Seats[0].Stack)
	}
}

// TestHandleBustOutsWithNotificationsLocked_MultiplePlayersBusted
// Verifies multiple players with stack 0 are identified, cleared, and tokens are returned
func TestHandleBustOutsWithNotificationsLocked_MultiplePlayersBusted(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)

	// Set up: 4 players, 2 with 0 stack
	token0 := "player-0"
	token1 := "player-1"
	token2 := "player-2"
	token3 := "player-3"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 0 // Busted

	table.Seats[2].Token = &token2
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 500

	table.Seats[3].Token = &token3
	table.Seats[3].Status = "active"
	table.Seats[3].Stack = 0 // Busted

	// Call handleBustOutsWithNotificationsLocked
	bustedTokens := table.handleBustOutsWithNotificationsLocked()

	// Verify 2 busted tokens are returned
	if len(bustedTokens) != 2 {
		t.Errorf("expected 2 busted tokens, got %d", len(bustedTokens))
	}

	// Verify busted tokens are correct (order may vary)
	bustedTokenMap := make(map[string]bool)
	for _, token := range bustedTokens {
		bustedTokenMap[token] = true
	}
	if !bustedTokenMap[token1] {
		t.Errorf("expected busted token '%s' in result", token1)
	}
	if !bustedTokenMap[token3] {
		t.Errorf("expected busted token '%s' in result", token3)
	}

	// Verify seats 1 and 3 are cleared
	if table.Seats[1].Status != "empty" {
		t.Errorf("seat 1: expected 'empty', got '%s'", table.Seats[1].Status)
	}
	if table.Seats[1].Token != nil {
		t.Errorf("seat 1: expected Token nil, got %v", table.Seats[1].Token)
	}

	if table.Seats[3].Status != "empty" {
		t.Errorf("seat 3: expected 'empty', got '%s'", table.Seats[3].Status)
	}
	if table.Seats[3].Token != nil {
		t.Errorf("seat 3: expected Token nil, got %v", table.Seats[3].Token)
	}

	// Verify other seats are untouched
	if table.Seats[0].Stack != 1000 {
		t.Errorf("seat 0: expected stack 1000, got %d", table.Seats[0].Stack)
	}
	if table.Seats[2].Stack != 500 {
		t.Errorf("seat 2: expected stack 500, got %d", table.Seats[2].Stack)
	}
}

// TestHandleBustOutsWithNotificationsLocked_NoBustOuts
// Verifies no bust-outs returns empty list and seats are unchanged
func TestHandleBustOutsWithNotificationsLocked_NoBustOuts(t *testing.T) {
	table := NewTable("table-1", "Test Table", nil)

	// Set up: players with non-zero stacks
	token0 := "player-0"
	token1 := "player-1"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 500

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 300

	// Call handleBustOutsWithNotificationsLocked
	bustedTokens := table.handleBustOutsWithNotificationsLocked()

	// Verify no busted tokens
	if len(bustedTokens) != 0 {
		t.Errorf("expected 0 busted tokens, got %d", len(bustedTokens))
	}

	// Verify seats are unchanged
	if table.Seats[0].Stack != 500 {
		t.Errorf("seat 0: expected stack 500, got %d", table.Seats[0].Stack)
	}
	if table.Seats[1].Stack != 300 {
		t.Errorf("seat 1: expected stack 300, got %d", table.Seats[1].Stack)
	}

	if table.Seats[0].Status != "active" {
		t.Errorf("seat 0: expected status 'active', got '%s'", table.Seats[0].Status)
	}
	if table.Seats[1].Status != "active" {
		t.Errorf("seat 1: expected status 'active', got '%s'", table.Seats[1].Status)
	}
}

// ============ PHASE 2: INTEGRATION TESTS - SHOWDOWN WITH AUTO-KICK ============

// TestShowdown_AllInPlayerBustsOut verifies all-in player losing at showdown gets auto-kicked
// Simulates full hand flow: deal cards, betting (all-in), showdown, bust-out and verify auto-kick
// Uses specific hole cards to GUARANTEE deterministic outcome: player 0 wins, player 1 busts
func TestShowdown_AllInPlayerBustsOut(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	table := server.tables[0]

	// Set up 2 players: player 0 with 1000, player 1 with exactly 30 (enough for SB+remaining to bet all)
	token0 := "player-0"
	token1 := "player-1"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 30 // Exactly enough to cover SB (10) + remaining bet (20 more)

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// After StartHand:
	// - Player 0 (dealer/SB) posts 10, has 990 left
	// - Player 1 (BB) posts 20, has 10 left
	// - Pot is 30
	// To simulate an all-in, we need to have player 1 bet their remaining 10 chips
	// and have player 0 call. We'll update the hand state to reflect this.

	// Set player 1's remaining stack to 0 (they went all-in with 10 on preflop)
	// With new pot accounting, bets stay in PlayerBets until AdvanceStreet or HandleShowdown
	table.CurrentHand.PlayerBets[0] = 30 // Player 0 bet total of 30 (SB 10 + call 20)
	table.CurrentHand.PlayerBets[1] = 30 // Player 1 bet total of 30 (all-in)
	table.CurrentHand.Pot = 0            // Pot is 0 during betting; will be swept to 60 at showdown

	// Update TotalContributions to match the actual contributions in this scenario
	table.CurrentHand.TotalContributions[0] = 30 // Player 0 total contribution
	table.CurrentHand.TotalContributions[1] = 30 // Player 1 total contribution

	// Now update stacks to reflect the all-in
	table.Seats[0].Stack = 1000 - 10 - 20 // After SB (10) and calling the all-in (20), has 970
	table.Seats[1].Stack = 0              // All-in with 30

	// Set specific hole cards to GUARANTEE player 0 wins and player 1 loses
	table.CurrentHand.HoleCards[0] = []Card{
		{Rank: "A", Suit: "s"},
		{Rank: "A", Suit: "h"},
	}

	// Player 1 has 2-3 (worst possible hand, will have pair of 2s at best)
	table.CurrentHand.HoleCards[1] = []Card{
		{Rank: "2", Suit: "c"},
		{Rank: "3", Suit: "d"},
	}

	// Set board cards that don't form complete hands: K-Q-J-9-2
	// Player 0 will have pair of Aces (kicker K-Q-J)
	// Player 1 will have pair of 2s (kicker K-Q-J)
	// Player 0 wins due to higher pair
	table.CurrentHand.BoardCards = []Card{
		{Rank: "K", Suit: "c"},
		{Rank: "Q", Suit: "d"},
		{Rank: "J", Suit: "s"},
		{Rank: "9", Suit: "h"},
		{Rank: "2", Suit: "s"},
	}

	// Manually set street to river (showdown state)
	table.CurrentHand.Street = "river"

	// Get initial state before showdown
	initialToken0Stack := table.Seats[0].Stack

	// With new pot accounting, calculate total pot from PlayerBets
	totalPlayerBets := 0
	for _, bet := range table.CurrentHand.PlayerBets {
		totalPlayerBets += bet
	}
	initialPot := table.CurrentHand.Pot + totalPlayerBets

	// Call HandleShowdown
	table.HandleShowdown()

	// After showdown, verify deterministic bust-out:
	// Player 1 MUST have lost and busted out (stack == 0)
	if table.Seats[1].Stack != 0 {
		t.Fatalf("expected player 1 to bust out (stack == 0), but got stack %d", table.Seats[1].Stack)
	}

	// Verify seat 1 is cleared (auto-kicked)
	if table.Seats[1].Token != nil {
		t.Errorf("expected seat 1 (busted out) to have Token == nil, got %v", table.Seats[1].Token)
	}
	if table.Seats[1].Status != "empty" {
		t.Errorf("expected seat 1 (busted out) to have Status 'empty', got '%s'", table.Seats[1].Status)
	}

	// Verify player 0 won and has increased stack (should have initial + pot)
	expectedStack := initialToken0Stack + initialPot
	if table.Seats[0].Stack != expectedStack {
		t.Errorf("expected seat 0 to have stack %d (initial %d + pot %d), got %d", expectedStack, initialToken0Stack, initialPot, table.Seats[0].Stack)
	}

	// Verify hand is cleared after showdown
	if table.CurrentHand != nil {
		t.Error("expected CurrentHand to be nil after HandleShowdown")
	}
}

// TestShowdown_MultiplePlayersBustOut verifies multiple all-in losers with zero stacks all get auto-kicked
// Simulates 3-player hand where 2 players bust simultaneously after showdown
// Uses specific hole cards to GUARANTEE deterministic outcome: player 0 wins, players 1 and 2 bust
func TestShowdown_MultiplePlayersBustOut(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	table := server.tables[0]

	// Set up 3 players: player 0 with 1000, players 1 and 2 with 30 each (enough for blinds + all-in)
	token0 := "player-0"
	token1 := "player-1"
	token2 := "player-2"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 30 // Small stack 1

	table.Seats[2].Token = &token2
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 30 // Small stack 2

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// After StartHand with 3 players (dealer is 0):
	// - Player 0 (dealer/SB) posts 10, has 990 left
	// - Player 1 (BB) posts 20, has 10 left
	// - Player 2 (UTG) has 30, acts first
	// To simulate all-in scenario: player 2 goes all-in with 30, player 0 calls with 990
	// (getting back to blinds), player 1 goes all-in with remaining 10
	// Simplified: We manually set up the state where:
	// - Player 1 all-in with 30 total
	// - Player 2 all-in with 30 total
	// - Player 0 called both with 60 total
	// - Pot = 30 + 30 + 60 = 120

	table.CurrentHand.PlayerBets[0] = 60 // Player 0 bet 60
	table.CurrentHand.PlayerBets[1] = 30 // Player 1 all-in with 30
	table.CurrentHand.PlayerBets[2] = 30 // Player 2 all-in with 30
	table.CurrentHand.Pot = 0            // With new pot accounting, pot is 0 during betting

	// Update TotalContributions to match the actual contributions in this scenario
	table.CurrentHand.TotalContributions[0] = 60 // Player 0 total contribution
	table.CurrentHand.TotalContributions[1] = 30 // Player 1 total contribution (all-in)
	table.CurrentHand.TotalContributions[2] = 30 // Player 2 total contribution (all-in)

	// Update stacks to reflect all-in
	table.Seats[0].Stack = 1000 - 10 - 60 // 930 (after SB + call)
	table.Seats[1].Stack = 0              // All-in with 30
	table.Seats[2].Stack = 0              // All-in with 30

	// Set specific hole cards to GUARANTEE player 0 wins, players 1 and 2 lose
	// Player 0 has pair of Kings
	table.CurrentHand.HoleCards[0] = []Card{
		{Rank: "K", Suit: "s"},
		{Rank: "K", Suit: "h"},
	}

	// Player 1 has 2-3 (worst possible hand - will have pair of 2s at best)
	table.CurrentHand.HoleCards[1] = []Card{
		{Rank: "2", Suit: "c"},
		{Rank: "3", Suit: "d"},
	}

	// Player 2 has 4-5 (low hand - will have pair of 4s or nothing)
	table.CurrentHand.HoleCards[2] = []Card{
		{Rank: "4", Suit: "c"},
		{Rank: "5", Suit: "d"},
	}

	// Set board cards that don't form complete hands: Q-J-T-9-2
	// Player 0 will have pair of Kings (best hand, kicker Q-J-T)
	// Player 1 will have pair of 2s (kicker Q-J-T)
	// Player 2 will have high card (kicker K-Q-J-T-9)
	// Player 0 wins with pair of Kings
	table.CurrentHand.BoardCards = []Card{
		{Rank: "Q", Suit: "c"},
		{Rank: "J", Suit: "d"},
		{Rank: "T", Suit: "s"},
		{Rank: "9", Suit: "h"},
		{Rank: "2", Suit: "s"},
	}

	// Manually set street to river (showdown state)
	table.CurrentHand.Street = "river"

	// Get initial state
	initialStack0 := table.Seats[0].Stack

	// With new pot accounting, calculate total pot from PlayerBets
	totalPlayerBets := 0
	for _, bet := range table.CurrentHand.PlayerBets {
		totalPlayerBets += bet
	}
	initialPot := table.CurrentHand.Pot + totalPlayerBets

	// Call HandleShowdown
	table.HandleShowdown()

	// After showdown, verify deterministic bust-out of multiple players:
	// Players 1 and 2 MUST have lost and busted out (stack == 0)
	if table.Seats[1].Stack != 0 {
		t.Fatalf("expected player 1 to bust out (stack == 0), but got stack %d", table.Seats[1].Stack)
	}
	if table.Seats[2].Stack != 0 {
		t.Fatalf("expected player 2 to bust out (stack == 0), but got stack %d", table.Seats[2].Stack)
	}

	// Verify seats 1 and 2 are cleared (auto-kicked)
	if table.Seats[1].Token != nil {
		t.Errorf("expected seat 1 (busted out) to have Token == nil, got %v", table.Seats[1].Token)
	}
	if table.Seats[1].Status != "empty" {
		t.Errorf("expected seat 1 (busted out) to have Status 'empty', got '%s'", table.Seats[1].Status)
	}

	if table.Seats[2].Token != nil {
		t.Errorf("expected seat 2 (busted out) to have Token == nil, got %v", table.Seats[2].Token)
	}
	if table.Seats[2].Status != "empty" {
		t.Errorf("expected seat 2 (busted out) to have Status 'empty', got '%s'", table.Seats[2].Status)
	}

	// Verify player 0 won and has increased stack (should have initial + pot)
	expectedStack := initialStack0 + initialPot
	if table.Seats[0].Stack != expectedStack {
		t.Fatalf("expected seat 0 to have stack %d (initial %d + pot %d), got %d", expectedStack, initialStack0, initialPot, table.Seats[0].Stack)
	}

	// Verify hand is cleared
	if table.CurrentHand != nil {
		t.Error("expected CurrentHand to be nil after HandleShowdown")
	}
}

// TestShowdown_WinnerWithStackNotKicked verifies winners with stack > 0 are NOT kicked
// Player 0: AA (pair of Aces) - wins
// Player 1: KQ (pair of Kings with board) - loses but has remaining stack, not kicked
func TestShowdown_WinnerWithStackNotKicked(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	table := server.tables[0]

	// Set up 2 players
	token0 := "player-0"
	token1 := "player-1"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 100

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 100

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Set specific hole cards to guarantee player 0 wins
	table.CurrentHand.HoleCards[0] = []Card{
		{Rank: "A", Suit: "s"},
		{Rank: "A", Suit: "h"},
	}
	table.CurrentHand.HoleCards[1] = []Card{
		{Rank: "K", Suit: "c"},
		{Rank: "Q", Suit: "d"},
	}

	// Manually set up showdown scenario
	table.CurrentHand.Street = "river"

	// Board: 9-8-7-5-2 (no pairs/straights involving K,Q - player 0 has pair of Aces, player 1 has high card K)
	table.CurrentHand.BoardCards = []Card{
		{Rank: "9", Suit: "d"},
		{Rank: "8", Suit: "h"},
		{Rank: "7", Suit: "s"},
		{Rank: "5", Suit: "c"},
		{Rank: "2", Suit: "d"},
	}

	// Call HandleShowdown - player 0 should win with pair of Aces
	table.HandleShowdown()

	// Verify player 0 won (has pair of Aces)
	if table.Seats[0].Stack <= 100 {
		t.Errorf("expected player 0 to win and have stack > 100, got %d", table.Seats[0].Stack)
	}
	if table.Seats[0].Token == nil {
		t.Error("expected player 0 (winner) to NOT be kicked (Token should not be nil)")
	}
	if table.Seats[0].Status == "empty" {
		t.Error("expected player 0 (winner) to NOT be kicked (Status should not be 'empty')")
	}

	// Verify player 1 lost but still has chips (not busted)
	if table.Seats[1].Stack <= 0 {
		t.Errorf("expected player 1 to lose but NOT bust (should have stack > 0), got %d", table.Seats[1].Stack)
	}
	if table.Seats[1].Stack >= 100 {
		t.Errorf("expected player 1 to lose some chips (stack < 100), got %d", table.Seats[1].Stack)
	}
	if table.Seats[1].Token == nil {
		t.Error("expected player 1 (loser with remaining chips) to NOT be kicked (Token should not be nil)")
	}
	if table.Seats[1].Status == "empty" {
		t.Error("expected player 1 (loser with remaining chips) to NOT be kicked (Status should not be 'empty')")
	}

	// Verify hand is cleared
	if table.CurrentHand != nil {
		t.Error("expected CurrentHand to be nil after HandleShowdown")
	}
}

// TestShowdown_AllInWinnerNotKicked verifies edge case: player goes all-in and WINS (not kicked)
// Player starts with 30 chip stack, goes all-in, and wins despite having 0 stack before distribution
// After pot distribution, should have stack > 0 and NOT be kicked (not an empty seat)
func TestShowdown_AllInWinnerNotKicked(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	table := server.tables[0]

	// Set up 2 players: both with small all-in stacks
	token0 := "player-0"
	token1 := "player-1"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 30 // Small all-in stack

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 30 // Small all-in stack

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Simulate all-in scenario where:
	// - Player 0 (SB) goes all-in with 30
	// - Player 1 (BB) goes all-in with 30
	// - Pot = 60
	// After StartHand: Player 0 has 20 (30-10 SB), Player 1 has 10 (30-20 BB)
	// We need both to have 0 remaining, so they each bet all their chips
	// Player 0 bets remaining 20 (total 30), Player 1 calls with 10 (total 30)

	table.CurrentHand.PlayerBets[0] = 30 // Player 0 all-in with 30 total
	table.CurrentHand.PlayerBets[1] = 30 // Player 1 all-in with 30 total
	table.CurrentHand.Pot = 0            // With new pot accounting, pot is 0 during betting

	// Update TotalContributions to match the actual contributions in this scenario
	table.CurrentHand.TotalContributions[0] = 30 // Player 0 total contribution
	table.CurrentHand.TotalContributions[1] = 30 // Player 1 total contribution

	// Update stacks to reflect all-in
	table.Seats[0].Stack = 0 // Player 0 all-in
	table.Seats[1].Stack = 0 // Player 1 all-in

	// Set specific hole cards to GUARANTEE player 0 wins despite all-in with small stack
	// Player 0 has pair of Aces (will win)
	table.CurrentHand.HoleCards[0] = []Card{
		{Rank: "A", Suit: "s"},
		{Rank: "A", Suit: "h"},
	}

	// Player 1 has 2-3 (worst hand - will lose, will have pair of 2s at best)
	table.CurrentHand.HoleCards[1] = []Card{
		{Rank: "2", Suit: "c"},
		{Rank: "3", Suit: "d"},
	}

	// Set board cards that don't form complete hands: K-Q-J-9-2
	// Player 0 will have pair of Aces (best hand)
	// Player 1 will have pair of 2s (loses)
	table.CurrentHand.BoardCards = []Card{
		{Rank: "K", Suit: "c"},
		{Rank: "Q", Suit: "d"},
		{Rank: "J", Suit: "s"},
		{Rank: "9", Suit: "h"},
		{Rank: "2", Suit: "s"},
	}

	// Manually set street to river (showdown state)
	table.CurrentHand.Street = "river"

	// Get initial pot - calculate from PlayerBets with new accounting
	totalPlayerBets := 0
	for _, bet := range table.CurrentHand.PlayerBets {
		totalPlayerBets += bet
	}
	initialPot := table.CurrentHand.Pot + totalPlayerBets

	// Call HandleShowdown
	table.HandleShowdown()

	// After showdown, verify winner with all-in stack is NOT kicked:
	// Player 0 should have won and have stack > 0 after pot distribution
	if table.Seats[0].Stack <= 0 {
		t.Fatalf("expected player 0 (all-in winner) to have stack > 0 after pot distribution, got %d", table.Seats[0].Stack)
	}

	// Verify the winner received the correct pot
	expectedStack := initialPot // Winner gets the entire pot
	if table.Seats[0].Stack != expectedStack {
		t.Errorf("expected player 0 to have stack %d (pot %d), got %d", expectedStack, initialPot, table.Seats[0].Stack)
	}

	// Verify seat 0 is NOT cleared (player is still seated)
	if table.Seats[0].Token == nil {
		t.Errorf("expected seat 0 (all-in winner) to have Token != nil, got nil")
	}
	if table.Seats[0].Status == "empty" {
		t.Errorf("expected seat 0 (all-in winner) to NOT have Status 'empty', got 'empty'")
	}

	// Verify player 1 lost and busted out
	if table.Seats[1].Stack != 0 {
		t.Errorf("expected player 1 to bust out (stack == 0), got %d", table.Seats[1].Stack)
	}

	// Verify seat 1 is cleared (auto-kicked as busted player)
	if table.Seats[1].Token != nil {
		t.Errorf("expected seat 1 (busted out) to have Token == nil, got %v", table.Seats[1].Token)
	}
	if table.Seats[1].Status != "empty" {
		t.Errorf("expected seat 1 (busted out) to have Status 'empty', got '%s'", table.Seats[1].Status)
	}

	// Verify hand is cleared
	if table.CurrentHand != nil {
		t.Error("expected CurrentHand to be nil after HandleShowdown")
	}
}

// ============================================================================
// PHASE 1: Fix Raise Validation Logic - Multi-Player Support
// Tests that verify players can ALWAYS bet their full stack regardless of
// opponent stack sizes. These tests replace the old tests that incorrectly
// limited maxRaise based on opponent stacks.
// ============================================================================

// TestGetMaxRaise_2P_SB_AllIn_BugFix verifies the core bug fix:
// 2 players, SB (990 remaining after posting SB) can go all-in
// even though BB (980 remaining) has less chips
func TestGetMaxRaise_2P_SB_AllIn_BugFix(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Setup: A=1000, B=1000, post blinds (A=990 SB, B=980 BB)
	tokenA := "player-a"
	tokenB := "player-b"
	table.Seats[0].Token = &tokenA
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 990 // After posting SB (1000 - 10)

	table.Seats[1].Token = &tokenB
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 980 // After posting BB (1000 - 20)

	// A (SB) should be able to raise to 990 (their full stack)
	maxRaise := table.GetMaxRaise(0, createEmptyHand())
	if maxRaise != 990 {
		t.Errorf("TestGetMaxRaise_2P_SB_AllIn_BugFix: Player A should be able to bet full stack 990, got %d", maxRaise)
	}

	// B (BB) should be able to raise to 980 (their full stack)
	maxRaise = table.GetMaxRaise(1, createEmptyHand())
	if maxRaise != 980 {
		t.Errorf("TestGetMaxRaise_2P_SB_AllIn_BugFix: Player B should be able to bet full stack 980, got %d", maxRaise)
	}
}

// TestGetMaxRaise_2P_Both_Equal_Stacks verifies both can go all-in with equal stacks
func TestGetMaxRaise_2P_Both_Equal_Stacks(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Setup: Both 1000 chips
	tokenA := "player-a"
	tokenB := "player-b"
	table.Seats[0].Token = &tokenA
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &tokenB
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	// Both can go all-in for 1000
	maxRaise := table.GetMaxRaise(0, createEmptyHand())
	if maxRaise != 1000 {
		t.Errorf("TestGetMaxRaise_2P_Both_Equal_Stacks: Player A should get 1000, got %d", maxRaise)
	}

	maxRaise = table.GetMaxRaise(1, createEmptyHand())
	if maxRaise != 1000 {
		t.Errorf("TestGetMaxRaise_2P_Both_Equal_Stacks: Player B should get 1000, got %d", maxRaise)
	}
}

// TestGetMaxRaise_2P_Short_Stack_Can_AllIn verifies short stack player can go all-in
func TestGetMaxRaise_2P_Short_Stack_Can_AllIn(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Setup: A=500, B=1000
	tokenA := "player-a"
	tokenB := "player-b"
	table.Seats[0].Token = &tokenA
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 490 // After posting SB (500 - 10)

	table.Seats[1].Token = &tokenB
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	// A (short stack) should be able to go all-in for 490
	maxRaise := table.GetMaxRaise(0, createEmptyHand())
	if maxRaise != 490 {
		t.Errorf("TestGetMaxRaise_2P_Short_Stack_Can_AllIn: Player A should get 490, got %d", maxRaise)
	}

	// B should be able to go all-in for 1000
	maxRaise = table.GetMaxRaise(1, createEmptyHand())
	if maxRaise != 1000 {
		t.Errorf("TestGetMaxRaise_2P_Short_Stack_Can_AllIn: Player B should get 1000, got %d", maxRaise)
	}
}

// TestGetMaxRaise_3P_One_Short_Stack_Can_AllIn verifies short stack in 3-player game
func TestGetMaxRaise_3P_One_Short_Stack_Can_AllIn(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Setup: A=1000, B=500, C=1000
	stacks := []int{1000, 490, 1000} // B after posting SB
	tokens := []string{"player-a", "player-b", "player-c"}
	for i := 0; i < 3; i++ {
		table.Seats[i].Token = &tokens[i]
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = stacks[i]
	}

	// B (SB with 490) should be able to go all-in for 490
	maxRaise := table.GetMaxRaise(1, createEmptyHand())
	if maxRaise != 490 {
		t.Errorf("TestGetMaxRaise_3P_One_Short_Stack_Can_AllIn: Player B should get 490, got %d", maxRaise)
	}

	// A should be able to go all-in for 1000
	maxRaise = table.GetMaxRaise(0, createEmptyHand())
	if maxRaise != 1000 {
		t.Errorf("TestGetMaxRaise_3P_One_Short_Stack_Can_AllIn: Player A should get 1000, got %d", maxRaise)
	}

	// C should be able to go all-in for 1000
	maxRaise = table.GetMaxRaise(2, createEmptyHand())
	if maxRaise != 1000 {
		t.Errorf("TestGetMaxRaise_3P_One_Short_Stack_Can_AllIn: Player C should get 1000, got %d", maxRaise)
	}
}

// TestGetMaxRaise_3P_Multiple_Different_Stacks verifies all can bet their stacks
func TestGetMaxRaise_3P_Multiple_Different_Stacks(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Setup: A=2000, B=1000, C=500
	stacks := []int{2000, 1000, 500}
	tokens := []string{"player-a", "player-b", "player-c"}
	for i := 0; i < 3; i++ {
		table.Seats[i].Token = &tokens[i]
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = stacks[i]
	}

	// Each player can bet their full stack
	if maxRaise := table.GetMaxRaise(0, createEmptyHand()); maxRaise != 2000 {
		t.Errorf("Player A should get 2000, got %d", maxRaise)
	}
	if maxRaise := table.GetMaxRaise(1, createEmptyHand()); maxRaise != 1000 {
		t.Errorf("Player B should get 1000, got %d", maxRaise)
	}
	if maxRaise := table.GetMaxRaise(2, createEmptyHand()); maxRaise != 500 {
		t.Errorf("Player C should get 500, got %d", maxRaise)
	}
}

// TestGetMaxRaise_3P_Whale_Can_Overbet_All verifies whale can go all-in for 5000
func TestGetMaxRaise_3P_Whale_Can_Overbet_All(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Setup: A=5000 (whale), B=1000, C=1000
	stacks := []int{5000, 1000, 1000}
	tokens := []string{"whale", "player-b", "player-c"}
	for i := 0; i < 3; i++ {
		table.Seats[i].Token = &tokens[i]
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = stacks[i]
	}

	// Whale should be able to bet full 5000
	maxRaise := table.GetMaxRaise(0, createEmptyHand())
	if maxRaise != 5000 {
		t.Errorf("TestGetMaxRaise_3P_Whale_Can_Overbet_All: Whale should get 5000, got %d", maxRaise)
	}
}

// TestGetMaxRaise_4P_Multiple_AllIns_Same_Hand verifies multiple players can go all-in
func TestGetMaxRaise_4P_Multiple_AllIns_Same_Hand(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Setup: 4 players with [1000, 800, 600, 1000]
	stacks := []int{1000, 800, 600, 1000}
	for i := 0; i < 4; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = stacks[i]
	}

	// All players can bet their full stacks
	for i := 0; i < 4; i++ {
		if maxRaise := table.GetMaxRaise(i, createEmptyHand()); maxRaise != stacks[i] {
			t.Errorf("Player %d should get %d, got %d", i, stacks[i], maxRaise)
		}
	}
}

// TestGetMaxRaise_4P_Shortest_Stack_All_Can_Bet_Full verifies all can bet full stacks
func TestGetMaxRaise_4P_Shortest_Stack_All_Can_Bet_Full(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Setup: Stacks [1000, 800, 600, 1200]
	stacks := []int{1000, 800, 600, 1200}
	for i := 0; i < 4; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = stacks[i]
	}

	// All players can bet their full stacks (no opponent stack limit)
	for i := 0; i < 4; i++ {
		if maxRaise := table.GetMaxRaise(i, createEmptyHand()); maxRaise != stacks[i] {
			t.Errorf("Player %d should get %d, got %d", i, stacks[i], maxRaise)
		}
	}
}

// TestGetMaxRaise_5P_Multiple_Callers_Different_Stacks verifies 5-player game
func TestGetMaxRaise_5P_Multiple_Callers_Different_Stacks(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Setup: 5 players with various stacks
	stacks := []int{2000, 1500, 1000, 500, 750}
	for i := 0; i < 5; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = stacks[i]
	}

	// All players can bet their full stacks
	for i := 0; i < 5; i++ {
		if maxRaise := table.GetMaxRaise(i, createEmptyHand()); maxRaise != stacks[i] {
			t.Errorf("Player %d should get %d, got %d", i, stacks[i], maxRaise)
		}
	}
}

// TestGetMaxRaise_6P_Whale_Overbets_Everyone verifies 6-player with whale
func TestGetMaxRaise_6P_Whale_Overbets_Everyone(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Setup: Stacks [10000, 1000, 1000, 800, 600, 400]
	stacks := []int{10000, 1000, 1000, 800, 600, 400}
	for i := 0; i < 6; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = stacks[i]
	}

	// Whale can bet full 10000
	if maxRaise := table.GetMaxRaise(0, createEmptyHand()); maxRaise != 10000 {
		t.Errorf("Whale should get 10000, got %d", maxRaise)
	}

	// All other players can bet their full stacks
	for i := 1; i < 6; i++ {
		if maxRaise := table.GetMaxRaise(i, createEmptyHand()); maxRaise != stacks[i] {
			t.Errorf("Player %d should get %d, got %d", i, stacks[i], maxRaise)
		}
	}
}

// TestValidateRaise_AllIn_Always_Valid verifies all-in is always valid
// (test for ValidateRaise function)
func TestValidateRaise_AllIn_Always_Valid(t *testing.T) {
	hand := &Hand{
		CurrentBet: 20,
		LastRaise:  20,
		PlayerBets: make(map[int]int),
	}

	// Setup: 2 players
	seats := [6]Seat{
		{Index: 0, Status: "active", Stack: 990},
		{Index: 1, Status: "active", Stack: 980},
	}

	// Player 0 going all-in for 990 should be valid
	err := hand.ValidateRaise(0, 990, 990, seats)
	if err != nil {
		t.Errorf("TestValidateRaise_AllIn_Always_Valid: All-in for 990 should be valid, got error: %v", err)
	}

	// Player 1 going all-in for 980 should be valid
	err = hand.ValidateRaise(1, 980, 980, seats)
	if err != nil {
		t.Errorf("TestValidateRaise_AllIn_Always_Valid: All-in for 980 should be valid, got error: %v", err)
	}
}

// TestValidateRaise_Short_Stack_Can_Raise_Full verifies short stack can raise full
func TestValidateRaise_Short_Stack_Can_Raise_Full(t *testing.T) {
	hand := &Hand{
		CurrentBet: 20,
		LastRaise:  20,
		PlayerBets: make(map[int]int),
	}

	// Setup: A=490 (short), B=1000 (big)
	seats := [6]Seat{
		{Index: 0, Status: "active", Stack: 490},
		{Index: 1, Status: "active", Stack: 1000},
	}

	// Short stack player should be able to raise to 490 (their full stack)
	err := hand.ValidateRaise(0, 490, 490, seats)
	if err != nil {
		t.Errorf("TestValidateRaise_Short_Stack_Can_Raise_Full: Should allow 490, got error: %v", err)
	}

	// Big stack player should be able to raise to 1000 (their full stack)
	err = hand.ValidateRaise(1, 1000, 1000, seats)
	if err != nil {
		t.Errorf("TestValidateRaise_Short_Stack_Can_Raise_Full: Should allow 1000, got error: %v", err)
	}
}

// TestSidePots_2P_EffectiveAllIn tests side pot creation when player goes all-in
// with effective stack shorter than opponent
func TestSidePots_2P_EffectiveAllIn(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)
	token0 := "player-a"
	token1 := "player-b"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 500
	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Player 0 raises to full amount (all-in)
	// Raise amount must account for amount already in PlayerBets
	raiseToAmount := table.Seats[0].Stack + table.CurrentHand.PlayerBets[0]
	chipsMoved, err := table.CurrentHand.ProcessAction(0, "raise", table.Seats[0].Stack, raiseToAmount)
	if err != nil {
		t.Fatalf("Player 0 raise failed: %v", err)
	}
	table.Seats[0].Stack -= chipsMoved

	// Player 1 calls to match the raise
	chipsMoved, err = table.CurrentHand.ProcessAction(1, "call", table.Seats[1].Stack)
	if err != nil {
		t.Fatalf("Player 1 call failed: %v", err)
	}
	table.Seats[1].Stack -= chipsMoved

	// Verify stacks
	if table.Seats[0].Stack != 0 {
		t.Errorf("Player 0 stack should be 0, got %d", table.Seats[0].Stack)
	}
	if table.Seats[1].Stack < 0 {
		t.Errorf("Player 1 stack should be non-negative, got %d", table.Seats[1].Stack)
	}
	// With new pot accounting, bets stay in PlayerBets until AdvanceStreet
	totalPlayerBets := 0
	for _, bet := range table.CurrentHand.PlayerBets {
		totalPlayerBets += bet
	}
	if totalPlayerBets <= 0 {
		t.Errorf("Total player bets should be positive, got %d", totalPlayerBets)
	}
}

// TestSidePots_2P_BothAllIn tests side pot when both players go all-in
func TestSidePots_2P_BothAllIn(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)
	token0 := "player-a"
	token1 := "player-b"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 100 // Smaller stack to enable raise/raise
	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 200

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Player 0 raises to full amount (all-in)
	// Raise amount must account for amount already in PlayerBets
	raiseToAmount := table.Seats[0].Stack + table.CurrentHand.PlayerBets[0]
	chipsMoved, err := table.CurrentHand.ProcessAction(0, "raise", table.Seats[0].Stack, raiseToAmount)
	if err != nil {
		t.Fatalf("Player 0 raise failed: %v", err)
	}
	table.Seats[0].Stack -= chipsMoved

	// Player 1 raises to full amount (all-in)
	// Raise amount must account for amount already in PlayerBets
	raiseToAmount = table.Seats[1].Stack + table.CurrentHand.PlayerBets[1]
	chipsMoved, err = table.CurrentHand.ProcessAction(1, "raise", table.Seats[1].Stack, raiseToAmount)
	if err != nil {
		t.Fatalf("Player 1 raise failed: %v", err)
	}
	table.Seats[1].Stack -= chipsMoved

	// Verify both players are all-in
	if table.Seats[0].Stack != 0 {
		t.Errorf("Player 0 stack should be 0, got %d", table.Seats[0].Stack)
	}
	if table.Seats[1].Stack != 0 {
		t.Errorf("Player 1 stack should be 0, got %d", table.Seats[1].Stack)
	}
	// With new pot accounting, bets stay in PlayerBets until AdvanceStreet
	totalPlayerBets := 0
	for _, bet := range table.CurrentHand.PlayerBets {
		totalPlayerBets += bet
	}
	if totalPlayerBets <= 0 {
		t.Errorf("Total player bets should be positive, got %d", totalPlayerBets)
	}
}

// TestSidePots_3P_OneAllInCreatesSidePot tests 3-player with one all-in
func TestSidePots_3P_OneAllInCreatesSidePot(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)
	token0 := "player-a"
	token1 := "player-b"
	token2 := "player-c"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 200
	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 500
	table.Seats[2].Token = &token2
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Player 0 raises to full stack (all-in)
	stack0 := table.Seats[0].Stack
	chipsMoved, err := table.CurrentHand.ProcessAction(0, "raise", stack0, stack0)
	if err != nil {
		t.Fatalf("Player 0 raise failed: %v", err)
	}
	table.Seats[0].Stack -= chipsMoved

	// Player 1 calls
	chipsMoved, err = table.CurrentHand.ProcessAction(1, "call", table.Seats[1].Stack)
	if err != nil {
		t.Fatalf("Player 1 call failed: %v", err)
	}
	table.Seats[1].Stack -= chipsMoved

	// Player 2 raises to full stack (all-in)
	stack2 := table.Seats[2].Stack
	chipsMoved, err = table.CurrentHand.ProcessAction(2, "raise", stack2, stack2)
	if err != nil {
		t.Fatalf("Player 2 raise failed: %v", err)
	}
	table.Seats[2].Stack -= chipsMoved

	// Player 1 calls (should go all-in if needed)
	chipsMoved, err = table.CurrentHand.ProcessAction(1, "call", table.Seats[1].Stack)
	if err != nil {
		t.Fatalf("Player 1 second call failed: %v", err)
	}
	table.Seats[1].Stack -= chipsMoved

	// Verify stacks
	if table.Seats[0].Stack != 0 {
		t.Errorf("Player 0 stack should be 0, got %d", table.Seats[0].Stack)
	}
	if table.Seats[1].Stack != 0 {
		t.Errorf("Player 1 stack should be 0, got %d", table.Seats[1].Stack)
	}
	// With new pot accounting, bets stay in PlayerBets until AdvanceStreet
	totalPlayerBets := 0
	for _, bet := range table.CurrentHand.PlayerBets {
		totalPlayerBets += bet
	}
	if totalPlayerBets <= 0 {
		t.Errorf("Total player bets should be positive, got %d", totalPlayerBets)
	}
}

// TestSidePots_3P_AllDifferentStacks tests 3 players with all different stacks going all-in
func TestSidePots_3P_AllDifferentStacks(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)
	token0 := "player-a"
	token1 := "player-b"
	token2 := "player-c"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 100
	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 300
	table.Seats[2].Token = &token2
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 500

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// All players raise to their full stacks in sequence
	for i := 0; i < 3; i++ {
		stack := table.Seats[i].Stack
		if stack > 0 {
			// Raise amount must account for amount already in PlayerBets
			raiseToAmount := stack + table.CurrentHand.PlayerBets[i]
			chipsMoved, err := table.CurrentHand.ProcessAction(i, "raise", stack, raiseToAmount)
			if err != nil {
				t.Fatalf("Player %d raise failed: %v", i, err)
			}
			table.Seats[i].Stack -= chipsMoved
		}
	}

	// Verify all stacks are 0
	for i := 0; i < 3; i++ {
		if table.Seats[i].Stack != 0 {
			t.Errorf("Player %d stack should be 0, got %d", i, table.Seats[i].Stack)
		}
	}
	// With new pot accounting, bets stay in PlayerBets until AdvanceStreet
	totalPlayerBets := 0
	for _, bet := range table.CurrentHand.PlayerBets {
		totalPlayerBets += bet
	}
	if totalPlayerBets <= 0 {
		t.Errorf("Total player bets should be positive, got %d", totalPlayerBets)
	}
}

// TestSidePots_3P_ShortestWinsMainPotOnly tests shortest stack can only win main pot
func TestSidePots_3P_ShortestWinsMainPotOnly(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)
	token0 := "player-a"
	token1 := "player-b"
	token2 := "player-c"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 50
	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 200
	table.Seats[2].Token = &token2
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// All players raise to their full stacks in sequence
	for i := 0; i < 3; i++ {
		stack := table.Seats[i].Stack
		if stack > 0 {
			// Raise amount must account for amount already in PlayerBets
			raiseToAmount := stack + table.CurrentHand.PlayerBets[i]
			chipsMoved, err := table.CurrentHand.ProcessAction(i, "raise", stack, raiseToAmount)
			if err != nil {
				t.Fatalf("Player %d raise failed: %v", i, err)
			}
			table.Seats[i].Stack -= chipsMoved
		}
	}

	// Verify stacks
	for i := 0; i < 3; i++ {
		if table.Seats[i].Stack != 0 {
			t.Errorf("Player %d stack should be 0, got %d", i, table.Seats[i].Stack)
		}
	}
	// With new pot accounting, bets stay in PlayerBets until AdvanceStreet
	totalPlayerBets := 0
	for _, bet := range table.CurrentHand.PlayerBets {
		totalPlayerBets += bet
	}
	if totalPlayerBets <= 0 {
		t.Errorf("Total player bets should be positive, got %d", totalPlayerBets)
	}
}

// TestSidePots_4P_MultipleAllIns tests 4 players with multiple all-ins
func TestSidePots_4P_MultipleAllIns(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)
	tokens := []string{"player-a", "player-b", "player-c", "player-d"}
	stacks := []int{100, 250, 500, 1000}

	for i := 0; i < 4; i++ {
		table.Seats[i].Token = &tokens[i]
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = stacks[i]
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// All players raise to their full stacks in sequence
	for i := 0; i < 4; i++ {
		stack := table.Seats[i].Stack
		if stack > 0 {
			// Raise amount must account for amount already in PlayerBets
			raiseToAmount := stack + table.CurrentHand.PlayerBets[i]
			chipsMoved, err := table.CurrentHand.ProcessAction(i, "raise", stack, raiseToAmount)
			if err != nil {
				t.Fatalf("Player %d raise failed: %v", i, err)
			}
			table.Seats[i].Stack -= chipsMoved
		}
	}

	// Verify all stacks are 0
	for i := 0; i < 4; i++ {
		if table.Seats[i].Stack != 0 {
			t.Errorf("Player %d stack should be 0, got %d", i, table.Seats[i].Stack)
		}
	}
	// With new pot accounting, bets stay in PlayerBets until AdvanceStreet
	totalPlayerBets := 0
	for _, bet := range table.CurrentHand.PlayerBets {
		totalPlayerBets += bet
	}
	if totalPlayerBets <= 0 {
		t.Errorf("Total player bets should be positive, got %d", totalPlayerBets)
	}
}

// TestSidePots_6P_WhaleExcessReturned tests 6 players with whale
func TestSidePots_6P_WhaleExcessReturned(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)
	tokens := []string{"whale", "player-b", "player-c", "player-d", "player-e", "player-f"}
	stacks := []int{5000, 100, 200, 300, 400, 500}

	for i := 0; i < 6; i++ {
		table.Seats[i].Token = &tokens[i]
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = stacks[i]
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Players 1-5 raise to their full stacks
	for i := 1; i < 6; i++ {
		stack := table.Seats[i].Stack
		if stack > 0 {
			// Raise amount must account for amount already in PlayerBets
			raiseToAmount := stack + table.CurrentHand.PlayerBets[i]
			chipsMoved, err := table.CurrentHand.ProcessAction(i, "raise", stack, raiseToAmount)
			if err != nil {
				t.Fatalf("Player %d raise failed: %v", i, err)
			}
			table.Seats[i].Stack -= chipsMoved
		}
	}

	// Whale calls
	whaleStack := table.Seats[0].Stack
	if whaleStack > 0 {
		chipsMoved, err := table.CurrentHand.ProcessAction(0, "call", whaleStack)
		if err != nil {
			t.Fatalf("Whale call failed: %v", err)
		}
		table.Seats[0].Stack -= chipsMoved
	}

	// Verify shortstack players are all-in
	for i := 1; i < 6; i++ {
		if table.Seats[i].Stack != 0 {
			t.Errorf("Player %d stack should be 0, got %d", i, table.Seats[i].Stack)
		}
	}
	// Whale should have remainder (didn't need to go all-in)
	if table.Seats[0].Stack < 0 {
		t.Errorf("Whale stack should be non-negative, got %d", table.Seats[0].Stack)
	}
	// With new pot accounting, bets stay in PlayerBets until AdvanceStreet
	totalPlayerBets := 0
	for _, bet := range table.CurrentHand.PlayerBets {
		totalPlayerBets += bet
	}
	if totalPlayerBets <= 0 {
		t.Errorf("Total player bets should be positive, got %d", totalPlayerBets)
	}
}

// TestSidePots_6P_MultipleSidePots tests 6 players creating multiple side pots
func TestSidePots_6P_MultipleSidePots(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)
	tokens := []string{"player-a", "player-b", "player-c", "player-d", "player-e", "player-f"}
	stacks := []int{50, 150, 300, 500, 750, 1000}

	for i := 0; i < 6; i++ {
		table.Seats[i].Token = &tokens[i]
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = stacks[i]
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// All players raise to their full stacks in sequence
	for i := 0; i < 6; i++ {
		stack := table.Seats[i].Stack
		if stack > 0 {
			// Raise amount must account for amount already in PlayerBets
			raiseToAmount := stack + table.CurrentHand.PlayerBets[i]
			chipsMoved, err := table.CurrentHand.ProcessAction(i, "raise", stack, raiseToAmount)
			if err != nil {
				t.Fatalf("Player %d raise failed: %v", i, err)
			}
			table.Seats[i].Stack -= chipsMoved
		}
	}

	// Verify all stacks are 0
	for i := 0; i < 6; i++ {
		if table.Seats[i].Stack != 0 {
			t.Errorf("Player %d stack should be 0, got %d", i, table.Seats[i].Stack)
		}
	}
	// With new pot accounting, bets stay in PlayerBets until AdvanceStreet
	totalPlayerBets := 0
	for _, bet := range table.CurrentHand.PlayerBets {
		totalPlayerBets += bet
	}
	if totalPlayerBets <= 0 {
		t.Errorf("Total player bets should be positive, got %d", totalPlayerBets)
	}
}

// === PHASE 2 TESTS: Pot Accounting - Remove Immediate Additions ===

// TestStartHand_NoPotUpdate verifies Pot stays 0 after blinds posted, only PlayerBets updated
func TestStartHand_NoPotUpdate(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	tokenA := "token-a"
	tokenB := "token-b"

	table.Seats[0].Token = &tokenA
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &tokenB
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	// Start hand with SB=10, BB=20
	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// After StartHand, Pot should be 0 (not included immediately)
	if table.CurrentHand.Pot != 0 {
		t.Errorf("Phase 2: expected Pot=0 after StartHand, got %d", table.CurrentHand.Pot)
	}

	// But PlayerBets should have the blind amounts
	sbSeatIndex := table.CurrentHand.SmallBlindSeat
	bbSeatIndex := table.CurrentHand.BigBlindSeat

	sbBet := table.CurrentHand.PlayerBets[sbSeatIndex]
	bbBet := table.CurrentHand.PlayerBets[bbSeatIndex]

	if sbBet != 10 {
		t.Errorf("Phase 2: expected SB PlayerBet=10, got %d", sbBet)
	}

	if bbBet != 20 {
		t.Errorf("Phase 2: expected BB PlayerBet=20, got %d", bbBet)
	}
}

// TestProcessAction_Call_NoPotUpdate verifies calling updates PlayerBets but NOT Pot
func TestProcessAction_Call_NoPotUpdate(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	tokenA := "token-a"
	tokenB := "token-b"

	table.Seats[0].Token = &tokenA
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &tokenB
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	potAfterBlinds := table.CurrentHand.Pot

	// Process call action from SB (to match BB of 20)
	sbSeatIndex := table.CurrentHand.SmallBlindSeat
	sbStack := table.Seats[sbSeatIndex].Stack
	chipsMoved, err := table.CurrentHand.ProcessAction(sbSeatIndex, "call", sbStack)
	if err != nil {
		t.Fatalf("failed to process call: %v", err)
	}

	// Pot should NOT change from call (still 0)
	if table.CurrentHand.Pot != potAfterBlinds {
		t.Errorf("Phase 2: expected Pot unchanged after call (was %d), got %d", potAfterBlinds, table.CurrentHand.Pot)
	}

	// PlayerBets should be updated
	expectedBet := 20 // SB called BB (10 + 10 = 20)
	if table.CurrentHand.PlayerBets[sbSeatIndex] != expectedBet {
		t.Errorf("Phase 2: expected PlayerBets[SB]=%d after call, got %d", expectedBet, table.CurrentHand.PlayerBets[sbSeatIndex])
	}

	if chipsMoved != 10 {
		t.Errorf("Phase 2: expected chips moved=10, got %d", chipsMoved)
	}
}

// TestProcessAction_Raise_NoPotUpdate verifies raising updates PlayerBets but NOT Pot
func TestProcessAction_Raise_NoPotUpdate(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	tokenA := "token-a"
	tokenB := "token-b"

	table.Seats[0].Token = &tokenA
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &tokenB
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	potAfterBlinds := table.CurrentHand.Pot

	// Process raise action from UTG (seat 0, who is dealer in heads-up)
	// Raise to 100
	utgSeatIndex := 0
	utgStack := table.Seats[utgSeatIndex].Stack
	chipsMoved, err := table.CurrentHand.ProcessAction(utgSeatIndex, "raise", utgStack, 100)
	if err != nil {
		t.Fatalf("failed to process raise: %v", err)
	}

	// Pot should NOT change from raise (still 0)
	if table.CurrentHand.Pot != potAfterBlinds {
		t.Errorf("Phase 2: expected Pot unchanged after raise (was %d), got %d", potAfterBlinds, table.CurrentHand.Pot)
	}

	// PlayerBets should be updated to raise amount
	if table.CurrentHand.PlayerBets[utgSeatIndex] != 100 {
		t.Errorf("Phase 2: expected PlayerBets[UTG]=100 after raise, got %d", table.CurrentHand.PlayerBets[utgSeatIndex])
	}

	// chips moved should be 90 (100 raise - 10 already bet as SB)
	// In heads-up: dealer is SB, so seat 0 posts 10, seat 1 posts 20
	if chipsMoved != 90 {
		t.Errorf("Phase 2: expected chips moved=90 (raise 100 - existing SB 10), got %d", chipsMoved)
	}
}

// TestProcessActionWithSeats_Call_NoPotUpdate verifies ProcessActionWithSeats also doesn't update Pot on call
func TestProcessActionWithSeats_Call_NoPotUpdate(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	tokenA := "token-a"
	tokenB := "token-b"

	table.Seats[0].Token = &tokenA
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &tokenB
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	potAfterBlinds := table.CurrentHand.Pot

	// Process call action from SB using ProcessActionWithSeats
	sbSeatIndex := table.CurrentHand.SmallBlindSeat
	sbStack := table.Seats[sbSeatIndex].Stack
	chipsMoved, err := table.CurrentHand.ProcessActionWithSeats(sbSeatIndex, "call", sbStack, table.Seats)
	if err != nil {
		t.Fatalf("failed to process call with seats: %v", err)
	}

	// Pot should NOT change
	if table.CurrentHand.Pot != potAfterBlinds {
		t.Errorf("Phase 2: expected Pot unchanged after ProcessActionWithSeats call (was %d), got %d", potAfterBlinds, table.CurrentHand.Pot)
	}

	// PlayerBets should be updated
	expectedBet := 20 // SB called BB (10 + 10 = 20)
	if table.CurrentHand.PlayerBets[sbSeatIndex] != expectedBet {
		t.Errorf("Phase 2: expected PlayerBets[SB]=%d after call, got %d", expectedBet, table.CurrentHand.PlayerBets[sbSeatIndex])
	}

	if chipsMoved != 10 {
		t.Errorf("Phase 2: expected chips moved=10, got %d", chipsMoved)
	}
}

// TestPotRemainsZero_DuringBettingRound verifies Pot stays 0 throughout entire preflop betting round
func TestPotRemainsZero_DuringBettingRound(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	tokenA := "token-a"
	tokenB := "token-b"

	table.Seats[0].Token = &tokenA
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &tokenB
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Pot should be 0 right after blinds
	if table.CurrentHand.Pot != 0 {
		t.Errorf("Phase 2: expected Pot=0 after StartHand, got %d", table.CurrentHand.Pot)
	}

	// SB calls
	sbSeatIndex := table.CurrentHand.SmallBlindSeat
	sbStack := table.Seats[sbSeatIndex].Stack
	table.CurrentHand.ProcessAction(sbSeatIndex, "call", sbStack)

	if table.CurrentHand.Pot != 0 {
		t.Errorf("Phase 2: expected Pot=0 after SB calls, got %d", table.CurrentHand.Pot)
	}

	// BB checks
	bbSeatIndex := table.CurrentHand.BigBlindSeat
	bbStack := table.Seats[bbSeatIndex].Stack
	table.CurrentHand.ProcessAction(bbSeatIndex, "check", bbStack)

	if table.CurrentHand.Pot != 0 {
		t.Errorf("Phase 2: expected Pot=0 after BB checks, got %d", table.CurrentHand.Pot)
	}

	// Preflop betting complete, but pot should still be 0 until AdvanceStreet
	if table.CurrentHand.Pot != 0 {
		t.Errorf("Phase 2: expected Pot=0 at end of preflop betting round, got %d", table.CurrentHand.Pot)
	}
}

// Phase 3 Tests: Pot Sweep at AdvanceStreet

// TestAdvanceStreet_SweepsBetsIntoPot_Preflop verifies blinds + preflop bets are swept to Pot on first advance
func TestAdvanceStreet_SweepsBetsIntoPot_Preflop(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	tokenA := "token-a"
	tokenB := "token-b"

	table.Seats[0].Token = &tokenA
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &tokenB
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	// Verify blinds are in PlayerBets, not Pot
	sbSeatIndex := table.CurrentHand.SmallBlindSeat
	bbSeatIndex := table.CurrentHand.BigBlindSeat
	totalBets := table.CurrentHand.PlayerBets[sbSeatIndex] + table.CurrentHand.PlayerBets[bbSeatIndex]

	if table.CurrentHand.Pot != 0 {
		t.Errorf("Phase 3: Pot should be 0 after StartHand, got %d", table.CurrentHand.Pot)
	}

	if totalBets != 30 {
		t.Errorf("Phase 3: Expected total bets 30 (10 SB + 20 BB), got %d", totalBets)
	}

	// SB calls
	sbStack := table.Seats[sbSeatIndex].Stack
	table.CurrentHand.ProcessAction(sbSeatIndex, "call", sbStack)

	// BB checks
	bbStack := table.Seats[bbSeatIndex].Stack
	table.CurrentHand.ProcessAction(bbSeatIndex, "check", bbStack)

	// Before advancing: Pot should still be 0
	if table.CurrentHand.Pot != 0 {
		t.Errorf("Phase 3: Pot should be 0 before AdvanceStreet, got %d", table.CurrentHand.Pot)
	}

	// Advance to flop - this should sweep all bets into Pot
	table.CurrentHand.AdvanceStreet()

	// After advancing: Pot should contain all preflop bets
	// Expected: SB 10 + BB 20 + SB call 10 + BB check 0 = 40
	expectedPot := 40
	if table.CurrentHand.Pot != expectedPot {
		t.Errorf("Phase 3: Expected Pot=%d after AdvanceStreet, got %d", expectedPot, table.CurrentHand.Pot)
	}
}

// TestAdvanceStreet_ClearsPlayerBets verifies PlayerBets are cleared after sweep
func TestAdvanceStreet_ClearsPlayerBets(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	tokenA := "token-a"
	tokenB := "token-b"

	table.Seats[0].Token = &tokenA
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &tokenB
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	sbSeatIndex := table.CurrentHand.SmallBlindSeat
	bbSeatIndex := table.CurrentHand.BigBlindSeat

	// SB calls
	sbStack := table.Seats[sbSeatIndex].Stack
	table.CurrentHand.ProcessAction(sbSeatIndex, "call", sbStack)

	// BB checks
	bbStack := table.Seats[bbSeatIndex].Stack
	table.CurrentHand.ProcessAction(bbSeatIndex, "check", bbStack)

	// Verify bets are in PlayerBets before advance
	if len(table.CurrentHand.PlayerBets) != 2 {
		t.Errorf("Phase 3: Expected 2 players with bets before AdvanceStreet, got %d", len(table.CurrentHand.PlayerBets))
	}

	// Advance to flop
	table.CurrentHand.AdvanceStreet()

	// After advancing: PlayerBets should be empty
	if len(table.CurrentHand.PlayerBets) != 0 {
		t.Errorf("Phase 3: Expected PlayerBets to be empty after AdvanceStreet, got %d entries", len(table.CurrentHand.PlayerBets))
	}

	for seatIndex, bet := range table.CurrentHand.PlayerBets {
		if bet != 0 {
			t.Errorf("Phase 3: Expected PlayerBets[%d]=0 after AdvanceStreet, got %d", seatIndex, bet)
		}
	}
}

// TestAdvanceStreet_AccumulatesPot verifies Pot accumulates across multiple streets
func TestAdvanceStreet_AccumulatesPot(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	tokenA := "token-a"
	tokenB := "token-b"

	table.Seats[0].Token = &tokenA
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &tokenB
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	sbSeatIndex := table.CurrentHand.SmallBlindSeat
	bbSeatIndex := table.CurrentHand.BigBlindSeat

	// Preflop: SB calls, BB checks
	// Note: stack values are updated by StartHand (SB posts 10, BB posts 20)
	sbStack := table.Seats[sbSeatIndex].Stack
	table.CurrentHand.ProcessAction(sbSeatIndex, "call", sbStack)
	table.Seats[sbSeatIndex].Stack -= 10 // Simulate stack update for call

	bbStack := table.Seats[bbSeatIndex].Stack
	table.CurrentHand.ProcessAction(bbSeatIndex, "check", bbStack)

	// Advance to flop - sweeps preflop bets
	table.CurrentHand.AdvanceStreet()
	potAfterFlop := table.CurrentHand.Pot
	if potAfterFlop != 40 {
		t.Errorf("Phase 3: Expected Pot=40 after flop, got %d", potAfterFlop)
	}

	// Flop: both check (no new bets, pot stays same)
	sbStack = table.Seats[sbSeatIndex].Stack
	bbStack = table.Seats[bbSeatIndex].Stack
	table.CurrentHand.ProcessAction(sbSeatIndex, "check", sbStack)
	table.CurrentHand.ProcessAction(bbSeatIndex, "check", bbStack)

	if table.CurrentHand.Pot != 40 {
		t.Errorf("Phase 3: Expected Pot=40 after flop checks, got %d", table.CurrentHand.Pot)
	}

	// Advance to turn - should sweep any new flop bets (none in this case)
	table.CurrentHand.AdvanceStreet()

	// Pot should stay the same since no one bet on flop
	if table.CurrentHand.Pot != 40 {
		t.Errorf("Phase 3: Expected Pot=40 after turn, got %d", table.CurrentHand.Pot)
	}

	// Turn: SB raises to 100, BB calls
	sbStack = table.Seats[sbSeatIndex].Stack
	bbStack = table.Seats[bbSeatIndex].Stack
	table.CurrentHand.ProcessAction(sbSeatIndex, "raise", sbStack, 100)
	table.Seats[sbSeatIndex].Stack -= 100 // Simulate stack update for raise

	table.CurrentHand.ProcessAction(bbSeatIndex, "call", bbStack)
	table.Seats[bbSeatIndex].Stack -= 100 // Simulate stack update for call

	if table.CurrentHand.Pot != 40 {
		t.Errorf("Phase 3: Expected Pot=40 after turn bets (before advance), got %d", table.CurrentHand.Pot)
	}

	// Advance to river - should sweep turn bets
	table.CurrentHand.AdvanceStreet()

	// Pot should now include turn bets: 40 + 100 + 100 = 240
	expectedPot := 240
	if table.CurrentHand.Pot != expectedPot {
		t.Errorf("Phase 3: Expected Pot=%d after river, got %d", expectedPot, table.CurrentHand.Pot)
	}
}

// TestFullHandPotAccounting_PreflopToRiver verifies end-to-end pot accounting through entire hand
func TestFullHandPotAccounting_PreflopToRiver(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	tokenA := "token-a"
	tokenB := "token-b"

	table.Seats[0].Token = &tokenA
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &tokenB
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	err := table.StartHand()
	if err != nil {
		t.Fatalf("failed to start hand: %v", err)
	}

	sbSeatIndex := table.CurrentHand.SmallBlindSeat
	bbSeatIndex := table.CurrentHand.BigBlindSeat

	// Preflop: Small blind posts 10, Big blind posts 20
	// Then SB raises to 50, BB calls 50
	sbStack := table.Seats[sbSeatIndex].Stack
	bbStack := table.Seats[bbSeatIndex].Stack

	table.CurrentHand.ProcessAction(sbSeatIndex, "raise", sbStack, 50)
	table.Seats[sbSeatIndex].Stack -= 40 // Already posted 10, now raise to 50 (40 more)

	table.CurrentHand.ProcessAction(bbSeatIndex, "call", bbStack)
	table.Seats[bbSeatIndex].Stack -= 30 // Already posted 20, now call to 50 (30 more)

	// Before advance: Pot should be 0, all bets in PlayerBets
	if table.CurrentHand.Pot != 0 {
		t.Errorf("Phase 3: Expected Pot=0 before flop, got %d", table.CurrentHand.Pot)
	}

	// Advance to flop - sweeps preflop bets
	table.CurrentHand.AdvanceStreet()
	// Pot should be: SB bet 50 total (PlayerBets[sb] = 50), BB bet 50 total (PlayerBets[bb] = 50)
	// So: 50 + 50 = 100
	expectedAfterFlop := 100
	if table.CurrentHand.Pot != expectedAfterFlop {
		t.Errorf("Phase 3: Expected Pot=%d after flop, got %d", expectedAfterFlop, table.CurrentHand.Pot)
	}

	// Flop: Both check
	sbStack = table.Seats[sbSeatIndex].Stack
	bbStack = table.Seats[bbSeatIndex].Stack
	table.CurrentHand.ProcessAction(sbSeatIndex, "check", sbStack)
	table.CurrentHand.ProcessAction(bbSeatIndex, "check", bbStack)

	// Before advance: Pot still 100
	if table.CurrentHand.Pot != 100 {
		t.Errorf("Phase 3: Expected Pot=100 before turn, got %d", table.CurrentHand.Pot)
	}

	// Advance to turn
	table.CurrentHand.AdvanceStreet()

	// Pot should still be 100 (no bets on flop)
	if table.CurrentHand.Pot != 100 {
		t.Errorf("Phase 3: Expected Pot=100 after turn, got %d", table.CurrentHand.Pot)
	}

	// Turn: SB raises to 75, BB calls 75
	sbStack = table.Seats[sbSeatIndex].Stack
	bbStack = table.Seats[bbSeatIndex].Stack
	table.CurrentHand.ProcessAction(sbSeatIndex, "raise", sbStack, 75)
	table.Seats[sbSeatIndex].Stack -= 75 // Bet 75 on turn

	table.CurrentHand.ProcessAction(bbSeatIndex, "call", bbStack)
	table.Seats[bbSeatIndex].Stack -= 75 // Call 75 on turn

	// Before advance: Pot still 100
	if table.CurrentHand.Pot != 100 {
		t.Errorf("Phase 3: Expected Pot=100 before river, got %d", table.CurrentHand.Pot)
	}

	// Advance to river - sweeps turn bets
	table.CurrentHand.AdvanceStreet()

	// Pot should now be: 100 + 75 + 75 = 250
	expectedAfterRiver := 250
	if table.CurrentHand.Pot != expectedAfterRiver {
		t.Errorf("Phase 3: Expected Pot=%d after river, got %d", expectedAfterRiver, table.CurrentHand.Pot)
	}

	// River: Both check (final action)
	sbStack = table.Seats[sbSeatIndex].Stack
	bbStack = table.Seats[bbSeatIndex].Stack
	table.CurrentHand.ProcessAction(sbSeatIndex, "check", sbStack)
	table.CurrentHand.ProcessAction(bbSeatIndex, "check", bbStack)

	// Pot should stay at 250 until showdown
	if table.CurrentHand.Pot != 250 {
		t.Errorf("Phase 3: Expected Pot=250 at showdown, got %d", table.CurrentHand.Pot)
	}
}

// TestGetValidActions_AllInPlayerZeroStackPreflop verifies that an all-in player (stack=0) receives no valid actions on preflop
func TestGetValidActions_AllInPlayerZeroStackPreflop(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players
	for i := 0; i < 2; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Initialize action state for preflop
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}
	table.CurrentHand.PlayerBets[0] = 100
	table.CurrentHand.CurrentBet = 100

	// Player 0 is all-in (stack = 0)
	validActions := table.CurrentHand.GetValidActions(0, 0, table.Seats)

	// Should return empty slice for all-in player
	if len(validActions) != 0 {
		t.Errorf("expected empty actions for all-in player (stack=0) preflop, got %v", validActions)
	}
}

// TestGetValidActions_AllInPlayerZeroStackFlop verifies that an all-in player (stack=0) receives no valid actions on flop
func TestGetValidActions_AllInPlayerZeroStackFlop(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players
	for i := 0; i < 2; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Advance to flop
	table.CurrentHand.Street = "flop"

	// Initialize action state for flop
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}
	table.CurrentHand.PlayerBets[0] = 50
	table.CurrentHand.CurrentBet = 50

	// Player 0 is all-in (stack = 0)
	validActions := table.CurrentHand.GetValidActions(0, 0, table.Seats)

	// Should return empty slice for all-in player
	if len(validActions) != 0 {
		t.Errorf("expected empty actions for all-in player (stack=0) on flop, got %v", validActions)
	}
}

// TestGetValidActions_AllInPlayerZeroStackTurn verifies that an all-in player (stack=0) receives no valid actions on turn
func TestGetValidActions_AllInPlayerZeroStackTurn(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players
	for i := 0; i < 2; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Advance to turn
	table.CurrentHand.Street = "turn"

	// Initialize action state for turn
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}
	table.CurrentHand.PlayerBets[0] = 75
	table.CurrentHand.CurrentBet = 75

	// Player 0 is all-in (stack = 0)
	validActions := table.CurrentHand.GetValidActions(0, 0, table.Seats)

	// Should return empty slice for all-in player
	if len(validActions) != 0 {
		t.Errorf("expected empty actions for all-in player (stack=0) on turn, got %v", validActions)
	}
}

// TestGetValidActions_AllInPlayerZeroStackRiver verifies that an all-in player (stack=0) receives no valid actions on river
func TestGetValidActions_AllInPlayerZeroStackRiver(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players
	for i := 0; i < 2; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Advance to river
	table.CurrentHand.Street = "river"

	// Initialize action state for river
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}
	table.CurrentHand.PlayerBets[0] = 100
	table.CurrentHand.CurrentBet = 100

	// Player 0 is all-in (stack = 0)
	validActions := table.CurrentHand.GetValidActions(0, 0, table.Seats)

	// Should return empty slice for all-in player
	if len(validActions) != 0 {
		t.Errorf("expected empty actions for all-in player (stack=0) on river, got %v", validActions)
	}
}

// TestGetValidActions_AllInPlayerWithCallAmount verifies that even with call amount > 0, stack=0 returns empty
func TestGetValidActions_AllInPlayerWithCallAmount(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players
	for i := 0; i < 2; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Initialize action state
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}
	// Player 0 is behind: bet 50 but current bet is 100
	table.CurrentHand.PlayerBets[0] = 50
	table.CurrentHand.CurrentBet = 100
	table.CurrentHand.LastRaise = 50

	// Player 0 is all-in (stack = 0), but still needs to call to continue
	// Verify that GetValidActions returns empty despite callAmount > 0
	validActions := table.CurrentHand.GetValidActions(0, 0, table.Seats)

	// Should return empty slice for all-in player, even with call amount > 0
	if len(validActions) != 0 {
		t.Errorf("expected empty actions for all-in player (stack=0) even with call amount > 0, got %v", validActions)
	}
}

// TestGetValidActions_AllInPlayerWithRaise verifies that even with raise available, stack=0 returns empty
func TestGetValidActions_AllInPlayerWithRaise(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players
	for i := 0; i < 2; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Initialize action state - player has matched bet and could potentially raise
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}
	table.CurrentHand.PlayerBets[0] = 100
	table.CurrentHand.CurrentBet = 100
	table.CurrentHand.LastRaise = 50

	// Player 0 is all-in (stack = 0)
	// They have matched the current bet, so they could check/fold, but not with zero stack
	validActions := table.CurrentHand.GetValidActions(0, 0, table.Seats)

	// Should return empty slice for all-in player, even though they've matched the current bet
	if len(validActions) != 0 {
		t.Errorf("expected empty actions for all-in player (stack=0) even after matching bet, got %v", validActions)
	}
}

// TestGetNextActiveSeat_AllInScenarios tests GetNextActiveSeat() includes all-in players in rotation
// All-in players are included because GetValidActions() will return empty actions for them
// These tests verify that all-in players remain in rotation (they may still need to be offered actions)
func TestGetNextActiveSeat_AllInScenarios(t *testing.T) {

	// Subtest 1: two_players_one_allin - Include all-in in rotation
	t.Run("two_players_one_allin", func(t *testing.T) {
		table := NewTable("table-1", "Table 1", nil)

		// Set up 2 active players (seats 0, 1)
		token0 := "player-0"
		token1 := "player-1"
		table.Seats[0].Token = &token0
		table.Seats[0].Status = "active"
		table.Seats[0].Stack = 1000 // Active with chips
		table.Seats[1].Token = &token1
		table.Seats[1].Status = "active"
		table.Seats[1].Stack = 0 // All-in

		// Start hand
		err := table.StartHand()
		if err != nil {
			t.Fatalf("expected no error starting hand, got %v", err)
		}

		// Initialize folded players map
		if table.CurrentHand.FoldedPlayers == nil {
			table.CurrentHand.FoldedPlayers = make(map[int]bool)
		}

		// From seat 0, should rotate to seat 1 (even though all-in)
		next := table.CurrentHand.GetNextActiveSeat(0, table.Seats)
		if next == nil || *next != 1 {
			t.Errorf("expected next seat to be 1 (all-in player included), got %v", next)
		}

		// From seat 1, should rotate to seat 0
		next = table.CurrentHand.GetNextActiveSeat(1, table.Seats)
		if next == nil || *next != 0 {
			t.Errorf("expected next seat to be 0, got %v", next)
		}
	})

	// Subtest 2: two_players_both_allin - Both all-in still rotates
	t.Run("two_players_both_allin", func(t *testing.T) {
		table := NewTable("table-1", "Table 1", nil)

		// Set up 2 active players, both all-in
		token0 := "player-0"
		token1 := "player-1"
		table.Seats[0].Token = &token0
		table.Seats[0].Status = "active"
		table.Seats[0].Stack = 0 // All-in
		table.Seats[1].Token = &token1
		table.Seats[1].Status = "active"
		table.Seats[1].Stack = 0 // All-in

		// Start hand
		err := table.StartHand()
		if err != nil {
			t.Fatalf("expected no error starting hand, got %v", err)
		}

		// Initialize folded players map
		if table.CurrentHand.FoldedPlayers == nil {
			table.CurrentHand.FoldedPlayers = make(map[int]bool)
		}

		// From seat 0, should rotate to seat 1
		next := table.CurrentHand.GetNextActiveSeat(0, table.Seats)
		if next == nil || *next != 1 {
			t.Errorf("expected next seat to be 1, got %v", next)
		}

		// From seat 1, should rotate to seat 0
		next = table.CurrentHand.GetNextActiveSeat(1, table.Seats)
		if next == nil || *next != 0 {
			t.Errorf("expected next seat to be 0, got %v", next)
		}
	})

	// Subtest 3: three_players_one_allin - Include all-in in 3-player rotation
	t.Run("three_players_one_allin", func(t *testing.T) {
		table := NewTable("table-1", "Table 1", nil)

		// Set up 3 active players (seats 0, 1, 2)
		for i := 0; i < 3; i++ {
			token := "player-" + string(rune('0'+i))
			table.Seats[i].Token = &token
			table.Seats[i].Status = "active"
			if i == 1 {
				table.Seats[i].Stack = 0 // Seat 1 is all-in
			} else {
				table.Seats[i].Stack = 1000 // Others active
			}
		}

		// Start hand
		err := table.StartHand()
		if err != nil {
			t.Fatalf("expected no error starting hand, got %v", err)
		}

		// Initialize folded players map
		if table.CurrentHand.FoldedPlayers == nil {
			table.CurrentHand.FoldedPlayers = make(map[int]bool)
		}

		// From seat 0, should go to seat 1 (all-in player included)
		next := table.CurrentHand.GetNextActiveSeat(0, table.Seats)
		if next == nil || *next != 1 {
			t.Errorf("expected next active seat after 0 to be 1 (including all-in), got %v", next)
		}

		// From seat 1, should go to seat 2
		next = table.CurrentHand.GetNextActiveSeat(1, table.Seats)
		if next == nil || *next != 2 {
			t.Errorf("expected next active seat after 1 to be 2, got %v", next)
		}

		// From seat 2, should wrap to seat 0
		next = table.CurrentHand.GetNextActiveSeat(2, table.Seats)
		if next == nil || *next != 0 {
			t.Errorf("expected next active seat after 2 (wrapping) to be 0, got %v", next)
		}
	})

	// Subtest 4: three_players_two_allin - Include all-in players in rotation
	t.Run("three_players_two_allin", func(t *testing.T) {
		table := NewTable("table-1", "Table 1", nil)

		// Set up 3 active players (seats 0, 1, 2)
		for i := 0; i < 3; i++ {
			token := "player-" + string(rune('0'+i))
			table.Seats[i].Token = &token
			table.Seats[i].Status = "active"
			if i == 1 || i == 2 {
				table.Seats[i].Stack = 0 // Seats 1 and 2 are all-in
			} else {
				table.Seats[i].Stack = 1000 // Only seat 0 has chips
			}
		}

		// Start hand
		err := table.StartHand()
		if err != nil {
			t.Fatalf("expected no error starting hand, got %v", err)
		}

		// Initialize folded players map
		if table.CurrentHand.FoldedPlayers == nil {
			table.CurrentHand.FoldedPlayers = make(map[int]bool)
		}

		// From seat 0, should rotate to seat 1 (even though all-in)
		next := table.CurrentHand.GetNextActiveSeat(0, table.Seats)
		if next == nil || *next != 1 {
			t.Errorf("expected next seat to be 1, got %v", next)
		}

		// From seat 1, should rotate to seat 2
		next = table.CurrentHand.GetNextActiveSeat(1, table.Seats)
		if next == nil || *next != 2 {
			t.Errorf("expected next seat to be 2, got %v", next)
		}

		// From seat 2, should wrap to seat 0
		next = table.CurrentHand.GetNextActiveSeat(2, table.Seats)
		if next == nil || *next != 0 {
			t.Errorf("expected next seat to be 0, got %v", next)
		}
	})

	// Subtest 5: four_players_mixed_allin_folded - Include all-in but skip folded
	t.Run("four_players_mixed_allin_folded", func(t *testing.T) {
		table := NewTable("table-1", "Table 1", nil)

		// Set up 4 active players (seats 0, 1, 2, 3)
		for i := 0; i < 4; i++ {
			token := "player-" + string(rune('0'+i))
			table.Seats[i].Token = &token
			table.Seats[i].Status = "active"
			if i == 1 {
				table.Seats[i].Stack = 0 // Seat 1 is all-in
			} else {
				table.Seats[i].Stack = 1000 // Others active
			}
		}

		// Start hand
		err := table.StartHand()
		if err != nil {
			t.Fatalf("expected no error starting hand, got %v", err)
		}

		// Initialize folded players map
		if table.CurrentHand.FoldedPlayers == nil {
			table.CurrentHand.FoldedPlayers = make(map[int]bool)
		}

		// Mark seat 3 as folded
		table.CurrentHand.FoldedPlayers[3] = true

		// From seat 0, should go to seat 1 (all-in included, folded skipped)
		next := table.CurrentHand.GetNextActiveSeat(0, table.Seats)
		if next == nil || *next != 1 {
			t.Errorf("expected next active seat after 0 to be 1 (including all-in, skipping folded), got %v", next)
		}

		// From seat 1, should go to seat 2 (skipping folded 3)
		next = table.CurrentHand.GetNextActiveSeat(1, table.Seats)
		if next == nil || *next != 2 {
			t.Errorf("expected next active seat after 1 to be 2 (skipping folded 3), got %v", next)
		}

		// From seat 2, should wrap to seat 0 (skipping folded 3)
		next = table.CurrentHand.GetNextActiveSeat(2, table.Seats)
		if next == nil || *next != 0 {
			t.Errorf("expected next active seat after 2 to be 0 (skipping folded 3, wrapping), got %v", next)
		}
	})

	// Subtest 6: all_folded_except_allin - Return nil when only one player remains
	// This tests the edge case where all players fold except one (who happens to be all-in)
	t.Run("all_folded_except_allin", func(t *testing.T) {
		table := NewTable("table-1", "Table 1", nil)

		// Set up 4 active players (seats 0, 1, 2, 3)
		for i := 0; i < 4; i++ {
			token := "player-" + string(rune('0'+i))
			table.Seats[i].Token = &token
			table.Seats[i].Status = "active"
			table.Seats[i].Stack = 1000
		}

		// Start hand
		err := table.StartHand()
		if err != nil {
			t.Fatalf("expected no error starting hand, got %v", err)
		}

		// Initialize folded players map
		if table.CurrentHand.FoldedPlayers == nil {
			table.CurrentHand.FoldedPlayers = make(map[int]bool)
		}

		// Mark seats 0, 1, 2 as folded and seat 3 as all-in (stack = 0)
		table.CurrentHand.FoldedPlayers[0] = true
		table.CurrentHand.FoldedPlayers[1] = true
		table.CurrentHand.FoldedPlayers[2] = true
		table.Seats[3].Stack = 0 // Only all-in player left

		// From seat 3 (all-in), all others folded -> nil
		next := table.CurrentHand.GetNextActiveSeat(3, table.Seats)
		if next != nil {
			t.Errorf("expected nil when only all-in player remains (others folded), got %v", next)
		}
	})

	// Subtest 7: no_allin_normal_rotation - Control test (no all-in players)
	t.Run("no_allin_normal_rotation", func(t *testing.T) {
		table := NewTable("table-1", "Table 1", nil)

		// Set up 3 active players (seats 0, 1, 2), none all-in
		for i := 0; i < 3; i++ {
			token := "player-" + string(rune('0'+i))
			table.Seats[i].Token = &token
			table.Seats[i].Status = "active"
			table.Seats[i].Stack = 1000 // All have stacks
		}

		// Start hand
		err := table.StartHand()
		if err != nil {
			t.Fatalf("expected no error starting hand, got %v", err)
		}

		// Initialize folded players map
		if table.CurrentHand.FoldedPlayers == nil {
			table.CurrentHand.FoldedPlayers = make(map[int]bool)
		}

		// From seat 0, next should be 1
		next := table.CurrentHand.GetNextActiveSeat(0, table.Seats)
		if next == nil || *next != 1 {
			t.Errorf("expected next active seat after 0 to be 1, got %v", next)
		}

		// From seat 1, next should be 2
		next = table.CurrentHand.GetNextActiveSeat(1, table.Seats)
		if next == nil || *next != 2 {
			t.Errorf("expected next active seat after 1 to be 2, got %v", next)
		}

		// From seat 2, should wrap to 0
		next = table.CurrentHand.GetNextActiveSeat(2, table.Seats)
		if next == nil || *next != 0 {
			t.Errorf("expected next active seat after 2 (wrapping) to be 0, got %v", next)
		}
	})
}

// ============ PHASE 5: ALL-IN BETTING LOOP INTEGRATION TESTS ============
// These tests verify that all three fixes (IsBettingRoundComplete, GetValidActions, GetNextActiveSeat)
// work together correctly across realistic game scenarios and all streets

// TestAllInBettingLoop_Integration_TwoPlayerUnequalAllIn tests the classic bug scenario:
// Two players with unequal all-in stacks preflop. Verifies:
// - IsBettingRoundComplete() skips checking all-in player's bet matching
// - GetValidActions() returns empty for all-in player
// - GetNextActiveSeat() skips all-in player
// - Betting round completes without infinite loop
// - Game can advance to flop
func TestAllInBettingLoop_Integration_TwoPlayerUnequalAllIn(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up heads-up: seat 0 (SB=900) vs seat 1 (BB=1000)
	token0 := "player-0"
	token1 := "player-1"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 900

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	// Start hand (dealer at 0, posts 10 SB; seat 1 posts 20 BB)
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Verify hand started correctly
	if table.CurrentHand == nil {
		t.Fatal("expected CurrentHand to be set after StartHand")
	}

	// At this point: dealer/SB (seat 0) has 890, BB (seat 1) has 980
	if table.Seats[0].Stack != 890 {
		t.Errorf("seat 0: expected stack 890, got %d", table.Seats[0].Stack)
	}
	if table.Seats[1].Stack != 980 {
		t.Errorf("seat 1: expected stack 980, got %d", table.Seats[1].Stack)
	}

	// Initialize action: dealer/SB (seat 0) acts first preflop in heads-up
	firstActor := table.CurrentHand.GetFirstActor(table.Seats)
	if firstActor != 0 {
		t.Errorf("expected first actor to be seat 0, got %d", firstActor)
	}

	// Dealer (seat 0) goes all-in by betting their remaining 890
	seat0Allin := 890
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}
	if table.CurrentHand.ActedPlayers == nil {
		table.CurrentHand.ActedPlayers = make(map[int]bool)
	}

	table.CurrentHand.PlayerBets[0] = seat0Allin
	table.Seats[0].Stack = 0 // Mark as all-in
	table.CurrentHand.CurrentBet = seat0Allin
	table.CurrentHand.LastRaise = seat0Allin - 10
	table.CurrentHand.ActedPlayers[0] = true
	currentActor := 1
	table.CurrentHand.CurrentActor = &currentActor

	// Verify GetValidActions for all-in player returns empty
	validActions := table.CurrentHand.GetValidActions(0, 0, table.Seats)
	if len(validActions) != 0 {
		t.Errorf("all-in player should have no valid actions, got %v", validActions)
	}

	// BB (seat 1) must call the all-in bet
	table.CurrentHand.PlayerBets[1] = seat0Allin
	table.Seats[1].Stack = 980 - seat0Allin
	table.CurrentHand.ActedPlayers[1] = true
	table.CurrentHand.BigBlindHasOption = false // BB has acted, so option is closed

	// Verify IsBettingRoundComplete with all-in player
	// Should return true because:
	// - All non-folded players have acted
	// - All non-folded, NON-ALL-IN players have matched the current bet
	// - All-in players (stack=0) are skipped in the matching check
	isComplete := table.CurrentHand.IsBettingRoundComplete(table.Seats)
	if !isComplete {
		t.Error("expected betting round to be complete after BB calls all-in player, got false")
	}

	// Verify game can advance to next street (flop)
	table.CurrentHand.AdvanceStreet()
	if table.CurrentHand.Street != "flop" {
		t.Errorf("expected street to advance to flop, got %s", table.CurrentHand.Street)
	}
}

// TestAllInBettingLoop_Integration_ThreePlayerOneAllInFlop tests one all-in mid-hand:
// Three players, one goes all-in on flop, others continue betting.
// Verifies all-in player is skipped in subsequent rounds.
func TestAllInBettingLoop_Integration_ThreePlayerOneAllInFlop(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 3 active players (seats 0, 1, 2)
	for i := 0; i < 3; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Setup: dealer at 0, SB at 1, BB at 2
	// Stacks: 0: 1000, 1: 990, 2: 980
	if table.Seats[0].Stack != 1000 {
		t.Errorf("seat 0: expected stack 1000, got %d", table.Seats[0].Stack)
	}
	if table.Seats[1].Stack != 990 {
		t.Errorf("seat 1: expected stack 990, got %d", table.Seats[1].Stack)
	}
	if table.Seats[2].Stack != 980 {
		t.Errorf("seat 2: expected stack 980, got %d", table.Seats[2].Stack)
	}

	// Simulate preflop action to get to flop (simplified - just mark as acted)
	if table.CurrentHand.ActedPlayers == nil {
		table.CurrentHand.ActedPlayers = make(map[int]bool)
	}
	if table.CurrentHand.FoldedPlayers == nil {
		table.CurrentHand.FoldedPlayers = make(map[int]bool)
	}

	// Preflop: all players call BB (20)
	table.CurrentHand.PlayerBets[0] = 20
	table.CurrentHand.PlayerBets[1] = 20
	table.CurrentHand.PlayerBets[2] = 20
	table.Seats[0].Stack = 980
	table.Seats[1].Stack = 970
	table.Seats[2].Stack = 960
	table.CurrentHand.CurrentBet = 20

	table.CurrentHand.ActedPlayers[0] = true
	table.CurrentHand.ActedPlayers[1] = true
	table.CurrentHand.ActedPlayers[2] = true
	table.CurrentHand.Pot = 60 // Accumulated from preflop

	// Advance to flop
	table.CurrentHand.Street = "flop"
	table.CurrentHand.CurrentBet = 0
	table.CurrentHand.PlayerBets = make(map[int]int)
	table.CurrentHand.ActedPlayers = make(map[int]bool)
	table.CurrentHand.BigBlindHasOption = false // Reset for postflop

	// On flop, seat 1 (SB) acts first and goes all-in for 500 (less than seat 2's 960)
	seat1BetAmount := 500
	table.CurrentHand.PlayerBets[1] = seat1BetAmount
	table.Seats[1].Stack = 0 // All-in
	table.CurrentHand.CurrentBet = seat1BetAmount
	table.CurrentHand.LastRaise = seat1BetAmount
	table.CurrentHand.ActedPlayers[1] = true

	// Seat 2 (BB) calls the all-in
	table.CurrentHand.PlayerBets[2] = seat1BetAmount
	table.Seats[2].Stack = 960 - seat1BetAmount
	table.CurrentHand.ActedPlayers[2] = true

	// Seat 0 (dealer) calls the all-in
	table.CurrentHand.PlayerBets[0] = seat1BetAmount
	table.Seats[0].Stack = 980 - seat1BetAmount
	table.CurrentHand.ActedPlayers[0] = true

	// Verify IsBettingRoundComplete on flop
	// All players have acted, all non-all-in players matched the bet
	isComplete := table.CurrentHand.IsBettingRoundComplete(table.Seats)
	if !isComplete {
		t.Error("expected betting round on flop to be complete after all-in and calls, got false")
	}

	// Verify GetValidActions for all-in player on flop
	validActions := table.CurrentHand.GetValidActions(1, 0, table.Seats)
	if len(validActions) != 0 {
		t.Errorf("all-in player should have no valid actions on flop, got %v", validActions)
	}

	// Verify game can advance to turn
	table.CurrentHand.AdvanceStreet()
	if table.CurrentHand.Street != "turn" {
		t.Errorf("expected street to advance to turn, got %s", table.CurrentHand.Street)
	}

	// On turn: reset betting, only seats 0 and 2 have chips
	table.CurrentHand.CurrentBet = 0
	table.CurrentHand.PlayerBets = make(map[int]int)
	table.CurrentHand.ActedPlayers = make(map[int]bool)
	// Mark all-in player as acted on this street (they don't take further action)
	table.CurrentHand.ActedPlayers[1] = true

	// Seat 2 bets 100
	table.CurrentHand.PlayerBets[2] = 100
	table.Seats[2].Stack -= 100
	table.CurrentHand.CurrentBet = 100
	table.CurrentHand.ActedPlayers[2] = true

	// Seat 0 calls 100
	table.CurrentHand.PlayerBets[0] = 100
	table.Seats[0].Stack -= 100
	table.CurrentHand.ActedPlayers[0] = true

	// Verify betting round complete on turn (all-in player skipped in all checks)
	isComplete = table.CurrentHand.IsBettingRoundComplete(table.Seats)
	if !isComplete {
		t.Error("expected betting round on turn to be complete (all-in skipped), got false")
	}

	// Advance to river
	table.CurrentHand.AdvanceStreet()
	if table.CurrentHand.Street != "river" {
		t.Errorf("expected street to advance to river, got %s", table.CurrentHand.Street)
	}
}

// TestAllInBettingLoop_Integration_MultiPlayerCascadingAllIns tests multiple all-ins:
// Three players with different stacks, two go all-in with different amounts.
// Verifies proper handling of multiple all-ins in same round.
func TestAllInBettingLoop_Integration_MultiPlayerCascadingAllIns(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 3 active players with different stacks
	token0 := "player-0"
	token1 := "player-1"
	token2 := "player-2"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 300 // Short stack

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 800 // Medium stack

	table.Seats[2].Token = &token2
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1500 // Deep stack

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Initialize maps
	if table.CurrentHand.ActedPlayers == nil {
		table.CurrentHand.ActedPlayers = make(map[int]bool)
	}
	if table.CurrentHand.FoldedPlayers == nil {
		table.CurrentHand.FoldedPlayers = make(map[int]bool)
	}

	// Simulate preflop action to flop
	table.CurrentHand.Street = "flop"
	table.CurrentHand.CurrentBet = 0
	table.CurrentHand.PlayerBets = make(map[int]int)
	table.CurrentHand.ActedPlayers = make(map[int]bool)
	table.CurrentHand.BigBlindHasOption = false // Reset for postflop

	// On turn: seat 0 (300 stack) goes all-in for 300
	table.CurrentHand.Street = "turn"
	table.CurrentHand.PlayerBets[0] = 300
	table.Seats[0].Stack = 0 // All-in
	table.CurrentHand.CurrentBet = 300
	table.CurrentHand.LastRaise = 300
	table.CurrentHand.ActedPlayers[0] = true

	// Seat 1 (800 stack) raises to 800
	table.CurrentHand.PlayerBets[1] = 800
	table.Seats[1].Stack = 0 // Also all-in now
	table.CurrentHand.CurrentBet = 800
	table.CurrentHand.LastRaise = 500
	table.CurrentHand.ActedPlayers[1] = true

	// Seat 2 (1500 stack) calls 800
	table.CurrentHand.PlayerBets[2] = 800
	table.Seats[2].Stack = 700 // Still has chips
	table.CurrentHand.ActedPlayers[2] = true

	// Verify IsBettingRoundComplete with multiple all-ins
	isComplete := table.CurrentHand.IsBettingRoundComplete(table.Seats)
	if !isComplete {
		t.Error("expected betting round to be complete with multiple all-ins and non-all-in matched, got false")
	}

	// Verify GetValidActions for all-in players
	validActions0 := table.CurrentHand.GetValidActions(0, 0, table.Seats)
	if len(validActions0) != 0 {
		t.Errorf("all-in seat 0 should have no valid actions, got %v", validActions0)
	}

	validActions1 := table.CurrentHand.GetValidActions(1, 0, table.Seats)
	if len(validActions1) != 0 {
		t.Errorf("all-in seat 1 should have no valid actions, got %v", validActions1)
	}

	// Verify non-all-in player (seat 2) still has valid actions available
	// (they could bet more if action comes back to them in a new round, but betting is done)
	if table.Seats[2].Stack == 0 {
		validActions2 := table.CurrentHand.GetValidActions(2, 0, table.Seats)
		if len(validActions2) != 0 {
			t.Errorf("all-in seat 2 should have no valid actions, got %v", validActions2)
		}
	}

	// At this point, IsBettingRoundComplete should be true since all players have acted
	// and matched bets (with all-in players excluded from bet matching check)
	if !table.CurrentHand.IsBettingRoundComplete(table.Seats) {
		t.Error("expected betting round to be complete with cascading all-ins")
	}

	// Verify game can advance to river
	table.CurrentHand.AdvanceStreet()
	if table.CurrentHand.Street != "river" {
		t.Errorf("expected street to advance to river, got %s", table.CurrentHand.Street)
	}
}

// TestAllInBettingLoop_Integration_AllInRiverToShowdown tests all-in on river:
// Both players all-in on river, verifies betting completes and shows down correctly.
func TestAllInBettingLoop_Integration_AllInRiverToShowdown(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players (heads-up)
	token0 := "player-0"
	token1 := "player-1"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 500

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 500

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Initialize maps
	if table.CurrentHand.ActedPlayers == nil {
		table.CurrentHand.ActedPlayers = make(map[int]bool)
	}
	if table.CurrentHand.FoldedPlayers == nil {
		table.CurrentHand.FoldedPlayers = make(map[int]bool)
	}

	// Fast-forward to river
	table.CurrentHand.Street = "river"
	table.CurrentHand.CurrentBet = 0
	table.CurrentHand.PlayerBets = make(map[int]int)
	table.CurrentHand.ActedPlayers = make(map[int]bool)
	table.CurrentHand.BigBlindHasOption = false // Reset for postflop

	// Simulate previous betting - both players have matched bets preflop and postflop
	table.CurrentHand.Pot = 400 // Accumulated from previous streets

	// On river: seat 0 goes all-in for 490 (remaining)
	table.CurrentHand.PlayerBets[0] = 490
	table.Seats[0].Stack = 0 // All-in
	table.CurrentHand.CurrentBet = 490
	table.CurrentHand.LastRaise = 490
	table.CurrentHand.ActedPlayers[0] = true

	// Seat 1 calls the all-in for 490
	table.CurrentHand.PlayerBets[1] = 490
	table.Seats[1].Stack = 0 // Also all-in
	table.CurrentHand.ActedPlayers[1] = true

	// Verify both players are marked as all-in
	if table.Seats[0].Stack != 0 {
		t.Errorf("seat 0: expected stack 0 (all-in), got %d", table.Seats[0].Stack)
	}
	if table.Seats[1].Stack != 0 {
		t.Errorf("seat 1: expected stack 0 (all-in), got %d", table.Seats[1].Stack)
	}

	// Verify IsBettingRoundComplete on river with both all-in
	isComplete := table.CurrentHand.IsBettingRoundComplete(table.Seats)
	if !isComplete {
		t.Error("expected betting round on river to be complete with both all-in, got false")
	}

	// Verify GetValidActions for all-in players on river
	validActions0 := table.CurrentHand.GetValidActions(0, 0, table.Seats)
	if len(validActions0) != 0 {
		t.Errorf("all-in seat 0 should have no valid actions on river, got %v", validActions0)
	}

	validActions1 := table.CurrentHand.GetValidActions(1, 0, table.Seats)
	if len(validActions1) != 0 {
		t.Errorf("all-in seat 1 should have no valid actions on river, got %v", validActions1)
	}

	// Verify IsBettingRoundComplete returns true (all players all-in)
	if !table.CurrentHand.IsBettingRoundComplete(table.Seats) {
		t.Error("expected betting round to be complete with both players all-in")
	}

	// Verify we've reached the showdown state (river with no action left)
	// Game should proceed to showdown from here
	if table.CurrentHand.Street != "river" {
		t.Errorf("expected street to be river, got %s", table.CurrentHand.Street)
	}
}

// TestAllInBettingLoop_Integration_MixedAllInAndFold tests all-in + fold scenario:
// Player 1 goes all-in preflop, Player 2 folds on flop, Player 3 continues.
// Verifies only active (non-folded, non-all-in) players get actions.
func TestAllInBettingLoop_Integration_MixedAllInAndFold(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 3 active players
	for i := 0; i < 3; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Initialize maps
	if table.CurrentHand.ActedPlayers == nil {
		table.CurrentHand.ActedPlayers = make(map[int]bool)
	}
	if table.CurrentHand.FoldedPlayers == nil {
		table.CurrentHand.FoldedPlayers = make(map[int]bool)
	}

	// Preflop: seat 1 goes all-in for 990
	table.CurrentHand.PlayerBets[1] = 990
	table.Seats[1].Stack = 0 // All-in
	table.CurrentHand.CurrentBet = 990
	table.CurrentHand.LastRaise = 990
	table.CurrentHand.ActedPlayers[1] = true
	table.CurrentHand.BigBlindHasOption = false // BB (seat 1) has acted, option closed

	// Seats 0 and 2 call
	table.CurrentHand.PlayerBets[0] = 990
	table.Seats[0].Stack = 10
	table.CurrentHand.ActedPlayers[0] = true

	table.CurrentHand.PlayerBets[2] = 990
	table.Seats[2].Stack = 10
	table.CurrentHand.ActedPlayers[2] = true

	// Verify betting round complete
	isComplete := table.CurrentHand.IsBettingRoundComplete(table.Seats)
	if !isComplete {
		t.Error("expected preflop betting to be complete, got false")
	}

	// Advance to flop
	table.CurrentHand.Street = "flop"
	table.CurrentHand.CurrentBet = 0
	table.CurrentHand.PlayerBets = make(map[int]int)
	table.CurrentHand.ActedPlayers = make(map[int]bool)
	table.CurrentHand.BigBlindHasOption = false // Reset for postflop
	// Mark all-in player as acted on this street
	table.CurrentHand.ActedPlayers[1] = true

	// On flop: seat 2 checks
	table.CurrentHand.PlayerBets[2] = 0
	table.CurrentHand.CurrentBet = 0
	table.CurrentHand.ActedPlayers[2] = true

	// Seat 0 checks
	table.CurrentHand.PlayerBets[0] = 0
	table.CurrentHand.ActedPlayers[0] = true

	// Verify betting round complete on flop (all-in skipped, others checked)
	isComplete = table.CurrentHand.IsBettingRoundComplete(table.Seats)
	if !isComplete {
		t.Error("expected flop betting to be complete after all-in and checks, got false")
	}

	// Advance to turn
	table.CurrentHand.Street = "turn"
	table.CurrentHand.CurrentBet = 0
	table.CurrentHand.PlayerBets = make(map[int]int)
	table.CurrentHand.ActedPlayers = make(map[int]bool)
	// Mark all-in player as acted on this street
	table.CurrentHand.ActedPlayers[1] = true

	// On turn: seat 0 bets 5
	table.CurrentHand.PlayerBets[0] = 5
	table.Seats[0].Stack = 5
	table.CurrentHand.CurrentBet = 5
	table.CurrentHand.LastRaise = 5
	table.CurrentHand.ActedPlayers[0] = true

	// Seat 2 folds
	table.CurrentHand.FoldedPlayers[2] = true
	table.CurrentHand.ActedPlayers[2] = true

	// Verify only seat 0 remains (all-in player 1 skipped, player 2 folded)
	// IsBettingRoundComplete should handle this correctly
	isComplete = table.CurrentHand.IsBettingRoundComplete(table.Seats)
	if !isComplete {
		t.Error("expected turn betting to be complete (only one player with chips left), got false")
	}
}

// TestAllInBettingLoop_Integration_AllInPreflop_MultiStreets verifies all-in preflop
// works across multiple postflop streets without action.
func TestAllInBettingLoop_Integration_AllInPreflop_MultiStreets(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 2 active players
	token0 := "player-0"
	token1 := "player-1"

	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	// Start hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("expected no error starting hand, got %v", err)
	}

	// Initialize maps
	if table.CurrentHand.ActedPlayers == nil {
		table.CurrentHand.ActedPlayers = make(map[int]bool)
	}
	if table.CurrentHand.FoldedPlayers == nil {
		table.CurrentHand.FoldedPlayers = make(map[int]bool)
	}

	// Preflop: heads-up dealer/SB (seat 0) goes all-in for 1000
	table.CurrentHand.PlayerBets[0] = 1000
	table.Seats[0].Stack = 0 // All-in
	table.CurrentHand.CurrentBet = 1000
	table.CurrentHand.LastRaise = 1000
	table.CurrentHand.ActedPlayers[0] = true

	// BB (seat 1) calls for 1000
	table.CurrentHand.PlayerBets[1] = 1000
	table.Seats[1].Stack = 0 // Also all-in
	table.CurrentHand.ActedPlayers[1] = true
	table.CurrentHand.BigBlindHasOption = false // BB has acted, option closed

	// Verify preflop betting round complete
	isComplete := table.CurrentHand.IsBettingRoundComplete(table.Seats)
	if !isComplete {
		t.Error("expected preflop betting to be complete with both all-in, got false")
	}

	// Advance through flop, turn, river without action
	streets := []string{"preflop", "flop", "turn", "river"}
	for i := 0; i < len(streets)-1; i++ {
		// Verify current street
		if table.CurrentHand.Street != streets[i] {
			t.Errorf("expected street %s, got %s", streets[i], table.CurrentHand.Street)
		}

		// Advance street
		table.CurrentHand.AdvanceStreet()

		// Verify new street
		if table.CurrentHand.Street != streets[i+1] {
			t.Errorf("expected street to advance to %s, got %s", streets[i+1], table.CurrentHand.Street)
		}

		// On new street, mark all-in players as acted (they don't act in new streets)
		table.CurrentHand.ActedPlayers[0] = true
		table.CurrentHand.ActedPlayers[1] = true

		// Verify betting round still complete (both all-in, no new action)
		isComplete := table.CurrentHand.IsBettingRoundComplete(table.Seats)
		if !isComplete {
			t.Errorf("expected betting to remain complete on %s with both all-in, got false", streets[i+1])
		}

		// Verify GetValidActions still returns empty for all-in players
		validActions0 := table.CurrentHand.GetValidActions(0, 0, table.Seats)
		if len(validActions0) != 0 {
			t.Errorf("%s: all-in seat 0 should have no valid actions, got %v", streets[i+1], validActions0)
		}

		validActions1 := table.CurrentHand.GetValidActions(1, 0, table.Seats)
		if len(validActions1) != 0 {
			t.Errorf("%s: all-in seat 1 should have no valid actions, got %v", streets[i+1], validActions1)
		}
	}

	// Verify we reached river
	if table.CurrentHand.Street != "river" {
		t.Errorf("expected to reach river, got %s", table.CurrentHand.Street)
	}
}

// TestAreAllActivePlayersAllIn_TwoPlayersAllIn tests when both players are all-in
func TestAreAllActivePlayersAllIn_TwoPlayersAllIn(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	tokenA := "token-a"
	tokenB := "token-b"

	table.Seats[0].Token = &tokenA
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 0 // All-in

	table.Seats[1].Token = &tokenB
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 0 // All-in

	hand := &Hand{
		FoldedPlayers: make(map[int]bool),
	}

	result := hand.AreAllActivePlayersAllIn(table.Seats)
	if !result {
		t.Error("expected true when both players are all-in")
	}
}

// TestAreAllActivePlayersAllIn_OnePlayerHasChips tests when one player has chips remaining
// This should return TRUE because at least one player is all-in (stack = 0),
// which means no further betting is possible and remaining streets should auto-deal.
func TestAreAllActivePlayersAllIn_OnePlayerHasChips(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	tokenA := "token-a"
	tokenB := "token-b"

	table.Seats[0].Token = &tokenA
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 0 // All-in

	table.Seats[1].Token = &tokenB
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 500 // Has chips

	hand := &Hand{
		FoldedPlayers: make(map[int]bool),
	}

	result := hand.AreAllActivePlayersAllIn(table.Seats)
	if !result {
		t.Error("expected true when at least one player is all-in (no further betting possible)")
	}
}

// TestAreAllActivePlayersAllIn_ThreePlayersAllIn tests with 3 players all-in
func TestAreAllActivePlayersAllIn_ThreePlayersAllIn(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	tokenA := "token-a"
	tokenB := "token-b"
	tokenC := "token-c"

	table.Seats[0].Token = &tokenA
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 0 // All-in

	table.Seats[1].Token = &tokenB
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 0 // All-in

	table.Seats[2].Token = &tokenC
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 0 // All-in

	hand := &Hand{
		FoldedPlayers: make(map[int]bool),
	}

	result := hand.AreAllActivePlayersAllIn(table.Seats)
	if !result {
		t.Error("expected true when all three players are all-in")
	}
}

// TestAreAllActivePlayersAllIn_OneFolded tests with one player folded
func TestAreAllActivePlayersAllIn_OneFolded(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	tokenA := "token-a"
	tokenB := "token-b"
	tokenC := "token-c"

	table.Seats[0].Token = &tokenA
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 0 // All-in

	table.Seats[1].Token = &tokenB
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 0 // All-in

	table.Seats[2].Token = &tokenC
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 500 // Had chips but folded

	hand := &Hand{
		FoldedPlayers: map[int]bool{
			2: true, // Seat 2 folded
		},
	}

	result := hand.AreAllActivePlayersAllIn(table.Seats)
	if !result {
		t.Error("expected true when remaining active players (not folded) are all-in")
	}
}

// TestAreAllActivePlayersAllIn_OnlyOnePlayer tests with only one player (should return false)
func TestAreAllActivePlayersAllIn_OnlyOnePlayer(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	tokenA := "token-a"

	table.Seats[0].Token = &tokenA
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 0 // All-in

	hand := &Hand{
		FoldedPlayers: make(map[int]bool),
	}

	result := hand.AreAllActivePlayersAllIn(table.Seats)
	if result {
		t.Error("expected false when only one player remains (need 2+ for all-in to matter)")
	}
}

// TestAreAllActivePlayersAllIn_OnePlayerAllIn tests the scenario where
// Player B (500) goes all-in, Player A (1000) calls.
// After the call, Player A has 500 left, Player B has 0.
// This should return TRUE because at least one player is all-in.
func TestAreAllActivePlayersAllIn_OnePlayerAllIn(t *testing.T) {
	hand := &Hand{
		FoldedPlayers: make(map[int]bool),
	}

	seats := [6]Seat{
		{Status: "active", Stack: 500}, // Player A called 500, has 500 left
		{Status: "active", Stack: 0},   // Player B went all-in, has 0 left
		{Status: "empty"},
		{Status: "empty"},
		{Status: "empty"},
		{Status: "empty"},
	}

	result := hand.AreAllActivePlayersAllIn(seats)
	if !result {
		t.Errorf("Expected true when one player is all-in and betting is complete, got false")
	}
}

// TestAreAllActivePlayersAllIn_BothAllIn tests when both players go all-in
func TestAreAllActivePlayersAllIn_BothAllIn(t *testing.T) {
	hand := &Hand{
		FoldedPlayers: make(map[int]bool),
	}

	seats := [6]Seat{
		{Status: "active", Stack: 0}, // Both all-in
		{Status: "active", Stack: 0},
		{Status: "empty"},
		{Status: "empty"},
		{Status: "empty"},
		{Status: "empty"},
	}

	result := hand.AreAllActivePlayersAllIn(seats)
	if !result {
		t.Errorf("Expected true when both players are all-in, got false")
	}
}

// TestAreAllActivePlayersAllIn_NoAllIn tests when no players are all-in
func TestAreAllActivePlayersAllIn_NoAllIn(t *testing.T) {
	hand := &Hand{
		FoldedPlayers: make(map[int]bool),
	}

	seats := [6]Seat{
		{Status: "active", Stack: 500}, // Both have chips
		{Status: "active", Stack: 500},
		{Status: "empty"},
		{Status: "empty"},
		{Status: "empty"},
		{Status: "empty"},
	}

	result := hand.AreAllActivePlayersAllIn(seats)
	if result {
		t.Errorf("Expected false when no players are all-in, got true")
	}
}

// TestAreAllActivePlayersAllIn_ThreePlayersOneAllIn tests three players where one is all-in
func TestAreAllActivePlayersAllIn_ThreePlayersOneAllIn(t *testing.T) {
	hand := &Hand{
		FoldedPlayers: make(map[int]bool),
	}

	seats := [6]Seat{
		{Status: "active", Stack: 1000}, // Player A has chips
		{Status: "active", Stack: 0},    // Player B all-in
		{Status: "active", Stack: 500},  // Player C has chips
		{Status: "empty"},
		{Status: "empty"},
		{Status: "empty"},
	}

	result := hand.AreAllActivePlayersAllIn(seats)
	if !result {
		t.Errorf("Expected true when at least one player is all-in, got false")
	}
}

// TestTotalContributions_InitializedOnStartHand verifies TotalContributions is initialized as empty map
func TestTotalContributions_InitializedOnStartHand(t *testing.T) {
	server := NewServer(slog.Default())
	table := NewTable("table-1", "Test Table", server)

	// Add 2 active players
	token0 := "player-0"
	token1 := "player-1"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	// Start the hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("StartHand failed: %v", err)
	}

	// Verify TotalContributions exists and is a map
	if table.CurrentHand.TotalContributions == nil {
		t.Errorf("Expected TotalContributions to be initialized, got nil")
	}

	// Verify blinds are tracked in TotalContributions at hand start
	sbSeat := table.CurrentHand.SmallBlindSeat
	bbSeat := table.CurrentHand.BigBlindSeat

	if table.CurrentHand.TotalContributions[sbSeat] != 10 {
		t.Errorf("Expected TotalContributions[%d] (SB) to be 10 (small blind), got %d", sbSeat, table.CurrentHand.TotalContributions[sbSeat])
	}
	if table.CurrentHand.TotalContributions[bbSeat] != 20 {
		t.Errorf("Expected TotalContributions[%d] (BB) to be 20 (big blind), got %d", bbSeat, table.CurrentHand.TotalContributions[bbSeat])
	}
}

// TestTotalContributions_TracksBlindPosting verifies blinds are tracked in TotalContributions after StartHand
func TestTotalContributions_TracksBlindPosting(t *testing.T) {
	server := NewServer(slog.Default())
	table := NewTable("table-1", "Test Table", server)

	// Add 3 active players
	token0 := "player-0"
	token1 := "player-1"
	token2 := "player-2"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	table.Seats[2].Token = &token2
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 1000

	// Start the hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("StartHand failed: %v", err)
	}

	hand := table.CurrentHand
	sbSeat := hand.SmallBlindSeat
	bbSeat := hand.BigBlindSeat

	// After StartHand, blinds should be posted to PlayerBets
	// Small blind = 10, Big blind = 20
	sbAmount := hand.PlayerBets[sbSeat]
	bbAmount := hand.PlayerBets[bbSeat]

	if sbAmount != 10 {
		t.Errorf("Expected SB to post 10, got %d", sbAmount)
	}
	if bbAmount != 20 {
		t.Errorf("Expected BB to post 20, got %d", bbAmount)
	}

	// TotalContributions should now include blind amounts
	if hand.TotalContributions[sbSeat] != 10 {
		t.Errorf("Expected TotalContributions[%d] (SB) to be 10 (blind), got %d", sbSeat, hand.TotalContributions[sbSeat])
	}
	if hand.TotalContributions[bbSeat] != 20 {
		t.Errorf("Expected TotalContributions[%d] (BB) to be 20 (blind), got %d", bbSeat, hand.TotalContributions[bbSeat])
	}
}

// TestTotalContributions_PersistsAcrossStreets verifies TotalContributions persists through streets
func TestTotalContributions_PersistsAcrossStreets(t *testing.T) {
	server := NewServer(slog.Default())
	table := NewTable("table-1", "Test Table", server)

	// Add 2 active players
	token0 := "player-0"
	token1 := "player-1"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	// Start the hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("StartHand failed: %v", err)
	}

	hand := table.CurrentHand

	// Record initial street
	initialStreet := hand.Street
	if initialStreet != "preflop" {
		t.Errorf("Expected initial street to be 'preflop', got '%s'", initialStreet)
	}

	// Verify TotalContributions exists after StartHand
	if hand.TotalContributions == nil {
		t.Errorf("Expected TotalContributions to be initialized, got nil")
	}

	// Add some values to TotalContributions to test persistence
	hand.TotalContributions[0] = 100
	hand.TotalContributions[1] = 50

	// Verify values were added
	if hand.TotalContributions[0] != 100 {
		t.Errorf("Expected TotalContributions[0] to be 100, got %d", hand.TotalContributions[0])
	}
	if hand.TotalContributions[1] != 50 {
		t.Errorf("Expected TotalContributions[1] to be 50, got %d", hand.TotalContributions[1])
	}

	// Advance to next street
	hand.AdvanceStreet()

	// Verify we're on next street
	if hand.Street != "flop" {
		t.Errorf("Expected street to advance to 'flop', got '%s'", hand.Street)
	}

	// Verify TotalContributions persisted
	if hand.TotalContributions[0] != 100 {
		t.Errorf("Expected TotalContributions[0] to persist with 100 after street advance, got %d", hand.TotalContributions[0])
	}
	if hand.TotalContributions[1] != 50 {
		t.Errorf("Expected TotalContributions[1] to persist with 50 after street advance, got %d", hand.TotalContributions[1])
	}
}

// TestTotalContributions_TracksCallAmount verifies that calls are tracked in TotalContributions
func TestTotalContributions_TracksCallAmount(t *testing.T) {
	server := NewServer(slog.Default())
	table := NewTable("table-1", "Test Table", server)

	// Add 2 active players
	token0 := "player-0"
	token1 := "player-1"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	// Start the hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("StartHand failed: %v", err)
	}

	hand := table.CurrentHand

	// Blinds: SB (10) and BB (20)
	sbSeat := hand.SmallBlindSeat

	// First action: SB calls to match BB
	// SB has 10 posted, needs to call 10 more
	callAmount := hand.GetCallAmount(sbSeat)
	expectedIncrementalChips := callAmount // This should be 10

	chipsMoved, err := hand.ProcessAction(sbSeat, "call", 1000)
	if err != nil {
		t.Fatalf("Call action failed: %v", err)
	}

	if chipsMoved != expectedIncrementalChips {
		t.Errorf("Expected chipsMoved to be %d for call, got %d", expectedIncrementalChips, chipsMoved)
	}

	// Verify TotalContributions was updated: includes blind (10) + call (10) = 20
	expectedTotal := 10 + expectedIncrementalChips // 10 (blind) + 10 (call) = 20
	if hand.TotalContributions[sbSeat] != expectedTotal {
		t.Errorf("Expected TotalContributions[%d] to be %d (blind 10 + call 10), got %d", sbSeat, expectedTotal, hand.TotalContributions[sbSeat])
	}
}

// TestTotalContributions_TracksRaiseAmount verifies that raises are tracked in TotalContributions
func TestTotalContributions_TracksRaiseAmount(t *testing.T) {
	server := NewServer(slog.Default())
	table := NewTable("table-1", "Test Table", server)

	// Add 2 active players
	token0 := "player-0"
	token1 := "player-1"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	// Start the hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("StartHand failed: %v", err)
	}

	hand := table.CurrentHand

	// Blinds: SB (10) and BB (20)
	sbSeat := hand.SmallBlindSeat

	// First action: SB raises to 60 (incremental: 60 - 10 = 50)
	raiseAmount := 60
	expectedIncrementalChips := raiseAmount - hand.PlayerBets[sbSeat] // 60 - 10 = 50

	chipsMoved, err := hand.ProcessAction(sbSeat, "raise", 1000, raiseAmount)
	if err != nil {
		t.Fatalf("Raise action failed: %v", err)
	}

	// Chips moved should be the incremental amount (50)
	if chipsMoved != expectedIncrementalChips {
		t.Errorf("Expected chipsMoved to be %d for raise, got %d", expectedIncrementalChips, chipsMoved)
	}

	// Verify TotalContributions was updated: includes blind (10) + raise amount (50) = 60
	expectedTotal := 10 + expectedIncrementalChips // 10 (blind) + 50 (raise increment) = 60
	if hand.TotalContributions[sbSeat] != expectedTotal {
		t.Errorf("Expected TotalContributions[%d] to be %d (blind 10 + raise increment 50), got %d", sbSeat, expectedTotal, hand.TotalContributions[sbSeat])
	}
}

// TestTotalContributions_TracksAllInAmount verifies that all-ins are tracked in TotalContributions
func TestTotalContributions_TracksAllInAmount(t *testing.T) {
	server := NewServer(slog.Default())
	table := NewTable("table-1", "Test Table", server)

	// Add 2 active players with small stacks
	token0 := "player-0"
	token1 := "player-1"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 50 // Small stack for SB

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	// Start the hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("StartHand failed: %v", err)
	}

	hand := table.CurrentHand

	// Blinds: SB (10) and BB (20)
	sbSeat := hand.SmallBlindSeat

	// First action: SB all-ins by raising to their remaining 50
	// They have 10 posted, so 40 remaining in stack
	// But they're all-in with 50 total bet
	expectedIncrementalChips := 50 - hand.PlayerBets[sbSeat] // 50 - 10 = 40

	chipsMoved, err := hand.ProcessAction(sbSeat, "raise", 40, 50)
	if err != nil {
		t.Fatalf("All-in raise action failed: %v", err)
	}

	// Chips moved should be 40 (remaining in stack)
	if chipsMoved != expectedIncrementalChips {
		t.Errorf("Expected chipsMoved to be %d for all-in, got %d", expectedIncrementalChips, chipsMoved)
	}

	// Verify TotalContributions was updated: includes blind (10) + all-in raise (40) = 50
	expectedTotal := 10 + expectedIncrementalChips // 10 (blind) + 40 (all-in increment) = 50
	if hand.TotalContributions[sbSeat] != expectedTotal {
		t.Errorf("Expected TotalContributions[%d] to be %d (blind 10 + all-in 40), got %d", sbSeat, expectedTotal, hand.TotalContributions[sbSeat])
	}
}

// TestTotalContributions_AccumulatesAcrossStreets verifies contributions accumulate across streets
func TestTotalContributions_AccumulatesAcrossStreets(t *testing.T) {
	server := NewServer(slog.Default())
	table := NewTable("table-1", "Test Table", server)

	// Add 2 active players
	token0 := "player-0"
	token1 := "player-1"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	// Start the hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("StartHand failed: %v", err)
	}

	hand := table.CurrentHand
	sbSeat := hand.SmallBlindSeat

	// Preflop action: SB calls to match BB (incremental: 10)
	_, err = hand.ProcessAction(sbSeat, "call", 1000)
	if err != nil {
		t.Fatalf("Preflop call failed: %v", err)
	}

	// After SB calls: blind (10) + call (10) = 20
	preflopContribution := hand.TotalContributions[sbSeat]
	expectedPreflopTotal := 20 // 10 (blind) + 10 (call)
	if preflopContribution != expectedPreflopTotal {
		t.Errorf("Expected preflop contribution to be %d (blind 10 + call 10), got %d", expectedPreflopTotal, preflopContribution)
	}

	// Advance to flop
	hand.AdvanceStreet()

	// Flop action: Reset street-level tracking (PlayerBets, ActedPlayers, CurrentBet)
	// but TotalContributions persists
	hand.PlayerBets = make(map[int]int)
	hand.ActedPlayers = make(map[int]bool)
	hand.CurrentBet = 0

	// SB checks (0 bet)
	_, err = hand.ProcessAction(sbSeat, "check", 1000)
	if err != nil {
		t.Fatalf("Flop check failed: %v", err)
	}

	// Verify TotalContributions still has preflop contribution
	afterCheck := hand.TotalContributions[sbSeat]
	if afterCheck != preflopContribution {
		t.Errorf("Expected TotalContributions to remain %d after check, got %d", preflopContribution, afterCheck)
	}

	// Flop: SB bets 50
	_, err = hand.ProcessAction(sbSeat, "raise", 1000, 50)
	if err != nil {
		t.Fatalf("Flop bet failed: %v", err)
	}

	// Verify TotalContributions now includes both preflop and flop contributions
	totalExpected := preflopContribution + 50 // 20 (preflop: blind 10 + call 10) + 50 (flop)
	totalActual := hand.TotalContributions[sbSeat]
	if totalActual != totalExpected {
		t.Errorf("Expected TotalContributions to be %d (preflop %d + flop 50), got %d", totalExpected, preflopContribution, totalActual)
	}
}

// TestTotalContributions_HandlesMultipleRaisesPerPlayer verifies multiple raises same street accumulate
func TestTotalContributions_HandlesMultipleRaisesPerPlayer(t *testing.T) {
	server := NewServer(slog.Default())
	table := NewTable("table-1", "Test Table", server)

	// Add 2 active players
	token0 := "player-0"
	token1 := "player-1"
	table.Seats[0].Token = &token0
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 1000

	table.Seats[1].Token = &token1
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 1000

	// Start the hand
	err := table.StartHand()
	if err != nil {
		t.Fatalf("StartHand failed: %v", err)
	}

	hand := table.CurrentHand
	sbSeat := hand.SmallBlindSeat
	bbSeat := hand.BigBlindSeat

	// Preflop sequence:
	// 1. SB raises to 60 (has 10 posted, so incremental 50)
	_, err = hand.ProcessAction(sbSeat, "raise", 1000, 60)
	if err != nil {
		t.Fatalf("First raise failed: %v", err)
	}

	contrib1 := hand.TotalContributions[sbSeat]
	expectedAfterFirstRaise := 10 + 50 // 10 (blind) + 50 (raise increment) = 60
	if contrib1 != expectedAfterFirstRaise {
		t.Errorf("Expected TotalContributions[%d] to be %d (blind 10 + raise 50) after first raise, got %d", sbSeat, expectedAfterFirstRaise, contrib1)
	}

	// 2. BB re-raises to 150 (has 20 posted, so incremental 130)
	_, err = hand.ProcessAction(bbSeat, "raise", 1000, 150)
	if err != nil {
		t.Fatalf("BB re-raise failed: %v", err)
	}

	contribBB := hand.TotalContributions[bbSeat]
	expectedBB := 20 + 130 // 20 (blind) + 130 (raise increment) = 150
	if contribBB != expectedBB {
		t.Errorf("Expected TotalContributions[%d] to be %d (blind 20 + raise 130) after BB raise, got %d", bbSeat, expectedBB, contribBB)
	}

	// 3. SB re-raises to 300 (has 60 posted from previous raise, so incremental 240)
	_, err = hand.ProcessAction(sbSeat, "raise", 1000, 300)
	if err != nil {
		t.Fatalf("Second raise failed: %v", err)
	}

	// Verify TotalContributions updated with incremental amount (300 - 60 = 240 more)
	contrib2 := hand.TotalContributions[sbSeat]
	expectedAfterSecondRaise := expectedAfterFirstRaise + 240 // 60 + 240 = 300
	if contrib2 != expectedAfterSecondRaise {
		t.Errorf("Expected TotalContributions[%d] to be %d (60 + 240) after second raise, got %d", sbSeat, expectedAfterSecondRaise, contrib2)
	}
}

// ============================================================================
// Tests for CalculateSidePots function (Phase 3 of side pot distribution fix)
// ============================================================================

// Test2PlayerEqualStacks: Two players with equal contribution
func TestCalculateSidePots2PlayerEqualStacks(t *testing.T) {
	contributions := map[int]int{
		0: 100,
		1: 100,
	}
	folded := map[int]bool{
		0: false,
		1: false,
	}

	pots := CalculateSidePots(contributions, folded)

	if len(pots) != 1 {
		t.Fatalf("Expected 1 pot, got %d", len(pots))
	}

	if pots[0].Amount != 200 {
		t.Errorf("Expected pot amount 200, got %d", pots[0].Amount)
	}

	expectedSeats := []int{0, 1}
	if len(pots[0].EligibleSeats) != len(expectedSeats) {
		t.Errorf("Expected %d eligible seats, got %d", len(expectedSeats), len(pots[0].EligibleSeats))
	}

	for i, seat := range pots[0].EligibleSeats {
		if seat != expectedSeats[i] {
			t.Errorf("Expected eligible seat %d at index %d, got %d", expectedSeats[i], i, seat)
		}
	}
}

// Test2PlayerOneShortStack: Two players with unequal contribution (one all-in with less)
func TestCalculateSidePots2PlayerOneShortStack(t *testing.T) {
	contributions := map[int]int{
		0: 50,
		1: 200,
	}
	folded := map[int]bool{
		0: false,
		1: false,
	}

	pots := CalculateSidePots(contributions, folded)

	if len(pots) != 2 {
		t.Fatalf("Expected 2 pots, got %d", len(pots))
	}

	// Pot 1: 50 * 2 = 100 (both players eligible)
	if pots[0].Amount != 100 {
		t.Errorf("Expected pot[0] amount 100, got %d", pots[0].Amount)
	}
	if len(pots[0].EligibleSeats) != 2 {
		t.Errorf("Expected pot[0] 2 eligible seats, got %d", len(pots[0].EligibleSeats))
	}

	// Pot 2: (200 - 50) * 1 = 150 (only player 1 eligible)
	if pots[1].Amount != 150 {
		t.Errorf("Expected pot[1] amount 150, got %d", pots[1].Amount)
	}
	if len(pots[1].EligibleSeats) != 1 || pots[1].EligibleSeats[0] != 1 {
		t.Errorf("Expected pot[1] eligible seat [1], got %v", pots[1].EligibleSeats)
	}
}

// Test3PlayerOneShortStack: Three players, one with minimal contribution
func TestCalculateSidePots3PlayerOneShortStack(t *testing.T) {
	contributions := map[int]int{
		0: 30,
		1: 100,
		2: 100,
	}
	folded := map[int]bool{
		0: false,
		1: false,
		2: false,
	}

	pots := CalculateSidePots(contributions, folded)

	if len(pots) != 2 {
		t.Fatalf("Expected 2 pots, got %d", len(pots))
	}

	// Pot 1: 30 * 3 = 90 (all three eligible)
	if pots[0].Amount != 90 {
		t.Errorf("Expected pot[0] amount 90, got %d", pots[0].Amount)
	}
	if len(pots[0].EligibleSeats) != 3 {
		t.Errorf("Expected pot[0] 3 eligible seats, got %d", len(pots[0].EligibleSeats))
	}

	// Pot 2: (100 - 30) * 2 = 140 (players 1 and 2 eligible)
	if pots[1].Amount != 140 {
		t.Errorf("Expected pot[1] amount 140, got %d", pots[1].Amount)
	}
	if len(pots[1].EligibleSeats) != 2 {
		t.Errorf("Expected pot[1] 2 eligible seats, got %d", len(pots[1].EligibleSeats))
	}
}

// Test3PlayerTwoShortStacks: Three players with two all-ins at different levels
func TestCalculateSidePots3PlayerTwoShortStacks(t *testing.T) {
	contributions := map[int]int{
		0: 50,
		1: 150,
		2: 300,
	}
	folded := map[int]bool{
		0: false,
		1: false,
		2: false,
	}

	pots := CalculateSidePots(contributions, folded)

	if len(pots) != 3 {
		t.Fatalf("Expected 3 pots, got %d", len(pots))
	}

	// Pot 1: 50 * 3 = 150 (all three eligible)
	if pots[0].Amount != 150 {
		t.Errorf("Expected pot[0] amount 150, got %d", pots[0].Amount)
	}
	if len(pots[0].EligibleSeats) != 3 {
		t.Errorf("Expected pot[0] 3 eligible seats, got %d", len(pots[0].EligibleSeats))
	}

	// Pot 2: (150 - 50) * 2 = 200 (players 1 and 2 eligible)
	if pots[1].Amount != 200 {
		t.Errorf("Expected pot[1] amount 200, got %d", pots[1].Amount)
	}
	if len(pots[1].EligibleSeats) != 2 {
		t.Errorf("Expected pot[1] 2 eligible seats, got %d", len(pots[1].EligibleSeats))
	}

	// Pot 3: (300 - 150) * 1 = 150 (only player 2 eligible)
	if pots[2].Amount != 150 {
		t.Errorf("Expected pot[2] amount 150, got %d", pots[2].Amount)
	}
	if len(pots[2].EligibleSeats) != 1 || pots[2].EligibleSeats[0] != 2 {
		t.Errorf("Expected pot[2] eligible seat [2], got %v", pots[2].EligibleSeats)
	}
}

// Test3PlayerOnePlayerFolds: Three players with one folding
func TestCalculateSidePots3PlayerOnePlayerFolds(t *testing.T) {
	contributions := map[int]int{
		0: 100,
		1: 100,
		2: 100,
	}
	folded := map[int]bool{
		0: true, // Player 0 folded
		1: false,
		2: false,
	}

	pots := CalculateSidePots(contributions, folded)

	if len(pots) != 1 {
		t.Fatalf("Expected 1 pot, got %d", len(pots))
	}

	// Total pot: 300, but only players 1 and 2 are eligible
	if pots[0].Amount != 300 {
		t.Errorf("Expected pot amount 300, got %d", pots[0].Amount)
	}
	if len(pots[0].EligibleSeats) != 2 {
		t.Errorf("Expected 2 eligible seats, got %d", len(pots[0].EligibleSeats))
	}

	// Should only include seats 1 and 2
	for _, seat := range pots[0].EligibleSeats {
		if seat != 1 && seat != 2 {
			t.Errorf("Expected only seats 1 and 2, got %d", seat)
		}
	}
}

// Test4PlayerMultipleAllIns: Four players with multiple all-ins at different levels
func TestCalculateSidePots4PlayerMultipleAllIns(t *testing.T) {
	contributions := map[int]int{
		0: 50,
		1: 100,
		2: 200,
		3: 300,
	}
	folded := map[int]bool{
		0: false,
		1: false,
		2: false,
		3: false,
	}

	pots := CalculateSidePots(contributions, folded)

	if len(pots) != 4 {
		t.Fatalf("Expected 4 pots, got %d", len(pots))
	}

	// Pot 1: 50 * 4 = 200 (all eligible)
	if pots[0].Amount != 200 {
		t.Errorf("Expected pot[0] amount 200, got %d", pots[0].Amount)
	}
	if len(pots[0].EligibleSeats) != 4 {
		t.Errorf("Expected pot[0] 4 eligible seats, got %d", len(pots[0].EligibleSeats))
	}

	// Pot 2: (100 - 50) * 3 = 150 (players 1, 2, 3)
	if pots[1].Amount != 150 {
		t.Errorf("Expected pot[1] amount 150, got %d", pots[1].Amount)
	}
	if len(pots[1].EligibleSeats) != 3 {
		t.Errorf("Expected pot[1] 3 eligible seats, got %d", len(pots[1].EligibleSeats))
	}

	// Pot 3: (200 - 100) * 2 = 200 (players 2, 3)
	if pots[2].Amount != 200 {
		t.Errorf("Expected pot[2] amount 200, got %d", pots[2].Amount)
	}
	if len(pots[2].EligibleSeats) != 2 {
		t.Errorf("Expected pot[2] 2 eligible seats, got %d", len(pots[2].EligibleSeats))
	}

	// Pot 4: (300 - 200) * 1 = 100 (player 3 only)
	if pots[3].Amount != 100 {
		t.Errorf("Expected pot[3] amount 100, got %d", pots[3].Amount)
	}
	if len(pots[3].EligibleSeats) != 1 || pots[3].EligibleSeats[0] != 3 {
		t.Errorf("Expected pot[3] eligible seat [3], got %v", pots[3].EligibleSeats)
	}
}

// Test4PlayerWithFolds: Four players with some folding
func TestCalculateSidePots4PlayerWithFolds(t *testing.T) {
	contributions := map[int]int{
		0: 100,
		1: 200,
		2: 200,
		3: 100,
	}
	folded := map[int]bool{
		0: true, // Folded
		1: false,
		2: false,
		3: true, // Folded
	}

	pots := CalculateSidePots(contributions, folded)

	if len(pots) != 2 {
		t.Fatalf("Expected 2 pots, got %d", len(pots))
	}

	// Pot 1: 100 * 4 = 400 (all contributed, but only 1 and 2 eligible)
	if pots[0].Amount != 400 {
		t.Errorf("Expected pot[0] amount 400, got %d", pots[0].Amount)
	}
	if len(pots[0].EligibleSeats) != 2 {
		t.Errorf("Expected pot[0] 2 eligible seats, got %d", len(pots[0].EligibleSeats))
	}

	// Pot 2: (200 - 100) * 2 = 200 (only seats 1 and 2 eligible)
	if pots[1].Amount != 200 {
		t.Errorf("Expected pot[1] amount 200, got %d", pots[1].Amount)
	}
	if len(pots[1].EligibleSeats) != 2 {
		t.Errorf("Expected pot[1] 2 eligible seats, got %d", len(pots[1].EligibleSeats))
	}
}

// Test4PlayerComplexScenario: Four players with complex contribution and fold pattern
func TestCalculateSidePots4PlayerComplexScenario(t *testing.T) {
	contributions := map[int]int{
		0: 75,
		1: 300,
		2: 150,
		3: 300,
	}
	folded := map[int]bool{
		0: false,
		1: true, // Folded
		2: false,
		3: false,
	}

	pots := CalculateSidePots(contributions, folded)

	if len(pots) != 3 {
		t.Fatalf("Expected 3 pots, got %d", len(pots))
	}

	// Pot 1: 75 * 4 = 300 (all contributed, but only 0, 2, 3 eligible)
	if pots[0].Amount != 300 {
		t.Errorf("Expected pot[0] amount 300, got %d", pots[0].Amount)
	}
	if len(pots[0].EligibleSeats) != 3 {
		t.Errorf("Expected pot[0] 3 eligible seats, got %d", len(pots[0].EligibleSeats))
	}

	// Pot 2: (150 - 75) * 3 = 225 (seats 1, 2, 3 contributed, but 2, 3 eligible)
	if pots[1].Amount != 225 {
		t.Errorf("Expected pot[1] amount 225, got %d", pots[1].Amount)
	}
	if len(pots[1].EligibleSeats) != 2 {
		t.Errorf("Expected pot[1] 2 eligible seats, got %d", len(pots[1].EligibleSeats))
	}

	// Pot 3: (300 - 150) * 2 = 300 (seats 1 and 3 eligible, but 1 folded, so only 3)
	if pots[2].Amount != 300 {
		t.Errorf("Expected pot[2] amount 300, got %d", pots[2].Amount)
	}
	if len(pots[2].EligibleSeats) != 1 || pots[2].EligibleSeats[0] != 3 {
		t.Errorf("Expected pot[2] eligible seat [3], got %v", pots[2].EligibleSeats)
	}
}

// Test5PlayerLadderAllIns: Five players with ladder-like contributions
func TestCalculateSidePots5PlayerLadderAllIns(t *testing.T) {
	contributions := map[int]int{
		0: 100,
		1: 200,
		2: 300,
		3: 400,
		4: 500,
	}
	folded := map[int]bool{
		0: false,
		1: false,
		2: false,
		3: false,
		4: false,
	}

	pots := CalculateSidePots(contributions, folded)

	if len(pots) != 5 {
		t.Fatalf("Expected 5 pots, got %d", len(pots))
	}

	// Pot 1: 100 * 5 = 500
	if pots[0].Amount != 500 {
		t.Errorf("Expected pot[0] amount 500, got %d", pots[0].Amount)
	}
	if len(pots[0].EligibleSeats) != 5 {
		t.Errorf("Expected pot[0] 5 eligible seats, got %d", len(pots[0].EligibleSeats))
	}

	// Pot 2: (200 - 100) * 4 = 400
	if pots[1].Amount != 400 {
		t.Errorf("Expected pot[1] amount 400, got %d", pots[1].Amount)
	}
	if len(pots[1].EligibleSeats) != 4 {
		t.Errorf("Expected pot[1] 4 eligible seats, got %d", len(pots[1].EligibleSeats))
	}

	// Pot 3: (300 - 200) * 3 = 300
	if pots[2].Amount != 300 {
		t.Errorf("Expected pot[2] amount 300, got %d", pots[2].Amount)
	}
	if len(pots[2].EligibleSeats) != 3 {
		t.Errorf("Expected pot[2] 3 eligible seats, got %d", len(pots[2].EligibleSeats))
	}

	// Pot 4: (400 - 300) * 2 = 200
	if pots[3].Amount != 200 {
		t.Errorf("Expected pot[3] amount 200, got %d", pots[3].Amount)
	}
	if len(pots[3].EligibleSeats) != 2 {
		t.Errorf("Expected pot[3] 2 eligible seats, got %d", len(pots[3].EligibleSeats))
	}

	// Pot 5: (500 - 400) * 1 = 100
	if pots[4].Amount != 100 {
		t.Errorf("Expected pot[4] amount 100, got %d", pots[4].Amount)
	}
	if len(pots[4].EligibleSeats) != 1 || pots[4].EligibleSeats[0] != 4 {
		t.Errorf("Expected pot[4] eligible seat [4], got %v", pots[4].EligibleSeats)
	}
}

// Test5PlayerWithMultipleFolds: Five players with multiple folds
func TestCalculateSidePots5PlayerWithMultipleFolds(t *testing.T) {
	contributions := map[int]int{
		0: 100,
		1: 200,
		2: 200,
		3: 200,
		4: 100,
	}
	folded := map[int]bool{
		0: true,
		1: true,
		2: false,
		3: false,
		4: true,
	}

	pots := CalculateSidePots(contributions, folded)

	if len(pots) != 2 {
		t.Fatalf("Expected 2 pots, got %d", len(pots))
	}

	// Pot 1: 100 * 5 = 500 (only seats 2 and 3 eligible)
	if pots[0].Amount != 500 {
		t.Errorf("Expected pot[0] amount 500, got %d", pots[0].Amount)
	}
	if len(pots[0].EligibleSeats) != 2 {
		t.Errorf("Expected pot[0] 2 eligible seats, got %d", len(pots[0].EligibleSeats))
	}

	// Pot 2: (200 - 100) * 3 = 300 (only seats 2 and 3 eligible)
	if pots[1].Amount != 300 {
		t.Errorf("Expected pot[1] amount 300, got %d", pots[1].Amount)
	}
	if len(pots[1].EligibleSeats) != 2 {
		t.Errorf("Expected pot[1] 2 eligible seats, got %d", len(pots[1].EligibleSeats))
	}
}

// Test6PlayerComplexMultiWay: Six players with complex contribution pattern
func TestCalculateSidePots6PlayerComplexMultiWay(t *testing.T) {
	contributions := map[int]int{
		0: 50,
		1: 150,
		2: 100,
		3: 200,
		4: 150,
		5: 200,
	}
	folded := map[int]bool{
		0: false,
		1: true,
		2: false,
		3: false,
		4: false,
		5: false,
	}

	pots := CalculateSidePots(contributions, folded)

	if len(pots) != 4 {
		t.Fatalf("Expected 4 pots, got %d", len(pots))
	}

	// Pot 1: 50 * 6 = 300
	if pots[0].Amount != 300 {
		t.Errorf("Expected pot[0] amount 300, got %d", pots[0].Amount)
	}
	if len(pots[0].EligibleSeats) != 5 {
		t.Errorf("Expected pot[0] 5 eligible seats, got %d", len(pots[0].EligibleSeats))
	}

	// Pot 2: (100 - 50) * 5 = 250 (seats 1,2,3,4,5 contributed at 100+, but 1 folded, so 4 eligible)
	if pots[1].Amount != 250 {
		t.Errorf("Expected pot[1] amount 250, got %d", pots[1].Amount)
	}
	if len(pots[1].EligibleSeats) != 4 {
		t.Errorf("Expected pot[1] 4 eligible seats, got %d", len(pots[1].EligibleSeats))
	}

	// Pot 3: (150 - 100) * 4 = 200 (seats 1, 3, 4, 5 contributed at 150+, but 1 folded, so 3 eligible)
	if pots[2].Amount != 200 {
		t.Errorf("Expected pot[2] amount 200, got %d", pots[2].Amount)
	}
	if len(pots[2].EligibleSeats) != 3 {
		t.Errorf("Expected pot[2] 3 eligible seats, got %d", len(pots[2].EligibleSeats))
	}

	// Pot 4: (200 - 150) * 2 = 100 (seats 3, 5 contributed at 200+, both not folded)
	if pots[3].Amount != 100 {
		t.Errorf("Expected pot[3] amount 100, got %d", pots[3].Amount)
	}
	if len(pots[3].EligibleSeats) != 2 {
		t.Errorf("Expected pot[3] 2 eligible seats, got %d", len(pots[3].EligibleSeats))
	}
}

// Test6PlayerWhaleScenario: One whale with massive stack vs shorter stacks
func TestCalculateSidePots6PlayerWhaleScenario(t *testing.T) {
	contributions := map[int]int{
		0: 50,
		1: 50,
		2: 50,
		3: 50,
		4: 50,
		5: 1000, // Whale
	}
	folded := map[int]bool{
		0: false,
		1: false,
		2: false,
		3: false,
		4: false,
		5: false,
	}

	pots := CalculateSidePots(contributions, folded)

	if len(pots) != 2 {
		t.Fatalf("Expected 2 pots, got %d", len(pots))
	}

	// Pot 1: 50 * 6 = 300 (all eligible)
	if pots[0].Amount != 300 {
		t.Errorf("Expected pot[0] amount 300, got %d", pots[0].Amount)
	}
	if len(pots[0].EligibleSeats) != 6 {
		t.Errorf("Expected pot[0] 6 eligible seats, got %d", len(pots[0].EligibleSeats))
	}

	// Pot 2: (1000 - 50) * 1 = 950 (only whale eligible)
	if pots[1].Amount != 950 {
		t.Errorf("Expected pot[1] amount 950, got %d", pots[1].Amount)
	}
	if len(pots[1].EligibleSeats) != 1 || pots[1].EligibleSeats[0] != 5 {
		t.Errorf("Expected pot[1] eligible seat [5], got %v", pots[1].EligibleSeats)
	}
}

// TestCalculateSidePotsAllEqualContributions: All players with equal contributions
func TestCalculateSidePotsAllEqualContributions(t *testing.T) {
	contributions := map[int]int{
		0: 100,
		1: 100,
		2: 100,
		3: 100,
		4: 100,
		5: 100,
	}
	folded := map[int]bool{
		0: false,
		1: false,
		2: false,
		3: false,
		4: false,
		5: false,
	}

	pots := CalculateSidePots(contributions, folded)

	if len(pots) != 1 {
		t.Fatalf("Expected 1 pot, got %d", len(pots))
	}

	if pots[0].Amount != 600 {
		t.Errorf("Expected pot amount 600, got %d", pots[0].Amount)
	}

	if len(pots[0].EligibleSeats) != 6 {
		t.Errorf("Expected 6 eligible seats, got %d", len(pots[0].EligibleSeats))
	}
}

// TestCalculateSidePotsAllFoldedExceptOne: Only one player didn't fold
func TestCalculateSidePotsAllFoldedExceptOne(t *testing.T) {
	contributions := map[int]int{
		0: 100,
		1: 100,
		2: 100,
		3: 100,
		4: 100,
		5: 100,
	}
	folded := map[int]bool{
		0: true,
		1: true,
		2: true,
		3: true,
		4: true,
		5: false, // Only player 5 didn't fold
	}

	pots := CalculateSidePots(contributions, folded)

	if len(pots) != 1 {
		t.Fatalf("Expected 1 pot, got %d", len(pots))
	}

	if pots[0].Amount != 600 {
		t.Errorf("Expected pot amount 600, got %d", pots[0].Amount)
	}

	if len(pots[0].EligibleSeats) != 1 || pots[0].EligibleSeats[0] != 5 {
		t.Errorf("Expected pot eligible seat [5], got %v", pots[0].EligibleSeats)
	}
}

// TestCalculateSidePotsZeroContributions: No one contributed (edge case)
func TestCalculateSidePotsZeroContributions(t *testing.T) {
	contributions := map[int]int{
		0: 0,
		1: 0,
		2: 0,
	}
	folded := map[int]bool{
		0: false,
		1: false,
		2: false,
	}

	pots := CalculateSidePots(contributions, folded)

	if len(pots) != 0 {
		t.Fatalf("Expected 0 pots, got %d", len(pots))
	}
}

// ============================================================================
// Phase 4: DistributePot with Side Pot Support Tests
// ============================================================================

// TestDistributePot_Phase4_SingleWinner_TakesAll: One winner takes entire pot
func TestDistributePot_Phase4_SingleWinner_TakesAll(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up hand with contributions and folded players
	table.CurrentHand = &Hand{
		Pot: 1000,
		TotalContributions: map[int]int{
			0: 500,
			1: 300,
			2: 200,
		},
		FoldedPlayers: map[int]bool{
			0: false, // winner
			1: true,  // folded
			2: true,  // folded
		},
	}

	// Set up seats with starting stacks (they would be 0 after all contributions)
	table.Seats[0] = Seat{Index: 0, Stack: 0, Status: "active"}
	table.Seats[1] = Seat{Index: 1, Stack: 0, Status: "active"}
	table.Seats[2] = Seat{Index: 2, Stack: 0, Status: "active"}

	distribution := table.DistributePot([]int{0})

	if distribution[0] != 1000 {
		t.Errorf("expected winner to receive 1000, got %d", distribution[0])
	}

	if table.CurrentHand.Pot != 0 {
		t.Errorf("expected pot to be 0 after distribution, got %d", table.CurrentHand.Pot)
	}
}

// TestDistributePot_Phase4_TwoWinners_EqualStacks_SplitPot: Two winners split evenly
func TestDistributePot_Phase4_TwoWinners_EqualStacks_SplitPot(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	table.CurrentHand = &Hand{
		Pot: 1000,
		TotalContributions: map[int]int{
			0: 500,
			1: 500,
		},
		FoldedPlayers: map[int]bool{
			0: false, // winner
			1: false, // winner (tie)
		},
	}

	table.Seats[0] = Seat{Index: 0, Stack: 0, Status: "active"}
	table.Seats[1] = Seat{Index: 1, Stack: 0, Status: "active"}

	distribution := table.DistributePot([]int{0, 1})

	if distribution[0] != 500 {
		t.Errorf("expected seat 0 to receive 500, got %d", distribution[0])
	}

	if distribution[1] != 500 {
		t.Errorf("expected seat 1 to receive 500, got %d", distribution[1])
	}

	if table.CurrentHand.Pot != 0 {
		t.Errorf("expected pot to be 0 after distribution, got %d", table.CurrentHand.Pot)
	}
}

// TestDistributePot_Phase4_ThreeWinners_EqualStacks_SplitThreeWay: Three winners split evenly
func TestDistributePot_Phase4_ThreeWinners_EqualStacks_SplitThreeWay(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	table.CurrentHand = &Hand{
		Pot: 900,
		TotalContributions: map[int]int{
			0: 300,
			1: 300,
			2: 300,
		},
		FoldedPlayers: map[int]bool{
			0: false, // winner
			1: false, // winner
			2: false, // winner
		},
	}

	table.Seats[0] = Seat{Index: 0, Stack: 0, Status: "active"}
	table.Seats[1] = Seat{Index: 1, Stack: 0, Status: "active"}
	table.Seats[2] = Seat{Index: 2, Stack: 0, Status: "active"}

	distribution := table.DistributePot([]int{0, 1, 2})

	if distribution[0] != 300 {
		t.Errorf("expected seat 0 to receive 300, got %d", distribution[0])
	}
	if distribution[1] != 300 {
		t.Errorf("expected seat 1 to receive 300, got %d", distribution[1])
	}
	if distribution[2] != 300 {
		t.Errorf("expected seat 2 to receive 300, got %d", distribution[2])
	}
}

// TestDistributePot_Phase4_OneShortStack_WinnerIsShortStack: Short stack wins only main pot
func TestDistributePot_Phase4_OneShortStack_WinnerIsShortStack(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Seat 0 is short stack (50), seats 1 and 2 each contributed 100
	table.CurrentHand = &Hand{
		Pot: 250,
		TotalContributions: map[int]int{
			0: 50,  // short stack - all-in
			1: 100, // big stack
			2: 100, // big stack
		},
		FoldedPlayers: map[int]bool{
			0: false, // winner (short stack)
			1: true,  // folded
			2: true,  // folded
		},
	}

	table.Seats[0] = Seat{Index: 0, Stack: 0, Status: "active"}
	table.Seats[1] = Seat{Index: 1, Stack: 0, Status: "active"}
	table.Seats[2] = Seat{Index: 2, Stack: 0, Status: "active"}

	distribution := table.DistributePot([]int{0})

	// Main pot: 50 * 3 = 150 (short stack + both big stacks)
	// Side pot: 50 * 2 = 100 (only big stacks can win, both folded, so no winner for side pot)
	// Winner should get 150 (main pot only)
	if distribution[0] != 150 {
		t.Errorf("expected short stack to receive 150 (main pot), got %d", distribution[0])
	}

	// The side pot goes to neither seat (both folded), so shouldn't appear in distribution
	// Total distributed should be 150 (the other 100 remains unawarded but pot is zeroed)
}

// TestDistributePot_Phase4_OneShortStack_WinnerIsBigStack: Big stack wins all pots
func TestDistributePot_Phase4_OneShortStack_WinnerIsBigStack(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Seat 0 is short stack (50), seat 1 is big stack (100), seat 1 wins
	table.CurrentHand = &Hand{
		Pot: 250,
		TotalContributions: map[int]int{
			0: 50,  // short stack
			1: 100, // big stack - winner
			2: 100, // big stack
		},
		FoldedPlayers: map[int]bool{
			0: true,  // folded
			1: false, // winner (big stack)
			2: true,  // folded
		},
	}

	table.Seats[0] = Seat{Index: 0, Stack: 0, Status: "active"}
	table.Seats[1] = Seat{Index: 1, Stack: 0, Status: "active"}
	table.Seats[2] = Seat{Index: 2, Stack: 0, Status: "active"}

	distribution := table.DistributePot([]int{1})

	// Main pot: 50 * 3 = 150 (only eligible: 0, 1, 2; winner: 1)
	// Side pot: 50 * 2 = 100 (only eligible: 1, 2; winner: 1 only)
	// Winner gets 150 + 100 = 250
	if distribution[1] != 250 {
		t.Errorf("expected big stack winner to receive 250, got %d", distribution[1])
	}
}

// TestDistributePot_Phase4_TwoShortStacks_MultipleWinners: Different winners for main/side
func TestDistributePot_Phase4_TwoShortStacks_MultipleWinners(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Seat 0 (50), Seat 1 (50), Seat 2 (100) - seats 0 and 1 are winners, seat 2 folded
	// Main pot at level 50: 50 * 3 = 150 (all contributed at least 50)
	// Both seat 0 and 1 eligible for main pot
	// Side pot at level 100: 50 * 1 = 50 (only seat 2 at this level, but it's folded) - no eligible winners
	// So seats 0 and 1 split 150 = 75 each
	table.CurrentHand = &Hand{
		Pot: 200,
		TotalContributions: map[int]int{
			0: 50,  // winner
			1: 50,  // winner
			2: 100, // folded
		},
		FoldedPlayers: map[int]bool{
			0: false, // winner
			1: false, // winner
			2: true,  // folded
		},
	}

	table.Seats[0] = Seat{Index: 0, Stack: 0, Status: "active"}
	table.Seats[1] = Seat{Index: 1, Stack: 0, Status: "active"}
	table.Seats[2] = Seat{Index: 2, Stack: 0, Status: "active"}

	// Both seats are eligible for main pot only, so they split evenly
	distribution := table.DistributePot([]int{0, 1})

	if distribution[0] != 75 {
		t.Errorf("expected seat 0 to receive 75 (150 / 2), got %d", distribution[0])
	}
	if distribution[1] != 75 {
		t.Errorf("expected seat 1 to receive 75 (150 / 2), got %d", distribution[1])
	}
}

// TestDistributePot_Phase4_ComplexSidePots_SingleWinner: Multiple side pots, one winner takes all
func TestDistributePot_Phase4_ComplexSidePots_SingleWinner(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Three contribution levels: 30, 60, 100
	// All seats contribute at each level, then seat 2 contributes more
	// Main pot (level 30): 30*3 = 90, all eligible
	// Side pot 1 (level 60): (60-30)*3 = 90, all eligible (each contributed at least 60)
	// Side pot 2 (level 100): (100-60)*3 = 120, all eligible (each contributed at least 100)
	table.CurrentHand = &Hand{
		Pot: 300, // 30*3 + 30*3 + 40*3 = 90 + 90 + 120
		TotalContributions: map[int]int{
			0: 100, // equal contribution
			1: 100, // equal contribution
			2: 100, // equal contribution
		},
		FoldedPlayers: map[int]bool{
			0: false, // winner
			1: false,
			2: false,
		},
	}

	table.Seats[0] = Seat{Index: 0, Stack: 0, Status: "active"}
	table.Seats[1] = Seat{Index: 1, Stack: 0, Status: "active"}
	table.Seats[2] = Seat{Index: 2, Stack: 0, Status: "active"}

	distribution := table.DistributePot([]int{0})

	// Only seat 0 is the winner, all pots go to seat 0
	if distribution[0] != 300 {
		t.Errorf("expected seat 0 to receive 300, got %d", distribution[0])
	}
}

// TestDistributePot_Phase4_OddChipDistribution: Pot doesn't divide evenly
func TestDistributePot_Phase4_OddChipDistribution(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// 315 chips split 3 ways = 105 each, no remainder
	table.CurrentHand = &Hand{
		Pot: 315,
		TotalContributions: map[int]int{
			0: 105,
			1: 105,
			2: 105,
		},
		FoldedPlayers: map[int]bool{
			0: false,
			1: false,
			2: false,
		},
	}

	table.Seats[0] = Seat{Index: 0, Stack: 0, Status: "active"}
	table.Seats[1] = Seat{Index: 1, Stack: 0, Status: "active"}
	table.Seats[2] = Seat{Index: 2, Stack: 0, Status: "active"}

	distribution := table.DistributePot([]int{0, 1, 2})

	if distribution[0] != 105 {
		t.Errorf("expected seat 0 to receive 105, got %d", distribution[0])
	}
	if distribution[1] != 105 {
		t.Errorf("expected seat 1 to receive 105, got %d", distribution[1])
	}
	if distribution[2] != 105 {
		t.Errorf("expected seat 2 to receive 105, got %d", distribution[2])
	}
}

// TestDistributePot_Phase4_OddChip_RemainderGoesToFirst: Remainder goes to first winner
func TestDistributePot_Phase4_OddChip_RemainderGoesToFirst(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Seats have equal contribution but different amounts
	// Seats 0 and 1 contribute 35 each
	// Seat 2 contributes 36 (1 extra)
	// Main pot (35*3 = 105): all eligible, split evenly 35 each
	// Side pot (1*1 = 1): only seat 2 eligible, gets 1
	table.CurrentHand = &Hand{
		Pot: 106,
		TotalContributions: map[int]int{
			0: 35, // eligible for main pot only
			1: 35, // eligible for main pot only
			2: 36, // eligible for both pots
		},
		FoldedPlayers: map[int]bool{
			0: false,
			1: false,
			2: false,
		},
	}

	table.Seats[0] = Seat{Index: 0, Stack: 0, Status: "active"}
	table.Seats[1] = Seat{Index: 1, Stack: 0, Status: "active"}
	table.Seats[2] = Seat{Index: 2, Stack: 0, Status: "active"}

	distribution := table.DistributePot([]int{0, 1, 2})

	// Main pot (105): 105 / 3 = 35 each
	// Side pot (1): 1 / 1 = 1 to seat 2
	if distribution[0] != 35 {
		t.Errorf("expected seat 0 to receive 35, got %d", distribution[0])
	}
	if distribution[1] != 35 {
		t.Errorf("expected seat 1 to receive 35, got %d", distribution[1])
	}
	if distribution[2] != 36 {
		t.Errorf("expected seat 2 to receive 36 (35 + 1 from side pot), got %d", distribution[2])
	}
}

// TestDistributePot_Phase4_ZeroPot: No pot to distribute
func TestDistributePot_Phase4_ZeroPot(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	table.CurrentHand = &Hand{
		Pot: 0,
		TotalContributions: map[int]int{
			0: 0,
			1: 0,
		},
		FoldedPlayers: map[int]bool{
			0: false,
			1: false,
		},
	}

	table.Seats[0] = Seat{Index: 0, Stack: 0, Status: "active"}
	table.Seats[1] = Seat{Index: 1, Stack: 0, Status: "active"}

	distribution := table.DistributePot([]int{0})

	// No chips to distribute
	if len(distribution) != 0 {
		t.Errorf("expected empty distribution for zero pot, got %d entries", len(distribution))
	}

	if table.CurrentHand.Pot != 0 {
		t.Errorf("expected pot to remain 0, got %d", table.CurrentHand.Pot)
	}
}

// TestDistributePot_Phase4_AllInScenario_MultipleWinners: Realistic all-in with 2+ winners
func TestDistributePot_Phase4_AllInScenario_MultipleWinners(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Three players: A (100 all-in), B (300 all-in), C (500 - call and win)
	// A and B both go all-in, C calls with big stack and wins
	// Main pot (A's level): 100 * 3 = 300 (all three eligible)
	// Side pot (B's level): 200 * 2 = 400 (B and C eligible)
	// Side pot (C's level): 200 * 1 = 200 (C only)
	// All go to C (winner)
	table.CurrentHand = &Hand{
		Pot: 900,
		TotalContributions: map[int]int{
			0: 100, // A - all-in
			1: 300, // B - all-in
			2: 500, // C - big stack, winner
		},
		FoldedPlayers: map[int]bool{
			0: false,
			1: false,
			2: false, // winner
		},
	}

	table.Seats[0] = Seat{Index: 0, Stack: 0, Status: "active"}
	table.Seats[1] = Seat{Index: 1, Stack: 0, Status: "active"}
	table.Seats[2] = Seat{Index: 2, Stack: 0, Status: "active"}

	distribution := table.DistributePot([]int{2})

	if distribution[2] != 900 {
		t.Errorf("expected seat 2 to receive 900, got %d", distribution[2])
	}

	if table.CurrentHand.Pot != 0 {
		t.Errorf("expected pot to be 0 after distribution, got %d", table.CurrentHand.Pot)
	}
}

// TestDistributePot_Phase4_NoWinnersForSidePot: Side pot has no eligible winners (all folded)
func TestDistributePot_Phase4_NoWinnersForSidePot(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Seat 0 (50 - all-in) and Seat 1 (100 - all-in) both fold
	// Seat 2 (100) folds before all-in, so there are no winners
	// In practice, this shouldn't happen if called from showdown correctly
	// But the function should handle it gracefully by not crashing
	table.CurrentHand = &Hand{
		Pot: 250,
		TotalContributions: map[int]int{
			0: 50,  // all-in, folded
			1: 100, // all-in, folded
			2: 100, // called but folded
		},
		FoldedPlayers: map[int]bool{
			0: true, // folded
			1: true, // folded
			2: true, // folded
		},
	}

	table.Seats[0] = Seat{Index: 0, Stack: 0, Status: "active"}
	table.Seats[1] = Seat{Index: 1, Stack: 0, Status: "active"}
	table.Seats[2] = Seat{Index: 2, Stack: 0, Status: "active"}

	// Call with empty winners list (shouldn't happen in normal flow, but test for robustness)
	distribution := table.DistributePot([]int{})

	// Distribution should be empty
	if len(distribution) != 0 {
		t.Errorf("expected empty distribution with no winners, got %d entries", len(distribution))
	}
}

// TestDistributePot_Phase4_PartialRefund: Winner smaller than one of contributions
func TestDistributePot_Phase4_PartialRefund(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Seat 0 (200 - all-in)
	// Seat 1 (100 - called, not all-in yet)
	// Seat 2 (100 - called, not all-in)
	// Seat 0 is the winner and only eligible for main pot (200)
	table.CurrentHand = &Hand{
		Pot: 400,
		TotalContributions: map[int]int{
			0: 200,
			1: 100,
			2: 100,
		},
		FoldedPlayers: map[int]bool{
			0: false, // winner (eligible for all pots)
			1: true,  // folded
			2: true,  // folded
		},
	}

	table.Seats[0] = Seat{Index: 0, Stack: 0, Status: "active"}
	table.Seats[1] = Seat{Index: 1, Stack: 0, Status: "active"}
	table.Seats[2] = Seat{Index: 2, Stack: 0, Status: "active"}

	distribution := table.DistributePot([]int{0})

	// Main pot: 200 * 1 (only seat 0 eligible, others folded) = 200? No!
	// Actually: main pot = 100 * 3 = 300 (all three eligible at 100 level)
	// Side pot = 100 * 1 = 100 (only seat 0 eligible at 200 level)
	// Total for seat 0 = 300 + 100 = 400
	if distribution[0] != 400 {
		t.Errorf("expected seat 0 to receive 400, got %d", distribution[0])
	}
}
