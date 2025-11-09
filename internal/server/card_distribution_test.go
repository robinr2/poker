package server

import (
	"log/slog"
	"testing"
)

// TestCardDistribution verifies that all suits appear in dealt cards over multiple hands
// This test checks for bias in card dealing by running multiple hand simulations
func TestCardDistribution(t *testing.T) {
	// Create a table with 3 players
	logger := slog.Default()
	server := &Server{
		logger: logger,
		tables: [4]*Table{},
	}
	table := NewTable("test-table", "Test Table", server)

	// Add 3 players
	token1 := "player1"
	token2 := "player2"
	token3 := "player3"

	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"
	table.Seats[0].Stack = 10000

	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"
	table.Seats[1].Stack = 10000

	table.Seats[2].Token = &token3
	table.Seats[2].Status = "active"
	table.Seats[2].Stack = 10000

	// Track suits dealt across 100 hands
	suitCounts := map[string]int{
		"s": 0, // spades
		"h": 0, // hearts
		"d": 0, // diamonds
		"c": 0, // clubs
	}

	totalCards := 0
	handsToRun := 100

	for i := 0; i < handsToRun; i++ {
		// Create and shuffle deck
		deck := NewDeck()
		err := ShuffleDeck(deck)
		if err != nil {
			t.Fatalf("hand %d: failed to shuffle deck: %v", i, err)
		}

		// Deal 6 cards (2 to each of 3 players)
		for j := 0; j < 6; j++ {
			card := deck[j]
			suitCounts[card.Suit]++
			totalCards++

			// Log first hand to see what's being dealt
			if i == 0 {
				t.Logf("Hand 0, Card %d: %s%s", j, card.Rank, card.Suit)
			}
		}
	}

	// Verify that all suits appeared
	t.Logf("Card distribution over %d hands (%d total cards):", handsToRun, totalCards)
	t.Logf("Spades (s):   %d (%.1f%%)", suitCounts["s"], float64(suitCounts["s"])/float64(totalCards)*100)
	t.Logf("Hearts (h):   %d (%.1f%%)", suitCounts["h"], float64(suitCounts["h"])/float64(totalCards)*100)
	t.Logf("Diamonds (d): %d (%.1f%%)", suitCounts["d"], float64(suitCounts["d"])/float64(totalCards)*100)
	t.Logf("Clubs (c):    %d (%.1f%%)", suitCounts["c"], float64(suitCounts["c"])/float64(totalCards)*100)

	// Check that each suit appeared at least once
	for suit, count := range suitCounts {
		if count == 0 {
			t.Errorf("Suit %s never appeared in %d hands!", suit, handsToRun)
		}
	}

	// Check for extreme bias (any suit < 10% or > 40% would be suspicious with 100 hands)
	expectedPct := 25.0
	tolerance := 10.0 // Allow +/- 10% deviation (15% to 35%)

	for suit, count := range suitCounts {
		pct := float64(count) / float64(totalCards) * 100
		if pct < expectedPct-tolerance || pct > expectedPct+tolerance {
			t.Errorf("Suit %s distribution is suspicious: %.1f%% (expected ~25%%, tolerance +/- %.0f%%)",
				suit, pct, tolerance)
		}
	}
}

// TestDealHoleCardsDistribution verifies that DealHoleCards produces varied suits
func TestDealHoleCardsDistribution(t *testing.T) {
	// Create a mock table
	logger := slog.Default()
	server := &Server{
		logger: logger,
		tables: [4]*Table{},
	}
	table := NewTable("test-table", "Test Table", server)

	// Add 3 players
	token1 := "player1"
	token2 := "player2"
	token3 := "player3"

	table.Seats[0].Token = &token1
	table.Seats[0].Status = "active"

	table.Seats[1].Token = &token2
	table.Seats[1].Status = "active"

	table.Seats[2].Token = &token3
	table.Seats[2].Status = "active"

	// Track suits in dealt hole cards
	suitCounts := map[string]int{
		"s": 0,
		"h": 0,
		"d": 0,
		"c": 0,
	}

	handsToRun := 50

	for i := 0; i < handsToRun; i++ {
		// Create hand with new deck
		hand := &Hand{
			Deck:      NewDeck(),
			HoleCards: make(map[int][]Card),
		}

		// Shuffle deck
		err := ShuffleDeck(hand.Deck)
		if err != nil {
			t.Fatalf("hand %d: failed to shuffle: %v", i, err)
		}

		// Deal hole cards
		err = hand.DealHoleCards(table.Seats)
		if err != nil {
			t.Fatalf("hand %d: failed to deal: %v", i, err)
		}

		// Count suits in dealt cards
		for seatNum, cards := range hand.HoleCards {
			for cardIdx, card := range cards {
				suitCounts[card.Suit]++

				// Log first hand
				if i == 0 {
					t.Logf("Hand 0, Seat %d, Card %d: %s%s", seatNum, cardIdx, card.Rank, card.Suit)
				}
			}
		}
	}

	totalCards := handsToRun * 6 // 3 players * 2 cards * 50 hands

	t.Logf("Hole cards distribution over %d hands (%d total cards):", handsToRun, totalCards)
	t.Logf("Spades (s):   %d (%.1f%%)", suitCounts["s"], float64(suitCounts["s"])/float64(totalCards)*100)
	t.Logf("Hearts (h):   %d (%.1f%%)", suitCounts["h"], float64(suitCounts["h"])/float64(totalCards)*100)
	t.Logf("Diamonds (d): %d (%.1f%%)", suitCounts["d"], float64(suitCounts["d"])/float64(totalCards)*100)
	t.Logf("Clubs (c):    %d (%.1f%%)", suitCounts["c"], float64(suitCounts["c"])/float64(totalCards)*100)

	// Verify all suits appeared
	for suit, count := range suitCounts {
		if count == 0 {
			t.Errorf("Suit %s never appeared in dealt hole cards!", suit)
		}
	}
}
