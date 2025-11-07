package server

import "sync"

// Seat represents a seat at a poker table
type Seat struct {
	Index int     // 0-5
	Token *string // nil = empty, non-nil = occupied
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
			Index: i,
			Token: nil,
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
