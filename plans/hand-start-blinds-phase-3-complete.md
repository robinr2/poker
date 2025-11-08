## Phase 3 Complete: Deck Shuffle & Hole Card Dealing

Successfully implemented cryptographically secure deck shuffling using Fisher-Yates algorithm and hole card dealing logic that distributes 2 cards to each active player while skipping waiting and empty seats.

**Files created/changed:**
- internal/server/table.go
- internal/server/table_test.go

**Functions created/changed:**
- ShuffleDeck() (new function)
- Hand.DealHoleCards() (new method)

**Tests created/changed:**
- TestShuffleDeck (new test)
- TestShuffleDeckRandomization (new test)
- TestDealHoleCardsToActivePlayers (new test)
- TestDealHoleCardsSkipsWaiting (new test)
- TestDealHoleCardsReducesDeck (new test)
- TestDealHoleCardsEmptySeats (new test)
- TestDealHoleCardsAllPlayersActive (new test)
- TestDealHoleCardsInsufficientCards (new test)

**Review Status:** APPROVED (after revision)

**Initial Review Finding:**
- Issue: Missing test for insufficient cards error path
- Fix Applied: Added TestDealHoleCardsInsufficientCards to verify error handling
- Result: Complete test coverage for all code paths

**Technical Highlights:**
- Uses crypto/rand for cryptographically secure shuffling (prevents predictable decks)
- Correct Fisher-Yates (Knuth) shuffle implementation (no bias)
- Proper error handling for insufficient cards and crypto failures
- Only deals to "active" status seats (skips "waiting" and "empty")
- HoleCards stored in map[int][]Card (key=seat number, value=2 cards)
- Deck size properly reduced after dealing

**Git Commit Message:**
```
feat: implement deck shuffling and hole card dealing

- Add ShuffleDeck function using Fisher-Yates algorithm with crypto/rand
- Add DealHoleCards method to deal 2 cards to active players
- Skip waiting and empty seats during dealing
- Validate sufficient cards available before dealing
- Update Hand.HoleCards map with dealt cards
- Reduce deck size after dealing cards
- Add comprehensive tests for shuffling, dealing, and error cases
```
