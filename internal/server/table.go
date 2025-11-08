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
	ID       string
	Name     string
	MaxSeats int     // Always 6
	Seats    [6]Seat // Fixed array of 6 seats
	mu       sync.RWMutex
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
