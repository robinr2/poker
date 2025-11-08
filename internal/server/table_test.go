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

// TestNextDealerFirstHand verifies first hand assigns dealer to first active seat
func TestNextDealerFirstHand(t *testing.T) {
	table := NewTable("table-1", "Table 1")

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
	table := NewTable("table-1", "Table 1")

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
	table := NewTable("table-1", "Table 1")

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
	table := NewTable("table-1", "Table 1")

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
	table := NewTable("table-1", "Table 1")

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
	table := NewTable("table-1", "Table 1")

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
	table := NewTable("table-1", "Table 1")

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
	table := NewTable("table-1", "Table 1")

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
	table := NewTable("table-1", "Table 1")

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
	table := NewTable("table-1", "Table 1")

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
	table := NewTable("table-1", "Table 1")

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
	table2 := NewTable("table-2", "Table 2")
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
	table := NewTable("table-1", "Table 1")

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
	table := NewTable("table-1", "Table 1")

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
	table := NewTable("table-1", "Table 1")

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
	table := NewTable("table-1", "Table 1")

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
	table := NewTable("table-1", "Table 1")

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
