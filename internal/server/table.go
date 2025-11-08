package server

import (
	"fmt"
	"sync"
)

// Card represents a playing card with rank and suit
type Card struct {
	Rank string // A, 2, 3, 4, 5, 6, 7, 8, 9, T, J, Q, K
	Suit string // s (spades), h (hearts), d (diamonds), c (clubs)
}

// String returns the 2-character string representation of a card (e.g., "As", "Kh")
func (c Card) String() string {
	return c.Rank + c.Suit
}

// NewDeck creates and returns a new 52-card deck with all unique cards
func NewDeck() []Card {
	ranks := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K"}
	suits := []string{"s", "h", "d", "c"}

	deck := make([]Card, 0, 52)
	for _, suit := range suits {
		for _, rank := range ranks {
			deck = append(deck, Card{Rank: rank, Suit: suit})
		}
	}
	return deck
}

// Hand represents the current game hand state
type Hand struct {
	DealerSeat     int            // Seat number of the dealer
	SmallBlindSeat int            // Seat number of the small blind
	BigBlindSeat   int            // Seat number of the big blind
	Pot            int            // Current pot amount
	Deck           []Card         // Cards remaining in the deck
	HoleCards      map[int][]Card // Hole cards for each seat (key = seat number, value = 2 cards)
}

// Seat represents a seat at a poker table
type Seat struct {
	Index  int     // 0-5
	Token  *string // nil = empty, non-nil = occupied
	Status string  // "empty", "waiting", "active"
	Stack  int     // Chip stack for the player (0 for empty seats, 1000 for new players)
}

// Table represents a poker table
type Table struct {
	ID         string
	Name       string
	MaxSeats   int     // Always 6
	Seats      [6]Seat // Fixed array of 6 seats
	DealerSeat *int    // Seat number of the current dealer (nil = no dealer assigned yet)
	mu         sync.RWMutex
}

// NewTable creates and returns a new Table instance with 6 empty seats
func NewTable(id, name string) *Table {
	table := &Table{
		ID:       id,
		Name:     name,
		MaxSeats: 6,
	}

	// Initialize all seats with Index and nil Token
	for i := 0; i < 6; i++ {
		table.Seats[i] = Seat{
			Index:  i,
			Token:  nil,
			Status: "empty",
		}
	}

	return table
}

// GetOccupiedSeatCount returns the count of seats with non-nil Token (thread-safe)
func (t *Table) GetOccupiedSeatCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	count := 0
	for _, seat := range t.Seats {
		if seat.Token != nil {
			count++
		}
	}
	return count
}

// AssignSeat assigns a player token to the first available seat (thread-safe)
// Returns the assigned seat (by value) and nil error on success
// Returns empty Seat and error if table is full
func (t *Table) AssignSeat(token *string) (Seat, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Find first empty seat
	for i := 0; i < 6; i++ {
		if t.Seats[i].Token == nil {
			t.Seats[i].Token = token
			t.Seats[i].Status = "waiting"
			t.Seats[i].Stack = 1000
			return t.Seats[i], nil
		}
	}

	// No empty seats found
	return Seat{}, fmt.Errorf("table is full")
}

// ClearSeat removes a player from a table by token (thread-safe)
// Returns nil error on success
// Returns error if token not found
func (t *Table) ClearSeat(token *string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Find seat with matching token
	for i := 0; i < 6; i++ {
		if t.Seats[i].Token != nil && *t.Seats[i].Token == *token {
			t.Seats[i].Token = nil
			t.Seats[i].Status = "empty"
			t.Seats[i].Stack = 0
			return nil
		}
	}

	// Token not found
	return fmt.Errorf("seat not found")
}

// GetSeatByToken returns the seat occupied by a player token (thread-safe)
// Returns the seat (by value) and true if found, empty Seat and false if not found
func (t *Table) GetSeatByToken(token *string) (Seat, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Find seat with matching token
	for i := 0; i < 6; i++ {
		if t.Seats[i].Token != nil && *t.Seats[i].Token == *token {
			return t.Seats[i], true
		}
	}

	// Not found
	return Seat{}, false
}

// NextDealer assigns the next dealer seat and returns the seat number.
// For the first hand (DealerSeat is nil), it finds the first active seat.
// For subsequent hands, it rotates clockwise to the next active seat.
// Only seats with "active" status are eligible for dealer position.
func (t *Table) NextDealer() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	var nextDealer int

	// If no dealer assigned yet (first hand), find first active seat
	if t.DealerSeat == nil {
		for i := 0; i < 6; i++ {
			if t.Seats[i].Status == "active" {
				nextDealer = i
				break
			}
		}
	} else {
		// Find next active seat after current dealer
		currentDealer := *t.DealerSeat
		nextDealer = currentDealer

		// Search for next active seat starting after current dealer
		for j := 0; j < 6; j++ {
			checkSeat := (currentDealer + 1 + j) % 6
			if t.Seats[checkSeat].Status == "active" {
				nextDealer = checkSeat
				break
			}
		}
	}

	// Update DealerSeat field
	t.DealerSeat = &nextDealer
	return nextDealer
}

// GetBlindPositions returns the seat numbers for small blind and big blind.
// - Returns error if fewer than 2 active players
// - For heads-up (exactly 2 active players): dealer is small blind, other is big blind
// - For normal (3+ active players): small blind is next active after dealer, big blind is next after small blind
func (t *Table) GetBlindPositions(dealerSeat int) (int, int, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Count active players and find their seat numbers
	activePlayers := []int{}
	for i := 0; i < 6; i++ {
		if t.Seats[i].Status == "active" {
			activePlayers = append(activePlayers, i)
		}
	}

	// Error if fewer than 2 active players
	if len(activePlayers) < 2 {
		return 0, 0, fmt.Errorf("insufficient active players for blinds: %d active, need at least 2", len(activePlayers))
	}

	// Heads-up (exactly 2 active players): dealer is SB, other is BB
	if len(activePlayers) == 2 {
		// Find the other active player (not the dealer)
		var otherPlayer int
		if activePlayers[0] == dealerSeat {
			otherPlayer = activePlayers[1]
		} else {
			otherPlayer = activePlayers[0]
		}
		return dealerSeat, otherPlayer, nil
	}

	// Normal case (3+ active players): SB is next active after dealer, BB is next after SB
	// Find index of dealer in activePlayers array
	dealerIndex := -1
	for i, seat := range activePlayers {
		if seat == dealerSeat {
			dealerIndex = i
			break
		}
	}

	// Validate that dealer seat is active
	if dealerIndex == -1 {
		return 0, 0, fmt.Errorf("dealer seat %d is not active", dealerSeat)
	}

	// SB is next active player after dealer
	sbIndex := (dealerIndex + 1) % len(activePlayers)
	smallBlind := activePlayers[sbIndex]

	// BB is next active player after SB
	bbIndex := (sbIndex + 1) % len(activePlayers)
	bigBlind := activePlayers[bbIndex]

	return smallBlind, bigBlind, nil
}
