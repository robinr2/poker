package server

import (
	"crypto/rand"
	"fmt"
	"math/big"
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
	ID          string
	Name        string
	MaxSeats    int     // Always 6
	Seats       [6]Seat // Fixed array of 6 seats
	DealerSeat  *int    // Seat number of the current dealer (nil = no dealer assigned yet)
	CurrentHand *Hand   // Currently active hand (nil = no hand running)
	Server      *Server // Reference to the server for broadcasting events
	mu          sync.RWMutex
}

// NewTable creates and returns a new Table instance with 6 empty seats
func NewTable(id, name string, server *Server) *Table {
	table := &Table{
		ID:       id,
		Name:     name,
		MaxSeats: 6,
		Server:   server,
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

// ShuffleDeck shuffles a deck of cards in place using crypto/rand and Fisher-Yates algorithm
// Returns error if random number generation fails
func ShuffleDeck(deck []Card) error {
	// Fisher-Yates shuffle using cryptographically secure random
	for i := len(deck) - 1; i > 0; i-- {
		// Generate random number from 0 to i (inclusive) using crypto/rand
		max := big.NewInt(int64(i + 1))
		randomBig, err := rand.Int(rand.Reader, max)
		if err != nil {
			return fmt.Errorf("failed to generate random number: %w", err)
		}

		j := int(randomBig.Int64())

		// Swap deck[i] with deck[j]
		deck[i], deck[j] = deck[j], deck[i]
	}

	return nil
}

// DealHoleCards deals 2 cards to each active player from the deck
// Only seats with Status == "active" receive cards
// Updates h.HoleCards and removes cards from h.Deck
// Returns error if unable to shuffle or if not enough cards in deck
func (h *Hand) DealHoleCards(seats [6]Seat) error {
	// Identify active seats
	activeSeats := []int{}
	for i := 0; i < 6; i++ {
		if seats[i].Status == "active" {
			activeSeats = append(activeSeats, i)
		}
	}

	// Check if we have enough cards in deck (2 per active player)
	cardsNeeded := len(activeSeats) * 2
	if len(h.Deck) < cardsNeeded {
		return fmt.Errorf("insufficient cards in deck: have %d, need %d", len(h.Deck), cardsNeeded)
	}

	// Initialize HoleCards map if needed
	if h.HoleCards == nil {
		h.HoleCards = make(map[int][]Card)
	}

	// Deal 2 cards to each active seat
	cardIndex := 0
	for _, seatNum := range activeSeats {
		// Deal 2 cards
		holeCards := make([]Card, 2)
		holeCards[0] = h.Deck[cardIndex]
		holeCards[1] = h.Deck[cardIndex+1]
		cardIndex += 2

		// Store in HoleCards map
		h.HoleCards[seatNum] = holeCards
	}

	// Remove dealt cards from deck
	h.Deck = h.Deck[cardIndex:]

	return nil
}

// CanStartHand checks if a new hand can be started
// Returns true if:
// - At least 2 players exist (waiting or active status)
// - No hand is currently running (CurrentHand == nil)
// Returns false otherwise
func (t *Table) CanStartHand() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Check if a hand is already running
	if t.CurrentHand != nil {
		return false
	}

	// Count players (both "waiting" and "active" can start a hand)
	playerCount := 0
	for i := 0; i < 6; i++ {
		if t.Seats[i].Status == "waiting" || t.Seats[i].Status == "active" {
			playerCount++
		}
	}

	// Need at least 2 players
	return playerCount >= 2
}

// StartHand initializes and starts a new poker hand
// This method orchestrates the full hand start sequence:
// 1. Transitions "waiting" players to "active" status
// 2. Validates that a hand can be started (≥2 active players, no hand running)
// 3. Assigns dealer via NextDealer()
// 4. Gets blind positions
// 5. Creates new deck and shuffles
// 6. Posts blinds (SB=10, BB=20), handles all-in if stack < blind
// 7. Deals hole cards to all active players
// 8. Sets CurrentHand with all game state
// 9. Broadcasts hand_started, blind_posted, and cards_dealt events
// Returns error if hand cannot be started or if operations fail
func (t *Table) StartHand() error {
	t.mu.Lock()

	// Step 0: Transition all "waiting" players to "active" status
	// Players become active when the first/next hand starts
	for i := 0; i < 6; i++ {
		if t.Seats[i].Status == "waiting" {
			t.Seats[i].Status = "active"
		}
	}

	// Check if hand can be started (must do this with lock held)
	// Count active players
	activeCount := 0
	for i := 0; i < 6; i++ {
		if t.Seats[i].Status == "active" {
			activeCount++
		}
	}

	if activeCount < 2 {
		t.mu.Unlock()
		return fmt.Errorf("insufficient active players to start hand: %d active, need at least 2", activeCount)
	}

	if t.CurrentHand != nil {
		t.mu.Unlock()
		return fmt.Errorf("hand already running")
	}

	// Step 1: Assign dealer via NextDealer
	dealerSeat := t.assignDealerLocked()

	// Step 2: Get blind positions
	sbSeat, bbSeat, err := t.getBlindPositionsLocked(dealerSeat)
	if err != nil {
		t.mu.Unlock()
		return fmt.Errorf("failed to get blind positions: %w", err)
	}

	// Step 3: Create new hand and deck
	hand := &Hand{
		DealerSeat:     dealerSeat,
		SmallBlindSeat: sbSeat,
		BigBlindSeat:   bbSeat,
		Pot:            0,
		Deck:           NewDeck(),
		HoleCards:      make(map[int][]Card),
	}

	// Step 4: Shuffle the deck
	err = ShuffleDeck(hand.Deck)
	if err != nil {
		t.mu.Unlock()
		return fmt.Errorf("failed to shuffle deck: %w", err)
	}

	// Step 5: Post blinds (handle all-in if necessary)
	const smallBlind = 10
	const bigBlind = 20

	// Post small blind
	sbPosted := smallBlind
	if t.Seats[sbSeat].Stack < smallBlind {
		// All-in with remaining chips
		sbPosted = t.Seats[sbSeat].Stack
		t.Seats[sbSeat].Stack = 0
	} else {
		t.Seats[sbSeat].Stack -= smallBlind
	}

	// Post big blind
	bbPosted := bigBlind
	if t.Seats[bbSeat].Stack < bigBlind {
		// All-in with remaining chips
		bbPosted = t.Seats[bbSeat].Stack
		t.Seats[bbSeat].Stack = 0
	} else {
		t.Seats[bbSeat].Stack -= bigBlind
	}

	hand.Pot = sbPosted + bbPosted

	// Step 6: Deal hole cards to all active players
	err = hand.DealHoleCards(t.Seats)
	if err != nil {
		t.mu.Unlock()
		return fmt.Errorf("failed to deal hole cards: %w", err)
	}

	// Step 7: Set CurrentHand
	t.CurrentHand = hand

	// Unlock before broadcasting to avoid holding the lock during network operations
	t.mu.Unlock()

	// Step 8: Broadcast events to all table clients
	if t.Server != nil {
		// Broadcast hand_started with dealer and blind positions
		err = t.Server.broadcastHandStarted(t)
		if err != nil {
			t.mu.Lock()
			// Revert the hand state on broadcast failure
			t.CurrentHand = nil
			t.mu.Unlock()
			return fmt.Errorf("failed to broadcast hand_started: %w", err)
		}

		// Broadcast small blind posted
		err = t.Server.broadcastBlindPosted(t, sbSeat, sbPosted)
		if err != nil {
			t.mu.Lock()
			// Revert the hand state on broadcast failure
			t.CurrentHand = nil
			t.mu.Unlock()
			return fmt.Errorf("failed to broadcast small blind: %w", err)
		}

		// Broadcast big blind posted
		err = t.Server.broadcastBlindPosted(t, bbSeat, bbPosted)
		if err != nil {
			t.mu.Lock()
			// Revert the hand state on broadcast failure
			t.CurrentHand = nil
			t.mu.Unlock()
			return fmt.Errorf("failed to broadcast big blind: %w", err)
		}

		// Broadcast hole cards dealt
		err = t.Server.broadcastCardsDealt(t)
		if err != nil {
			t.mu.Lock()
			// Revert the hand state on broadcast failure
			t.CurrentHand = nil
			t.mu.Unlock()
			return fmt.Errorf("failed to broadcast cards_dealt: %w", err)
		}
	}

	return nil
}

// assignDealerLocked assigns the next dealer seat (internal, must be called with lock held)
// For the first hand (DealerSeat is nil), it finds the first active seat.
// For subsequent hands, it rotates clockwise to the next active seat.
// Only seats with "active" status are eligible for dealer position.
func (t *Table) assignDealerLocked() int {
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

// getBlindPositionsLocked returns the seat numbers for small blind and big blind (internal, must be called with lock held)
// - Returns error if fewer than 2 active players
// - For heads-up (exactly 2 active players): dealer is small blind, other is big blind
// - For normal (3+ active players): small blind is next active after dealer, big blind is next after small blind
func (t *Table) getBlindPositionsLocked(dealerSeat int) (int, int, error) {
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
