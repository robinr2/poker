package server

import (
	"testing"
)

func TestEvaluateHand_RoyalFlush(t *testing.T) {
	// Royal Flush: A-K-Q-J-T all same suit
	holeCards := []Card{
		{Rank: "A", Suit: "s"},
		{Rank: "K", Suit: "s"},
	}
	boardCards := []Card{
		{Rank: "Q", Suit: "s"},
		{Rank: "J", Suit: "s"},
		{Rank: "T", Suit: "s"},
		{Rank: "9", Suit: "h"},
		{Rank: "8", Suit: "h"},
	}

	result := EvaluateHand(holeCards, boardCards)

	if result.Rank != 9 {
		t.Errorf("expected rank 9 (royal flush), got %d", result.Rank)
	}

	// Kickers should be [14, 13, 12, 11, 10] for A-K-Q-J-T
	expectedKickers := []int{14, 13, 12, 11, 10}
	if !sliceEqual(result.Kickers, expectedKickers) {
		t.Errorf("expected kickers %v, got %v", expectedKickers, result.Kickers)
	}
}

func TestEvaluateHand_StraightFlush(t *testing.T) {
	// Straight Flush: 9-8-7-6-5 all diamonds
	holeCards := []Card{
		{Rank: "9", Suit: "d"},
		{Rank: "8", Suit: "d"},
	}
	boardCards := []Card{
		{Rank: "7", Suit: "d"},
		{Rank: "6", Suit: "d"},
		{Rank: "5", Suit: "d"},
		{Rank: "K", Suit: "h"},
		{Rank: "Q", Suit: "h"},
	}

	result := EvaluateHand(holeCards, boardCards)

	if result.Rank != 8 {
		t.Errorf("expected rank 8 (straight flush), got %d", result.Rank)
	}

	// Kickers should be the top card of straight [9, 8, 7, 6, 5]
	expectedKickers := []int{9, 8, 7, 6, 5}
	if !sliceEqual(result.Kickers, expectedKickers) {
		t.Errorf("expected kickers %v, got %v", expectedKickers, result.Kickers)
	}
}

func TestEvaluateHand_FourOfAKind(t *testing.T) {
	// Four of a Kind: 4 Kings
	holeCards := []Card{
		{Rank: "K", Suit: "s"},
		{Rank: "K", Suit: "h"},
	}
	boardCards := []Card{
		{Rank: "K", Suit: "d"},
		{Rank: "K", Suit: "c"},
		{Rank: "Q", Suit: "s"},
		{Rank: "2", Suit: "h"},
		{Rank: "3", Suit: "h"},
	}

	result := EvaluateHand(holeCards, boardCards)

	if result.Rank != 7 {
		t.Errorf("expected rank 7 (four of a kind), got %d", result.Rank)
	}

	// Kickers should be [13, 12] - four Kings and Queen kicker
	expectedKickers := []int{13, 12}
	if !sliceEqual(result.Kickers, expectedKickers) {
		t.Errorf("expected kickers %v, got %v", expectedKickers, result.Kickers)
	}
}

func TestEvaluateHand_FullHouse(t *testing.T) {
	// Full House: 3 Aces and 2 Kings
	holeCards := []Card{
		{Rank: "A", Suit: "s"},
		{Rank: "A", Suit: "h"},
	}
	boardCards := []Card{
		{Rank: "A", Suit: "d"},
		{Rank: "K", Suit: "c"},
		{Rank: "K", Suit: "s"},
		{Rank: "2", Suit: "h"},
		{Rank: "3", Suit: "h"},
	}

	result := EvaluateHand(holeCards, boardCards)

	if result.Rank != 6 {
		t.Errorf("expected rank 6 (full house), got %d", result.Rank)
	}

	// Kickers should be [14, 13] - three Aces and pair of Kings
	expectedKickers := []int{14, 13}
	if !sliceEqual(result.Kickers, expectedKickers) {
		t.Errorf("expected kickers %v, got %v", expectedKickers, result.Kickers)
	}
}

func TestEvaluateHand_Flush(t *testing.T) {
	// Flush: 5 spades (A-K-Q-J-9)
	holeCards := []Card{
		{Rank: "A", Suit: "s"},
		{Rank: "K", Suit: "s"},
	}
	boardCards := []Card{
		{Rank: "Q", Suit: "s"},
		{Rank: "J", Suit: "s"},
		{Rank: "9", Suit: "s"},
		{Rank: "2", Suit: "h"},
		{Rank: "3", Suit: "h"},
	}

	result := EvaluateHand(holeCards, boardCards)

	if result.Rank != 5 {
		t.Errorf("expected rank 5 (flush), got %d", result.Rank)
	}

	// Kickers should be [14, 13, 12, 11, 9] - all flush cards in descending order
	expectedKickers := []int{14, 13, 12, 11, 9}
	if !sliceEqual(result.Kickers, expectedKickers) {
		t.Errorf("expected kickers %v, got %v", expectedKickers, result.Kickers)
	}
}

func TestEvaluateHand_Straight(t *testing.T) {
	// Straight: A-K-Q-J-T (high straight)
	holeCards := []Card{
		{Rank: "A", Suit: "s"},
		{Rank: "K", Suit: "h"},
	}
	boardCards := []Card{
		{Rank: "Q", Suit: "d"},
		{Rank: "J", Suit: "c"},
		{Rank: "T", Suit: "s"},
		{Rank: "2", Suit: "h"},
		{Rank: "3", Suit: "h"},
	}

	result := EvaluateHand(holeCards, boardCards)

	if result.Rank != 4 {
		t.Errorf("expected rank 4 (straight), got %d", result.Rank)
	}

	// Kickers should be [14, 13, 12, 11, 10] - just the top card matters for ranking
	expectedKickers := []int{14, 13, 12, 11, 10}
	if !sliceEqual(result.Kickers, expectedKickers) {
		t.Errorf("expected kickers %v, got %v", expectedKickers, result.Kickers)
	}
}

func TestEvaluateHand_ThreeOfAKind(t *testing.T) {
	// Three of a Kind: 3 Jacks
	holeCards := []Card{
		{Rank: "J", Suit: "s"},
		{Rank: "J", Suit: "h"},
	}
	boardCards := []Card{
		{Rank: "J", Suit: "d"},
		{Rank: "K", Suit: "c"},
		{Rank: "Q", Suit: "s"},
		{Rank: "2", Suit: "h"},
		{Rank: "3", Suit: "h"},
	}

	result := EvaluateHand(holeCards, boardCards)

	if result.Rank != 3 {
		t.Errorf("expected rank 3 (three of a kind), got %d", result.Rank)
	}

	// Kickers should be [11, 13, 12] - Jack triplet with K and Q kickers
	expectedKickers := []int{11, 13, 12}
	if !sliceEqual(result.Kickers, expectedKickers) {
		t.Errorf("expected kickers %v, got %v", expectedKickers, result.Kickers)
	}
}

func TestEvaluateHand_TwoPair(t *testing.T) {
	// Two Pair: Kings and Nines
	holeCards := []Card{
		{Rank: "K", Suit: "s"},
		{Rank: "K", Suit: "h"},
	}
	boardCards := []Card{
		{Rank: "9", Suit: "d"},
		{Rank: "9", Suit: "c"},
		{Rank: "Q", Suit: "s"},
		{Rank: "2", Suit: "h"},
		{Rank: "3", Suit: "h"},
	}

	result := EvaluateHand(holeCards, boardCards)

	if result.Rank != 2 {
		t.Errorf("expected rank 2 (two pair), got %d", result.Rank)
	}

	// Kickers should be [13, 9, 12] - high pair (K), low pair (9), and Queen kicker
	expectedKickers := []int{13, 9, 12}
	if !sliceEqual(result.Kickers, expectedKickers) {
		t.Errorf("expected kickers %v, got %v", expectedKickers, result.Kickers)
	}
}

func TestEvaluateHand_OnePair(t *testing.T) {
	// One Pair: Pair of Tens
	holeCards := []Card{
		{Rank: "T", Suit: "s"},
		{Rank: "T", Suit: "h"},
	}
	boardCards := []Card{
		{Rank: "K", Suit: "d"},
		{Rank: "Q", Suit: "c"},
		{Rank: "J", Suit: "s"},
		{Rank: "2", Suit: "h"},
		{Rank: "3", Suit: "h"},
	}

	result := EvaluateHand(holeCards, boardCards)

	if result.Rank != 1 {
		t.Errorf("expected rank 1 (one pair), got %d", result.Rank)
	}

	// Kickers should be [10, 13, 12, 11] - Ten pair with K, Q, J kickers
	expectedKickers := []int{10, 13, 12, 11}
	if !sliceEqual(result.Kickers, expectedKickers) {
		t.Errorf("expected kickers %v, got %v", expectedKickers, result.Kickers)
	}
}

func TestEvaluateHand_HighCard(t *testing.T) {
	// High Card: A-K-Q-J-9 (no other hand)
	holeCards := []Card{
		{Rank: "A", Suit: "s"},
		{Rank: "K", Suit: "h"},
	}
	boardCards := []Card{
		{Rank: "Q", Suit: "d"},
		{Rank: "J", Suit: "c"},
		{Rank: "9", Suit: "s"},
		{Rank: "2", Suit: "h"},
		{Rank: "3", Suit: "h"},
	}

	result := EvaluateHand(holeCards, boardCards)

	if result.Rank != 0 {
		t.Errorf("expected rank 0 (high card), got %d", result.Rank)
	}

	// Kickers should be [14, 13, 12, 11, 9] - all cards in descending order
	expectedKickers := []int{14, 13, 12, 11, 9}
	if !sliceEqual(result.Kickers, expectedKickers) {
		t.Errorf("expected kickers %v, got %v", expectedKickers, result.Kickers)
	}
}

func TestCompareHands_DifferentRanks(t *testing.T) {
	// Compare Two Pair vs One Pair
	pair := HandRank{Rank: 1, Kickers: []int{10, 13, 12, 11}}
	twoPair := HandRank{Rank: 2, Kickers: []int{13, 9, 12}}

	result := CompareHands(twoPair, pair)

	if result <= 0 {
		t.Errorf("expected two pair to beat one pair (result > 0), got %d", result)
	}

	result = CompareHands(pair, twoPair)

	if result >= 0 {
		t.Errorf("expected one pair to lose to two pair (result < 0), got %d", result)
	}
}

func TestCompareHands_SameRank_Kickers(t *testing.T) {
	// Compare two one-pair hands with different kickers
	pair1 := HandRank{Rank: 1, Kickers: []int{10, 13, 12, 11}}
	pair2 := HandRank{Rank: 1, Kickers: []int{10, 13, 12, 9}}

	result := CompareHands(pair1, pair2)

	if result <= 0 {
		t.Errorf("expected pair1 to beat pair2 based on kickers (result > 0), got %d", result)
	}

	result = CompareHands(pair2, pair1)

	if result >= 0 {
		t.Errorf("expected pair2 to lose to pair1 based on kickers (result < 0), got %d", result)
	}
}

func TestCompareHands_Tie(t *testing.T) {
	// Compare identical hands
	pair1 := HandRank{Rank: 1, Kickers: []int{10, 13, 12, 11}}
	pair2 := HandRank{Rank: 1, Kickers: []int{10, 13, 12, 11}}

	result := CompareHands(pair1, pair2)

	if result != 0 {
		t.Errorf("expected tie (result == 0), got %d", result)
	}
}

func TestEvaluateHand_WheelStraight(t *testing.T) {
	// Wheel Straight (A-2-3-4-5): Ace acts as low card
	holeCards := []Card{
		{Rank: "A", Suit: "s"},
		{Rank: "2", Suit: "h"},
	}
	boardCards := []Card{
		{Rank: "3", Suit: "d"},
		{Rank: "4", Suit: "c"},
		{Rank: "5", Suit: "s"},
		{Rank: "K", Suit: "h"},
		{Rank: "Q", Suit: "h"},
	}

	result := EvaluateHand(holeCards, boardCards)

	if result.Rank != 4 {
		t.Errorf("expected rank 4 (straight), got %d", result.Rank)
	}

	// Kickers should be [5, 4, 3, 2, 1] - wheel straight where Ace is low (1)
	expectedKickers := []int{5, 4, 3, 2, 1}
	if !sliceEqual(result.Kickers, expectedKickers) {
		t.Errorf("expected kickers %v for wheel straight, got %v", expectedKickers, result.Kickers)
	}
}

func TestEvaluateHand_AllSevenCards(t *testing.T) {
	// Test picking best 5 from all 7 cards
	// 2 hole + 5 board = 7 cards
	// Board has royal flush potential (A-K-Q-J-T all hearts)
	holeCards := []Card{
		{Rank: "A", Suit: "h"},
		{Rank: "K", Suit: "h"},
	}
	boardCards := []Card{
		{Rank: "Q", Suit: "h"},
		{Rank: "J", Suit: "h"},
		{Rank: "T", Suit: "h"},
		{Rank: "2", Suit: "c"},
		{Rank: "3", Suit: "d"},
	}

	result := EvaluateHand(holeCards, boardCards)

	// Should recognize royal flush in hearts (A-K-Q-J-T all hearts)
	if result.Rank != 9 {
		t.Errorf("expected rank 9 (royal flush), got %d", result.Rank)
	}

	expectedKickers := []int{14, 13, 12, 11, 10}
	if !sliceEqual(result.Kickers, expectedKickers) {
		t.Errorf("expected kickers %v, got %v", expectedKickers, result.Kickers)
	}
}

// Helper function for test assertions
func sliceEqual(a, b []int) bool {
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
