package server

import (
	"sort"
)

// HandRank represents the rank and kickers of a poker hand
type HandRank struct {
	Rank    int   // 0-9: HighCard, Pair, TwoPair, ThreeOfAKind, Straight, Flush, FullHouse, FourOfAKind, StraightFlush, RoyalFlush
	Kickers []int // Card ranks in order of importance for tiebreaking
}

// rankToNumeric converts a card rank string to numeric value
// A=14, K=13, Q=12, J=11, T=10, 9=9, ..., 2=2, and A=1 for wheel straight
func rankToNumeric(rank string) int {
	switch rank {
	case "A":
		return 14
	case "K":
		return 13
	case "Q":
		return 12
	case "J":
		return 11
	case "T":
		return 10
	default:
		// For "2" through "9"
		return int(rank[0]) - int('0')
	}
}

// EvaluateHand takes 2 hole cards and 5 board cards, returns the best 5-card hand
func EvaluateHand(holeCards []Card, boardCards []Card) HandRank {
	// Combine all cards
	allCards := make([]Card, len(holeCards)+len(boardCards))
	copy(allCards, holeCards)
	copy(allCards[len(holeCards):], boardCards)

	// Generate all possible 5-card combinations from 7 cards
	var bestHand HandRank
	bestHand.Rank = -1 // Initialize to invalid

	combinations := generate5CardCombinations(allCards)
	for _, combo := range combinations {
		hand := evaluateFixed5Cards(combo)
		if bestHand.Rank == -1 || compareHandRanks(hand, bestHand) > 0 {
			bestHand = hand
		}
	}

	return bestHand
}

// generate5CardCombinations generates all possible 5-card combinations from 7 cards
func generate5CardCombinations(cards []Card) [][]Card {
	var result [][]Card

	// Use bit masking: iterate through all 7-bit numbers, selecting indices where bit is 1
	for mask := 0; mask < 128; mask++ { // 2^7 = 128
		if popcount(mask) != 5 {
			continue
		}

		var combo []Card
		for i := 0; i < 7; i++ {
			if (mask & (1 << i)) != 0 {
				combo = append(combo, cards[i])
			}
		}
		result = append(result, combo)
	}

	return result
}

// popcount returns the number of set bits in a number
func popcount(n int) int {
	count := 0
	for n > 0 {
		count += n & 1
		n >>= 1
	}
	return count
}

// evaluateFixed5Cards evaluates exactly 5 cards and returns their hand rank
func evaluateFixed5Cards(cards []Card) HandRank {
	// Check hands from highest to lowest rank
	if hand := checkRoyalFlush(cards); hand.Rank >= 0 {
		return hand
	}
	if hand := checkStraightFlush(cards); hand.Rank >= 0 {
		return hand
	}
	if hand := checkFourOfAKind(cards); hand.Rank >= 0 {
		return hand
	}
	if hand := checkFullHouse(cards); hand.Rank >= 0 {
		return hand
	}
	if hand := checkFlush(cards); hand.Rank >= 0 {
		return hand
	}
	if hand := checkStraight(cards); hand.Rank >= 0 {
		return hand
	}
	if hand := checkThreeOfAKind(cards); hand.Rank >= 0 {
		return hand
	}
	if hand := checkTwoPair(cards); hand.Rank >= 0 {
		return hand
	}
	if hand := checkOnePair(cards); hand.Rank >= 0 {
		return hand
	}

	// High card
	return checkHighCard(cards)
}

// compareHandRanks returns 1 if h1 > h2, -1 if h1 < h2, 0 if equal
func compareHandRanks(h1, h2 HandRank) int {
	if h1.Rank > h2.Rank {
		return 1
	}
	if h1.Rank < h2.Rank {
		return -1
	}

	// Same rank, compare kickers
	for i := 0; i < len(h1.Kickers) && i < len(h2.Kickers); i++ {
		if h1.Kickers[i] > h2.Kickers[i] {
			return 1
		}
		if h1.Kickers[i] < h2.Kickers[i] {
			return -1
		}
	}

	return 0
}

// groupByRank groups cards by their rank and returns map of rank -> count
func groupByRank(cards []Card) map[int][]Card {
	grouped := make(map[int][]Card)
	for _, card := range cards {
		rank := rankToNumeric(card.Rank)
		grouped[rank] = append(grouped[rank], card)
	}
	return grouped
}

// isFlush checks if all 5 cards have the same suit
func isFlush(cards []Card) bool {
	if len(cards) != 5 {
		return false
	}
	suit := cards[0].Suit
	for i := 1; i < len(cards); i++ {
		if cards[i].Suit != suit {
			return false
		}
	}
	return true
}

// isStraight checks if cards form a straight
// Returns the high card of the straight (or 5 for wheel)
func isStraight(cards []Card) (bool, int) {
	if len(cards) != 5 {
		return false, 0
	}

	ranks := make([]int, len(cards))
	for i, card := range cards {
		ranks[i] = rankToNumeric(card.Rank)
	}
	sort.Ints(ranks)

	// Check for regular straight
	isStraightHand := true
	for i := 1; i < 5; i++ {
		if ranks[i] != ranks[i-1]+1 {
			isStraightHand = false
			break
		}
	}

	if isStraightHand {
		return true, ranks[4] // Return high card
	}

	// Check for wheel (A-2-3-4-5)
	if ranks[0] == 2 && ranks[1] == 3 && ranks[2] == 4 && ranks[3] == 5 && ranks[4] == 14 {
		return true, 5 // Wheel high card is 5, but Ace represents 1
	}

	return false, 0
}

// checkRoyalFlush checks for royal flush (A-K-Q-J-T all same suit)
func checkRoyalFlush(cards []Card) HandRank {
	if !isFlush(cards) {
		return HandRank{Rank: -1}
	}

	ranks := make([]int, len(cards))
	for i, card := range cards {
		ranks[i] = rankToNumeric(card.Rank)
	}
	sort.Ints(ranks)

	// Check if we have 10-11-12-13-14 (T-J-Q-K-A)
	if ranks[0] == 10 && ranks[1] == 11 && ranks[2] == 12 && ranks[3] == 13 && ranks[4] == 14 {
		return HandRank{
			Rank:    9,
			Kickers: []int{14, 13, 12, 11, 10},
		}
	}

	return HandRank{Rank: -1}
}

// checkStraightFlush checks for straight flush
func checkStraightFlush(cards []Card) HandRank {
	if !isFlush(cards) {
		return HandRank{Rank: -1}
	}

	isStraightHand, highCard := isStraight(cards)
	if !isStraightHand {
		return HandRank{Rank: -1}
	}

	// For wheel, special kickers
	if highCard == 5 {
		return HandRank{
			Rank:    8,
			Kickers: []int{5, 4, 3, 2, 1},
		}
	}

	// For regular straight, build kickers from high to low
	kickers := []int{highCard, highCard - 1, highCard - 2, highCard - 3, highCard - 4}

	return HandRank{
		Rank:    8,
		Kickers: kickers,
	}
}

// checkFourOfAKind checks for four of a kind
func checkFourOfAKind(cards []Card) HandRank {
	grouped := groupByRank(cards)

	var quadRank int
	var kicker int

	for rank, rankCards := range grouped {
		if len(rankCards) == 4 {
			quadRank = rank
		} else if len(rankCards) == 1 {
			kicker = rank
		}
	}

	if quadRank == 0 {
		return HandRank{Rank: -1}
	}

	return HandRank{
		Rank:    7,
		Kickers: []int{quadRank, kicker},
	}
}

// checkFullHouse checks for full house (3 of a kind + pair)
func checkFullHouse(cards []Card) HandRank {
	grouped := groupByRank(cards)

	var tripleRank int
	var pairRank int

	for rank, rankCards := range grouped {
		if len(rankCards) == 3 {
			tripleRank = rank
		} else if len(rankCards) == 2 {
			pairRank = rank
		}
	}

	if tripleRank == 0 || pairRank == 0 {
		return HandRank{Rank: -1}
	}

	return HandRank{
		Rank:    6,
		Kickers: []int{tripleRank, pairRank},
	}
}

// checkFlush checks for flush
func checkFlush(cards []Card) HandRank {
	if !isFlush(cards) {
		return HandRank{Rank: -1}
	}

	// Get all ranks and sort descending
	ranks := make([]int, len(cards))
	for i, card := range cards {
		ranks[i] = rankToNumeric(card.Rank)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(ranks)))

	return HandRank{
		Rank:    5,
		Kickers: ranks,
	}
}

// checkStraight checks for straight
func checkStraight(cards []Card) HandRank {
	isStraightHand, highCard := isStraight(cards)
	if !isStraightHand {
		return HandRank{Rank: -1}
	}

	// For wheel, special kickers
	if highCard == 5 {
		return HandRank{
			Rank:    4,
			Kickers: []int{5, 4, 3, 2, 1},
		}
	}

	// For regular straight, build kickers from high to low
	kickers := []int{highCard, highCard - 1, highCard - 2, highCard - 3, highCard - 4}

	return HandRank{
		Rank:    4,
		Kickers: kickers,
	}
}

// checkThreeOfAKind checks for three of a kind
func checkThreeOfAKind(cards []Card) HandRank {
	grouped := groupByRank(cards)

	var tripleRank int
	var kickers []int

	for rank, rankCards := range grouped {
		if len(rankCards) == 3 {
			tripleRank = rank
		} else if len(rankCards) == 1 {
			kickers = append(kickers, rank)
		}
	}

	if tripleRank == 0 {
		return HandRank{Rank: -1}
	}

	// Sort kickers descending
	sort.Sort(sort.Reverse(sort.IntSlice(kickers)))

	// Prepend triple rank
	finalKickers := []int{tripleRank}
	finalKickers = append(finalKickers, kickers...)

	return HandRank{
		Rank:    3,
		Kickers: finalKickers,
	}
}

// checkTwoPair checks for two pair
func checkTwoPair(cards []Card) HandRank {
	grouped := groupByRank(cards)

	var pairRanks []int
	var kicker int

	for rank, rankCards := range grouped {
		if len(rankCards) == 2 {
			pairRanks = append(pairRanks, rank)
		} else if len(rankCards) == 1 {
			kicker = rank
		}
	}

	if len(pairRanks) != 2 {
		return HandRank{Rank: -1}
	}

	// Sort pairs descending
	sort.Sort(sort.Reverse(sort.IntSlice(pairRanks)))

	finalKickers := []int{pairRanks[0], pairRanks[1], kicker}

	return HandRank{
		Rank:    2,
		Kickers: finalKickers,
	}
}

// checkOnePair checks for one pair
func checkOnePair(cards []Card) HandRank {
	grouped := groupByRank(cards)

	var pairRank int
	var kickers []int

	for rank, rankCards := range grouped {
		if len(rankCards) == 2 {
			pairRank = rank
		} else if len(rankCards) == 1 {
			kickers = append(kickers, rank)
		}
	}

	if pairRank == 0 {
		return HandRank{Rank: -1}
	}

	// Sort kickers descending
	sort.Sort(sort.Reverse(sort.IntSlice(kickers)))

	// Prepend pair rank
	finalKickers := []int{pairRank}
	finalKickers = append(finalKickers, kickers...)

	return HandRank{
		Rank:    1,
		Kickers: finalKickers,
	}
}

// checkHighCard checks for high card
func checkHighCard(cards []Card) HandRank {
	ranks := make([]int, len(cards))
	for i, card := range cards {
		ranks[i] = rankToNumeric(card.Rank)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(ranks)))

	return HandRank{
		Rank:    0,
		Kickers: ranks,
	}
}

// CompareHands compares two poker hands
// Returns 1 if hand1 is better, -1 if hand2 is better, 0 if tied
func CompareHands(hand1, hand2 HandRank) int {
	if hand1.Rank > hand2.Rank {
		return 1
	}
	if hand1.Rank < hand2.Rank {
		return -1
	}

	// Same rank, compare kickers
	for i := 0; i < len(hand1.Kickers) && i < len(hand2.Kickers); i++ {
		if hand1.Kickers[i] > hand2.Kickers[i] {
			return 1
		}
		if hand1.Kickers[i] < hand2.Kickers[i] {
			return -1
		}
	}

	return 0
}
