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
	BoardCards     []Card         // Community cards on the board (flop=3, turn=4, river=5)
	CurrentActor   *int           // Seat number of the player whose turn it is (nil if no active action)
	CurrentBet     int            // Current bet amount in this round (what players must match)
	PlayerBets     map[int]int    // Amount each player has bet in current round (key = seat number)
	FoldedPlayers  map[int]bool   // Players who have folded (key = seat number, value = true if folded)
	ActedPlayers   map[int]bool   // Players who have acted this round (key = seat number, value = true if acted)
	Street         string         // Current street: "preflop", "flop", "turn", "river"
	LastRaise      int            // Amount of the last raise increment (used to compute min-raise)
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

// DealFlop deals the flop (3 community cards) after burning 1 card
// Burn card is discarded (not stored)
// Returns error if deck has insufficient cards (need at least 4: 1 burn + 3 flop)
func (h *Hand) DealFlop() error {
	// Check if we have enough cards in deck (1 burn + 3 flop = 4 total)
	if len(h.Deck) < 4 {
		return fmt.Errorf("insufficient cards in deck: have %d, need 4", len(h.Deck))
	}

	// Burn 1 card (discard, don't store)
	// Deal 3 cards to board
	h.BoardCards = append(h.BoardCards, h.Deck[1], h.Deck[2], h.Deck[3])

	// Remove burnt card and dealt cards from deck
	h.Deck = h.Deck[4:]

	return nil
}

// DealTurn deals the turn (1 community card) after burning 1 card
// Burn card is discarded (not stored)
// Returns error if deck has insufficient cards (need at least 2: 1 burn + 1 turn)
func (h *Hand) DealTurn() error {
	// Check if we have enough cards in deck (1 burn + 1 turn = 2 total)
	if len(h.Deck) < 2 {
		return fmt.Errorf("insufficient cards in deck: have %d, need 2", len(h.Deck))
	}

	// Burn 1 card (discard, don't store)
	// Deal 1 card to board
	h.BoardCards = append(h.BoardCards, h.Deck[1])

	// Remove burnt card and dealt card from deck
	h.Deck = h.Deck[2:]

	return nil
}

// DealRiver deals the river (1 community card) after burning 1 card
// Burn card is discarded (not stored)
// Returns error if deck has insufficient cards (need at least 2: 1 burn + 1 river)
func (h *Hand) DealRiver() error {
	// Check if we have enough cards in deck (1 burn + 1 river = 2 total)
	if len(h.Deck) < 2 {
		return fmt.Errorf("insufficient cards in deck: have %d, need 2", len(h.Deck))
	}

	// Burn 1 card (discard, don't store)
	// Deal 1 card to board
	h.BoardCards = append(h.BoardCards, h.Deck[1])

	// Remove burnt card and dealt card from deck
	h.Deck = h.Deck[2:]

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

	// Blind amounts
	const smallBlind = 10
	const bigBlind = 20

	// Step 3: Create new hand and deck with action state initialized
	hand := &Hand{
		DealerSeat:     dealerSeat,
		SmallBlindSeat: sbSeat,
		BigBlindSeat:   bbSeat,
		Pot:            0,
		Deck:           NewDeck(),
		HoleCards:      make(map[int][]Card),
		BoardCards:     []Card{},
		Street:         "preflop",
		CurrentBet:     bigBlind,
		PlayerBets:     make(map[int]int),
		FoldedPlayers:  make(map[int]bool),
		ActedPlayers:   make(map[int]bool),
		LastRaise:      bigBlind,
	}

	// Step 4: Shuffle the deck
	err = ShuffleDeck(hand.Deck)
	if err != nil {
		t.mu.Unlock()
		return fmt.Errorf("failed to shuffle deck: %w", err)
	}

	// Step 5: Post blinds (handle all-in if necessary)
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

	// Update PlayerBets to track blinds posted
	hand.PlayerBets[sbSeat] = sbPosted
	hand.PlayerBets[bbSeat] = bbPosted

	// Step 6: Deal hole cards to all active players
	err = hand.DealHoleCards(t.Seats)
	if err != nil {
		t.mu.Unlock()
		return fmt.Errorf("failed to deal hole cards: %w", err)
	}

	// Step 6a: Set the first actor (who acts first preflop)
	firstActor := hand.GetFirstActor(t.Seats)
	hand.CurrentActor = &firstActor

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

		// Broadcast table state to sync card counts for all clients
		// This ensures players see card backs for opponents
		t.Server.broadcastTableState(t.ID, nil)

		// Broadcast the first action_request to the initial actor
		t.mu.RLock()
		hasCurrentActor := t.CurrentHand != nil && t.CurrentHand.CurrentActor != nil
		var seatIndex int
		var validActions []string
		var callAmount, currentBet, pot int
		if hasCurrentActor {
			seatIndex = *t.CurrentHand.CurrentActor
			playerStack := t.Seats[seatIndex].Stack
			validActions = t.CurrentHand.GetValidActions(seatIndex, playerStack, t.Seats)
			callAmount = t.CurrentHand.GetCallAmount(seatIndex)
			currentBet = t.CurrentHand.CurrentBet
			pot = t.CurrentHand.Pot
		}
		t.mu.RUnlock()

		if hasCurrentActor {
			err = t.Server.BroadcastActionRequest(
				t.ID,
				seatIndex,
				validActions,
				callAmount,
				currentBet,
				pot,
			)
			if err != nil {
				t.Server.logger.Warn("failed to broadcast first action_request", "error", err)
			}
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

// GetFirstActor determines who acts first preflop
// - Heads-up (2 active players): dealer acts first (dealer is small blind)
// - Multi-player (3+ active players): first seat after BB acts first (UTG position)
// Returns the seat number of the player who acts first
func (h *Hand) GetFirstActor(seats [6]Seat) int {
	// Count active players
	activeCount := 0
	activeSeats := []int{}
	for i := 0; i < 6; i++ {
		if seats[i].Status == "active" {
			activeCount++
			activeSeats = append(activeSeats, i)
		}
	}

	// Heads-up: dealer (small blind) acts first
	if activeCount == 2 {
		// Verify dealer is actually in the active seats list
		for _, seat := range activeSeats {
			if seat == h.DealerSeat {
				return h.DealerSeat
			}
		}
		// Fallback if dealer somehow not active
		return activeSeats[0]
	}

	// Multi-player: find first active player after BB
	// Find index of BB in activeSeats
	bbIndex := -1
	for i, seat := range activeSeats {
		if seat == h.BigBlindSeat {
			bbIndex = i
			break
		}
	}

	// Defensive check: if BB not found in activeSeats, fallback to first active player
	if bbIndex == -1 {
		// This should never happen if hand setup is correct, but defend against it
		return activeSeats[0]
	}

	// First to act is next active player after BB
	firstActorIndex := (bbIndex + 1) % len(activeSeats)
	return activeSeats[firstActorIndex]
}

// GetNextActiveSeat returns the next active (non-folded) player after fromSeat
// - Skips folded players
// - Wraps around from seat 5 to seat 0
// - If fromSeat is not in the active list, finds the next seat number greater than fromSeat
// - If no seat is found after fromSeat, wraps around to find the first seat
// - Returns nil if all other active players have folded (only one player left)
func (h *Hand) GetNextActiveSeat(fromSeat int, seats [6]Seat) *int {
	// Collect all active (not folded) seats
	activeSeatsList := []int{}
	for i := 0; i < 6; i++ {
		if seats[i].Status == "active" && !h.FoldedPlayers[i] {
			activeSeatsList = append(activeSeatsList, i)
		}
	}

	// If 0 or 1 active seats remain, return nil (betting round should be over)
	if len(activeSeatsList) <= 1 {
		return nil
	}

	// Find current position in activeSeatsList
	currentIndex := -1
	for i, seat := range activeSeatsList {
		if seat == fromSeat {
			currentIndex = i
			break
		}
	}

	// If fromSeat not in active list, find closest one after it
	if currentIndex == -1 {
		for i, seat := range activeSeatsList {
			if seat > fromSeat {
				nextSeat := activeSeatsList[i]
				return &nextSeat
			}
		}
		// If not found after, wrap around to first
		nextSeat := activeSeatsList[0]
		return &nextSeat
	}

	// Get next seat (with wrap-around)
	nextIndex := (currentIndex + 1) % len(activeSeatsList)
	nextSeat := activeSeatsList[nextIndex]
	return &nextSeat
}

// GetCallAmount returns the amount a player needs to call to match the current bet
// - Returns 0 if CurrentBet <= PlayerBet (player has already matched or exceeded the bet)
// - Returns CurrentBet - PlayerBet otherwise (amount needed to match)
func (h *Hand) GetCallAmount(seatIndex int) int {
	// Initialize maps if needed
	if h.PlayerBets == nil {
		h.PlayerBets = make(map[int]int)
	}

	playerBet := h.PlayerBets[seatIndex]
	amountToCall := h.CurrentBet - playerBet

	if amountToCall < 0 {
		amountToCall = 0
	}

	return amountToCall
}

// GetValidActions returns the list of valid actions for a player
// - If player must match a current bet: ["call", "fold", "raise"] (if enough chips to raise min)
// - If player has matched the current bet: ["check", "fold"]
// Raise is included when:
// - Player is behind (callAmount > 0) AND
// - Player has enough chips to call + raise minimum amount
func (h *Hand) GetValidActions(seatIndex int, playerStack int, seats [6]Seat) []string {
	callAmount := h.GetCallAmount(seatIndex)

	if callAmount > 0 {
		// Player is behind, must call to continue
		// Check if they can also raise
		minRaise := h.GetMinRaise()
		// Player needs: callAmount (to match current bet) + minRaise (for the raise)
		// But actually, to raise minimum, they need to put out callAmount + minRaise total
		// Wait, that's not right. MinRaise is the increment AFTER calling.
		// Actually, minRaise is already CurrentBet + LastRaise
		// So to raise minimum, player needs to put out minRaise total
		// But they're currently only bet PlayerBets[seatIndex]
		// So they need minRaise - PlayerBets[seatIndex] more chips

		chipsNeeded := minRaise - h.PlayerBets[seatIndex]
		if chipsNeeded <= playerStack {
			// Player can raise
			return []string{"fold", "call", "raise"}
		}
		// Player cannot raise
		return []string{"call", "fold"}
	}

	// Player has matched current bet, can check or fold
	return []string{"check", "fold"}
}

// GetMinRaise returns the minimum raise amount (what players must raise to at minimum)
// Minimum raise = CurrentBet + LastRaise
// Example: If BB=20, then min-raise to 40 (20 + 20)
// After raise to 60, min-raise becomes 100 (60 + 40)
func (h *Hand) GetMinRaise() int {
	return h.CurrentBet + h.LastRaise
}

// GetMaxRaise returns the maximum valid raise amount to prevent side pots
// Returns the minimum of:
// - Player's remaining stack
// - Smallest opponent's stack (to prevent side pots)
// This prevents creating side pots by ensuring the largest raise doesn't exceed
// what all other players can match
func (t *Table) GetMaxRaise(seatIndex int, seats [6]Seat) int {
	playerStack := seats[seatIndex].Stack

	// Find smallest opponent stack among active players
	smallestOpponentStack := -1
	for i := 0; i < 6; i++ {
		if i != seatIndex && seats[i].Status == "active" {
			if smallestOpponentStack == -1 || seats[i].Stack < smallestOpponentStack {
				smallestOpponentStack = seats[i].Stack
			}
		}
	}

	// If no opponents found (shouldn't happen in valid game), return player's stack
	if smallestOpponentStack == -1 {
		return playerStack
	}

	// Return minimum of player's stack and smallest opponent's stack
	if playerStack < smallestOpponentStack {
		return playerStack
	}
	return smallestOpponentStack
}

// ValidateRaise checks if a raise amount is valid
// Returns nil if the raise is valid, error otherwise
// Rules:
// - If raiseAmount equals playerStack (all-in): always valid, return nil
// - If raiseAmount < GetMinRaise(): return error "raise amount below minimum"
// - If raiseAmount > GetMaxRaise(): return error "raise would create side pot"
// - Otherwise: return nil
func (h *Hand) ValidateRaise(seatIndex, raiseAmount, playerStack int, seats [6]Seat) error {
	// Check if this is all-in (raiseAmount equals playerStack)
	if raiseAmount == playerStack {
		// All-in is always valid, even below minimum
		return nil
	}

	// Check minimum raise
	minRaise := h.GetMinRaise()
	if raiseAmount < minRaise {
		return fmt.Errorf("raise amount below minimum")
	}

	// Check maximum raise (prevent side pots)
	// We need to get a Table reference to call GetMaxRaise, but we only have Hand
	// We'll need to refactor this - let me check the calling pattern
	// Actually, ValidateRaise should work with the hand's context
	// We need the table to call GetMaxRaise, so let's make this work differently

	// For now, let's compute max raise inline here
	// Find smallest opponent stack among active players
	smallestOpponentStack := -1
	for i := 0; i < 6; i++ {
		if i != seatIndex && seats[i].Status == "active" {
			if smallestOpponentStack == -1 || seats[i].Stack < smallestOpponentStack {
				smallestOpponentStack = seats[i].Stack
			}
		}
	}

	// If no opponents, no side pot check needed
	if smallestOpponentStack == -1 {
		return nil
	}

	// Maximum allowed raise is the minimum of player's stack and smallest opponent's stack
	maxRaise := playerStack
	if smallestOpponentStack < playerStack {
		maxRaise = smallestOpponentStack
	}

	if raiseAmount > maxRaise {
		return fmt.Errorf("raise would create side pot")
	}

	return nil
}

// ProcessAction processes a player action (fold, check, call, or raise)
// - "fold": marks player as folded (no pot/stack changes)
// - "check": valid only when bet is matched; marks player as acted (no pot/stack changes)
// - "call": moves chips from stack to pot to match current bet; handles all-in
// - "raise": validates and processes raise, updating CurrentBet, LastRaise, PlayerBets, and Pot
// The amount parameter is required for "raise" action and ignored for other actions.
// The caller is responsible for updating the player's stack in the table's seats after calling this.
// Returns the number of chips moved (for updating player stack), or error if action is invalid
func (h *Hand) ProcessAction(seatIndex int, action string, playerStack int, amount ...int) (int, error) {
	// Initialize maps if needed
	if h.FoldedPlayers == nil {
		h.FoldedPlayers = make(map[int]bool)
	}
	if h.PlayerBets == nil {
		h.PlayerBets = make(map[int]int)
	}
	if h.ActedPlayers == nil {
		h.ActedPlayers = make(map[int]bool)
	}

	switch action {
	case "fold":
		// Mark player as folded
		h.FoldedPlayers[seatIndex] = true
		h.ActedPlayers[seatIndex] = true
		return 0, nil

	case "check":
		// Check is only valid when player has matched the current bet
		callAmount := h.GetCallAmount(seatIndex)
		if callAmount > 0 {
			return 0, fmt.Errorf("cannot check when behind current bet (need to call %d)", callAmount)
		}

		// Mark player as acted
		h.ActedPlayers[seatIndex] = true
		return 0, nil

	case "call":
		// Get amount needed to match current bet
		callAmount := h.GetCallAmount(seatIndex)

		// Amount to actually move (min of what they need to call and what they have)
		chipsToBet := callAmount
		if chipsToBet > playerStack {
			// Go all-in with available chips
			chipsToBet = playerStack
		}

		// Update pot
		h.Pot += chipsToBet

		// Update player's bet for this round
		h.PlayerBets[seatIndex] += chipsToBet

		// Mark player as acted
		h.ActedPlayers[seatIndex] = true
		return chipsToBet, nil

	case "raise":
		// Extract raise amount from variadic parameter
		if len(amount) == 0 {
			return 0, fmt.Errorf("raise action requires amount parameter")
		}
		raiseAmount := amount[0]

		// Validate raise amount (this will be called with current table seats)
		// For now, we'll do inline validation
		// Check if this is all-in (raiseAmount equals playerStack)
		if raiseAmount != playerStack {
			// Not all-in, so validate min/max bounds
			minRaise := h.GetMinRaise()
			if raiseAmount < minRaise {
				return 0, fmt.Errorf("raise amount below minimum")
			}
			// Note: Max raise validation would need table context, handled by caller
		}

		// Calculate chips to move (raise amount minus what was already bet)
		currentPlayerBet := h.PlayerBets[seatIndex]
		chipsToBet := raiseAmount - currentPlayerBet

		// Sanity check: don't exceed player's stack
		if chipsToBet > playerStack {
			return 0, fmt.Errorf("raise exceeds player stack")
		}

		// Update CurrentBet to this raise amount
		previousBet := h.CurrentBet
		h.CurrentBet = raiseAmount

		// Update LastRaise (increment from previous bet)
		h.LastRaise = raiseAmount - previousBet

		// Update pot with chips moved
		h.Pot += chipsToBet

		// Update player's total bet for this round
		h.PlayerBets[seatIndex] = raiseAmount

		// Mark player as acted
		h.ActedPlayers[seatIndex] = true

		return chipsToBet, nil

	default:
		return 0, fmt.Errorf("invalid action: %s", action)
	}
}

// IsBettingRoundComplete determines if the current betting round is over
// Returns true if:
// - Exactly one active player remains (all others have folded), OR
// - All active players have acted AND all active players have matched the current bet
// Returns false otherwise
func (h *Hand) IsBettingRoundComplete(seats [6]Seat) bool {
	// Initialize maps if needed
	if h.FoldedPlayers == nil {
		h.FoldedPlayers = make(map[int]bool)
	}
	if h.ActedPlayers == nil {
		h.ActedPlayers = make(map[int]bool)
	}
	if h.PlayerBets == nil {
		h.PlayerBets = make(map[int]int)
	}

	// Count active players and folded players
	activeCount := 0
	activePlayers := []int{}
	for i := 0; i < 6; i++ {
		if seats[i].Status == "active" {
			activeCount++
			activePlayers = append(activePlayers, i)
		}
	}

	// Count non-folded active players
	nonFoldedCount := 0
	for _, seatNum := range activePlayers {
		if !h.FoldedPlayers[seatNum] {
			nonFoldedCount++
		}
	}

	// If only one player remains (all others folded), round is complete
	if nonFoldedCount <= 1 {
		return true
	}

	// Check if all active (non-folded) players have acted
	for _, seatNum := range activePlayers {
		if !h.FoldedPlayers[seatNum] && !h.ActedPlayers[seatNum] {
			// This player hasn't acted yet, round not complete
			return false
		}
	}

	// Check if all active (non-folded) players have matched the current bet
	for _, seatNum := range activePlayers {
		if !h.FoldedPlayers[seatNum] {
			playerBet := h.PlayerBets[seatNum]
			if playerBet != h.CurrentBet {
				// This player hasn't matched the current bet
				return false
			}
		}
	}

	// All conditions met: all non-folded players have acted and matched the bet
	return true
}

// AdvanceAction moves the current actor to the next active player
// Returns the seat number of the next actor, or nil if no next actor exists (only one player left)
// Returns error if CurrentActor is nil (no current actor set)
func (h *Hand) AdvanceAction(seats [6]Seat) (*int, error) {
	if h.CurrentActor == nil {
		return nil, fmt.Errorf("current actor is nil")
	}

	// Get the next active seat after current actor
	nextSeat := h.GetNextActiveSeat(*h.CurrentActor, seats)
	return nextSeat, nil
}

// AdvanceStreet moves the hand to the next street and resets betting state
// Streets: preflop -> flop -> turn -> river
// Resets CurrentBet, LastRaise, and ActedPlayers for the new street
func (h *Hand) AdvanceStreet() {
	switch h.Street {
	case "preflop":
		h.Street = "flop"
	case "flop":
		h.Street = "turn"
	case "turn":
		h.Street = "river"
	case "river":
		// Hand is over (no advance from river)
		return
	}

	// Reset betting state for new street
	h.CurrentBet = 0
	h.LastRaise = 0
	h.PlayerBets = make(map[int]int)
	h.ActedPlayers = make(map[int]bool)
	h.CurrentActor = nil
}

// AdvanceToNextStreet advances the hand to the next street by dealing board cards and resetting betting state
// Streets: preflop -> flop -> turn -> river
// - On preflop: deals flop (3 cards)
// - On flop: deals turn (1 card)
// - On turn: deals river (1 card)
// - On river: no advancement (hand is complete)
// Returns error if dealing fails (e.g., insufficient cards in deck)
// This method combines board card dealing with betting state reset
func (h *Hand) AdvanceToNextStreet() error {
	// Deal board cards based on current street before advancing
	var err error
	switch h.Street {
	case "preflop":
		// Deal flop (3 cards)
		err = h.DealFlop()
		if err != nil {
			return err
		}
	case "flop":
		// Deal turn (1 card)
		err = h.DealTurn()
		if err != nil {
			return err
		}
	case "turn":
		// Deal river (1 card)
		err = h.DealRiver()
		if err != nil {
			return err
		}
	case "river":
		// Hand is complete, no more streets to advance to
		return nil
	}

	// Advance street and reset betting state
	h.AdvanceStreet()

	return nil
}
