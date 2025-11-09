package server

import (
	"log/slog"
	"sync"
	"testing"
)

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

	// Verify pot = 30 (SB 10 + BB 20)
	if table.CurrentHand.Pot != 30 {
		t.Errorf("expected pot 30, got %d", table.CurrentHand.Pot)
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

	// Verify pot = SB 10 + BB 20 = 30
	if table.CurrentHand.Pot != 30 {
		t.Errorf("expected pot 30, got %d", table.CurrentHand.Pot)
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

	// Verify pot = SB(5 all-in) + BB(20) = 25
	if table.CurrentHand.Pot != 25 {
		t.Errorf("expected pot 25 (5 SB all-in + 20 BB), got %d", table.CurrentHand.Pot)
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

	// Should allow check and fold
	hasCheck := false
	hasFold := false
	for _, action := range validActions {
		if action == "check" {
			hasCheck = true
		}
		if action == "fold" {
			hasFold = true
		}
	}

	if !hasCheck {
		t.Errorf("expected 'check' in valid actions, got %v", validActions)
	}
	if !hasFold {
		t.Errorf("expected 'fold' in valid actions, got %v", validActions)
	}
	if len(validActions) != 2 {
		t.Errorf("expected exactly 2 valid actions (check, fold), got %d: %v", len(validActions), validActions)
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
	initialPot := table.CurrentHand.Pot

	// Process call action (need to call 40 more)
	chipsMoved, err := table.CurrentHand.ProcessAction(0, "call", table.Seats[0].Stack)
	if err != nil {
		t.Errorf("expected no error processing call, got %v", err)
	}

	// Verify correct amount of chips were moved
	if chipsMoved != 40 {
		t.Errorf("expected 40 chips moved, got %d", chipsMoved)
	}

	// Verify pot increased by call amount (40)
	expectedPot := initialPot + 40
	if table.CurrentHand.Pot != expectedPot {
		t.Errorf("expected pot to be %d, got %d", expectedPot, table.CurrentHand.Pot)
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
	initialPot := table.CurrentHand.Pot

	// Process call action (go all-in with 30)
	chipsMoved, err := table.CurrentHand.ProcessAction(0, "call", playerStack)
	if err != nil {
		t.Errorf("expected no error processing partial call, got %v", err)
	}

	// Verify all-in amount returned (30)
	if chipsMoved != 30 {
		t.Errorf("expected 30 chips moved (all-in), got %d", chipsMoved)
	}

	// Verify pot increased by available chips (30)
	expectedPot := initialPot + 30
	if table.CurrentHand.Pot != expectedPot {
		t.Errorf("expected pot to be %d, got %d", expectedPot, table.CurrentHand.Pot)
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

// TestAdvanceAction moves to next active player and handles wrap-around
func TestAdvanceAction(t *testing.T) {
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
// After advancing from preflop to flop, LastRaise should reset to 0
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

	if hand.LastRaise != 0 {
		t.Errorf("LastRaise = %d, want 0 after street advance", hand.LastRaise)
	}

	if hand.CurrentBet != 0 {
		t.Errorf("CurrentBet = %d, want 0 after street advance", hand.CurrentBet)
	}
}

// ============ PHASE 2: MAX-RAISE AND SIDE POT PREVENTION TESTS ============

// TestGetMaxRaise_LimitedByPlayerStack verifies max raise equals minimum of player and smallest opponent stack
func TestGetMaxRaise_LimitedByPlayerStack(t *testing.T) {
	table := NewTable("table-1", "Table 1", nil)

	// Set up 3 active players: seat 0 (1000), seat 1 (1000), seat 2 (1000)
	// All equal stacks - max raise should be limited by player's own stack
	for i := 0; i < 3; i++ {
		token := "player-" + string(rune('0'+i))
		table.Seats[i].Token = &token
		table.Seats[i].Status = "active"
		table.Seats[i].Stack = 1000
	}

	// GetMaxRaise for seat 0: min(1000 player, 1000 smallest opponent) = 1000
	maxRaise := table.GetMaxRaise(0, table.Seats)
	if maxRaise != 1000 {
		t.Errorf("GetMaxRaise(seat 0) = %d, want 1000 (when all equal)", maxRaise)
	}

	// Now set seat 2 to 500 (smaller than seat 0)
	table.Seats[2].Stack = 500

	// GetMaxRaise for seat 0: min(1000 player, 500 smallest opponent) = 500
	maxRaise = table.GetMaxRaise(0, table.Seats)
	if maxRaise != 500 {
		t.Errorf("GetMaxRaise(seat 0) = %d, want 500 (limited by opponent)", maxRaise)
	}

	// GetMaxRaise for seat 1: min(1000 player, 500 smallest opponent) = 500
	maxRaise = table.GetMaxRaise(1, table.Seats)
	if maxRaise != 500 {
		t.Errorf("GetMaxRaise(seat 1) = %d, want 500 (limited by opponent)", maxRaise)
	}
}

// TestGetMaxRaise_LimitedByOpponentStack verifies max raise limited to smallest opponent stack
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

	// Seat 0 player has 1000, but smallest opponent has 300
	// Max raise should be limited to 300 to prevent side pots
	maxRaise := table.GetMaxRaise(0, table.Seats)
	if maxRaise != 300 {
		t.Errorf("GetMaxRaise(seat 0) = %d, want 300 (smallest opponent stack)", maxRaise)
	}

	// Seat 1 player has 500, but smallest opponent (seat 2) has 300
	// Max raise should be limited to 300
	maxRaise = table.GetMaxRaise(1, table.Seats)
	if maxRaise != 300 {
		t.Errorf("GetMaxRaise(seat 1) = %d, want 300 (smallest opponent stack)", maxRaise)
	}

	// Seat 2 player has 300 (smallest), opponents have 1000 and 500
	// Max raise should be seat 2's own stack (300), which is smaller than opponent stacks
	maxRaise = table.GetMaxRaise(2, table.Seats)
	if maxRaise != 300 {
		t.Errorf("GetMaxRaise(seat 2) = %d, want 300 (player's own stack, smallest)", maxRaise)
	}
}

// TestGetMaxRaise_HeadsUp verifies heads-up allows full player stack
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

	// In heads-up, seat 0 can raise up to 1000 (limited by opponent's 800 to prevent side pot)
	// Wait, that's not right. In heads-up with unequal stacks:
	// Seat 0 (1000) vs Seat 3 (800): to avoid side pots, seat 0 can only raise 800
	// This is the same rule as multi-player: limited to smallest opponent stack
	maxRaise := table.GetMaxRaise(0, table.Seats)
	if maxRaise != 800 {
		t.Errorf("GetMaxRaise(seat 0 heads-up) = %d, want 800 (opponent's stack)", maxRaise)
	}

	// Seat 3 (800 stack) can raise up to 800 (their full stack or opponent's smaller)
	maxRaise = table.GetMaxRaise(3, table.Seats)
	if maxRaise != 800 {
		t.Errorf("GetMaxRaise(seat 3 heads-up) = %d, want 800 (their own or opponent's)", maxRaise)
	}
}

// TestGetMaxRaise_MultiPlayer verifies multi-player correctly limits to smallest stack
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

	// Seat 0 (1000): should be limited to smallest opponent (300)
	maxRaise := table.GetMaxRaise(0, table.Seats)
	if maxRaise != 300 {
		t.Errorf("GetMaxRaise(seat 0, 4-player) = %d, want 300", maxRaise)
	}

	// Seat 1 (600): should be limited to smallest opponent (300)
	maxRaise = table.GetMaxRaise(1, table.Seats)
	if maxRaise != 300 {
		t.Errorf("GetMaxRaise(seat 1, 4-player) = %d, want 300", maxRaise)
	}

	// Seat 2 (300): should be limited by its own stack (300)
	maxRaise = table.GetMaxRaise(2, table.Seats)
	if maxRaise != 300 {
		t.Errorf("GetMaxRaise(seat 2, 4-player) = %d, want 300", maxRaise)
	}

	// Seat 3 (800): should be limited to smallest opponent (300)
	maxRaise = table.GetMaxRaise(3, table.Seats)
	if maxRaise != 300 {
		t.Errorf("GetMaxRaise(seat 3, 4-player) = %d, want 300", maxRaise)
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

// TestValidateRaise_AboveMaximum verifies error with message "raise would create side pot"
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

	// Seat 0 tries to raise 500, but max allowed is 300 (smallest opponent)
	// This should error with side pot message
	err := hand.ValidateRaise(0, 500, 1000, table.Seats)
	if err == nil {
		t.Fatal("expected error for raise above maximum, got nil")
	}

	expectedMsg := "raise would create side pot"
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

// TestValidateRaise_HeadsUp verifies validation works correctly in heads-up
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

	// In heads-up, seat 0 (1000) can raise up to 800 (opponent's stack)
	// Raise of 40 (minimum) should be valid
	err := hand.ValidateRaise(0, 40, 1000, table.Seats)
	if err != nil {
		t.Errorf("expected no error for valid raise in heads-up, got %v", err)
	}

	// Raise of 800 (at max) should be valid
	err = hand.ValidateRaise(0, 800, 1000, table.Seats)
	if err != nil {
		t.Errorf("expected no error for max raise in heads-up, got %v", err)
	}

	// Raise of 900 (exceeds max of 800) should error
	err = hand.ValidateRaise(0, 900, 1000, table.Seats)
	if err == nil {
		t.Fatal("expected error for raise exceeding max in heads-up, got nil")
	}

	if err.Error() != "raise would create side pot" {
		t.Errorf("expected error 'raise would create side pot', got '%s'", err.Error())
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
	table.CurrentHand.Pot = 100
	if table.CurrentHand.PlayerBets == nil {
		table.CurrentHand.PlayerBets = make(map[int]int)
	}
	table.CurrentHand.PlayerBets[0] = 0

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

	// Pot should be updated (100 + 150 = 250)
	if table.CurrentHand.Pot != 250 {
		t.Errorf("expected Pot=250, got %d", table.CurrentHand.Pot)
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
