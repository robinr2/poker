## Plan: Showdown & Settlement

This plan implements poker hand evaluation, winner determination, pot distribution, and transitioning to the next hand. When the river betting completes (or all but one player folds), the system will determine the winner, award the pot, update stacks, handle bust-outs, rotate the dealer, and allow starting the next hand via the existing "Start Hand" button.

**Phases: 6**

### 1. **Phase 1: Hand Evaluation Engine (Backend)**
- **Objective:** Build pure Go poker hand evaluator to rank 5-card hands from best 7 cards
- **Files/Functions to Create:**
  - `internal/server/hand_evaluator.go` - New file
    - `type HandRank struct` - Contains rank type (int) and kickers/tiebreakers
    - `EvaluateHand(holeCards []Card, boardCards []Card) HandRank` - Returns best 5-card hand rank
    - `CompareHands(rank1, rank2 HandRank) int` - Returns -1/0/1 for comparison
    - Helper functions: `isFlush()`, `isStraight()`, `groupByRank()`, `findBestFiveCards()`
- **Tests to Write:**
  - `internal/server/hand_evaluator_test.go`
    - `TestEvaluateHand_RoyalFlush`
    - `TestEvaluateHand_StraightFlush`
    - `TestEvaluateHand_FourOfAKind`
    - `TestEvaluateHand_FullHouse`
    - `TestEvaluateHand_Flush`
    - `TestEvaluateHand_Straight`
    - `TestEvaluateHand_ThreeOfAKind`
    - `TestEvaluateHand_TwoPair`
    - `TestEvaluateHand_OnePair`
    - `TestEvaluateHand_HighCard`
    - `TestCompareHands_DifferentRanks`
    - `TestCompareHands_SameRank_Kickers`
    - `TestCompareHands_Tie`
    - `TestEvaluateHand_WheelStraight` (A-2-3-4-5)
    - `TestEvaluateHand_AllSevenCards` (pick best 5 from 7)
- **Steps:**
  1. Write all 15 failing tests covering hand types, comparisons, and edge cases
  2. Run tests to confirm failures
  3. Implement `HandRank` struct (rank int 0-9, kickers []int)
  4. Implement helper functions (isFlush, isStraight, groupByRank)
  5. Implement `EvaluateHand()` - check from royal flush down to high card
  6. Implement `CompareHands()` - compare rank first, then kickers
  7. Run tests to confirm all pass
  8. Run full test suite

### 2. **Phase 2: Showdown Trigger & Winner Detection (Backend)**
- **Objective:** Detect when showdown occurs and determine winner(s) using hand evaluator
- **Files/Functions to Modify/Create:**
  - `internal/server/table.go`
    - `Hand.DetermineWinner(seats []*Seat) (winners []int, winningRank *HandRank)` - Returns seat indices of winner(s)
    - `Table.HandleShowdown()` - Main showdown orchestration method
  - `internal/server/handlers.go` (line 1270)
    - Add showdown trigger when river betting completes: `if h.Street == "river" { table.HandleShowdown() }`
- **Tests to Write:**
  - `internal/server/table_test.go`
    - `TestDetermineWinner_SingleWinner_HighCard`
    - `TestDetermineWinner_SingleWinner_Flush`
    - `TestDetermineWinner_Tie_TwoPlayers`
    - `TestDetermineWinner_Tie_ThreePlayers`
    - `TestDetermineWinner_HeadsUp`
    - `TestDetermineWinner_MultiWay_FourPlayers`
    - `TestDetermineWinner_SkipsFoldedPlayers`
    - `TestHandleShowdown_TriggersOnRiverComplete`
    - `TestHandleShowdown_EarlyWinner_AllFold`
  - `internal/server/handlers_test.go`
    - `TestHandlerFlow_RiverToShowdown`
    - `TestHandlerFlow_AllFoldBeforeShowdown`
- **Steps:**
  1. Write all 11 failing tests
  2. Run tests to confirm failures
  3. Implement `Hand.DetermineWinner()` - iterate non-folded players, evaluate hands, find best
  4. Handle ties (multiple winners with same rank)
  5. Handle early winner (single non-folded player, no evaluation needed)
  6. Implement `Table.HandleShowdown()` - call DetermineWinner, prepare for pot distribution
  7. Add showdown trigger in handlers.go line 1270
  8. Run tests to confirm all pass
  9. Run full test suite

### 3. **Phase 3: Pot Distribution & Stack Updates (Backend)**
- **Objective:** Award pot to winner(s), update stacks, detect bust-outs
- **Files/Functions to Modify:**
  - `internal/server/table.go`
    - Extend `Table.HandleShowdown()` to distribute pot and update stacks
    - `Table.DistributePot(winners []int, pot int) map[int]int` - Returns amountWon per seat
    - `Table.HandleBustOuts()` - Clear seats with stack == 0
- **Tests to Write:**
  - `internal/server/table_test.go`
    - `TestDistributePot_SingleWinner`
    - `TestDistributePot_TwoWayTie_EvenSplit`
    - `TestDistributePot_ThreeWayTie_EvenSplit`
    - `TestDistributePot_TwoWayTie_OddPot` (test remainder handling - first winner by seat order gets extra chip)
    - `TestHandleShowdown_UpdatesStacks`
    - `TestHandleShowdown_DetectsBustOut`
    - `TestHandleShowdown_ClearsBustOutSeat`
    - `TestHandleBustOuts_MultiplePlayersOut`
    - `TestHandleBustOuts_WinnerDoesNotBustOut`
  - `internal/server/handlers_test.go`
    - `TestHandlerFlow_ShowdownStackUpdates`
    - `TestHandlerFlow_BustOutAfterLoss`
- **Steps:**
  1. Write all 11 failing tests
  2. Run tests to confirm failures
  3. Implement `DistributePot()` - divide pot by number of winners, give remainder to first winner by seat order
  4. Extend `HandleShowdown()` to call DistributePot and update seat.Stack values
  5. Implement `HandleBustOuts()` - iterate seats, clear any with stack == 0
  6. Call `HandleBustOuts()` from `HandleShowdown()` after stack updates
  7. Run tests to confirm all pass
  8. Run full test suite

### 4. **Phase 4: Hand Cleanup & Next Hand Preparation (Backend)**
- **Objective:** Complete hand lifecycle - rotate dealer, clear hand state, prepare for "Start Hand" button
- **Files/Functions to Modify:**
  - `internal/server/table.go`
    - Extend `Table.HandleShowdown()` to complete hand lifecycle
    - Call existing `Table.NextDealer()` to rotate button
    - Set `table.CurrentHand = nil`
    - Do NOT auto-start next hand - rely on existing "Start Hand" button flow
- **Tests to Write:**
  - `internal/server/table_test.go`
    - `TestHandleShowdown_RotatesDealer`
    - `TestHandleShowdown_ClearsHandState`
    - `TestHandleShowdown_NoAutoStartNextHand`
    - `TestHandleShowdown_PromotesWaitingPlayersOnNextStart` (verify existing StartHand does this)
  - `internal/server/handlers_test.go`
    - `TestHandlerFlow_FullHandCycle_ManualNextHand`
    - `TestHandlerFlow_HandEndsWithBustOut`
    - `TestHandlerFlow_DealerRotatesAfterShowdown`
    - `TestHandlerFlow_StartHandButtonWorksAfterShowdown`
- **Steps:**
  1. Write all 8 failing tests
  2. Run tests to confirm failures
  3. Extend `HandleShowdown()` to call `NextDealer()` after pot distribution
  4. Set `table.CurrentHand = nil` to clear hand state
  5. Do NOT call StartHand() - leave that to existing button handler
  6. Run tests to confirm all pass
  7. Run full test suite

### 5. **Phase 5: WebSocket Broadcasts (Backend)**
- **Objective:** Send showdown results and hand completion events to frontend clients
- **Files/Functions to Create/Modify:**
  - `internal/server/handlers.go`
    - `type ShowdownResultPayload struct` - winnerSeats []int, winningHand string (simple name like "Flush"), potAmount int, amountWon map[int]int
    - `Server.broadcastShowdown(table *Table, winners []int, rank *HandRank, amountsWon map[int]int)`
    - `type HandCompletePayload struct` - message string
    - `Server.broadcastHandComplete(table *Table)`
  - `internal/server/table.go`
    - Modify `Table.HandleShowdown()` to accept server parameter and call broadcast methods
- **Tests to Write:**
  - `internal/server/handlers_test.go`
    - `TestBroadcastShowdown_SingleWinner`
    - `TestBroadcastShowdown_TieWithSplit`
    - `TestBroadcastShowdown_WinningHandDescription`
    - `TestBroadcastHandComplete_Message`
  - `internal/server/websocket_integration_test.go`
    - `TestWebSocketFlow_ShowdownMessages`
    - `TestWebSocketFlow_FullHandToShowdownToNextHand`
    - `TestWebSocketFlow_BustOutClearsSeats`
    - `TestWebSocketFlow_StackUpdatesAfterShowdown`
- **Steps:**
  1. Write all 8 failing tests
  2. Run tests to confirm failures
  3. Define `ShowdownResultPayload` and `HandCompletePayload` structs
  4. Implement `broadcastShowdown()` - marshal payload, send "showdown_result" message
  5. Implement `broadcastHandComplete()` - marshal payload, send "hand_complete" message
  6. Generate simple hand name (e.g., "Flush", "Straight", "Two Pair")
  7. Modify `HandleShowdown()` to accept server parameter and call broadcast methods
  8. Run tests to confirm all pass
  9. Run full test suite

### 6. **Phase 6: Frontend Showdown Display**
- **Objective:** Handle showdown events, reveal all cards simultaneously, highlight winner, show winning hand, update stacks
- **Files/Functions to Modify:**
  - `frontend/src/hooks/useWebSocket.ts`
    - Add `showdown_result` event handler
    - Add `hand_complete` event handler
    - Update GameState with showdown info, revealed cards, winner, hand description
  - `frontend/src/components/TableView.tsx`
    - Display revealed hole cards for all non-folded players (simultaneous flip)
    - Highlight winner seat(s) with border/glow effect
    - Show winning hand description overlay
    - Update stack values
    - Show hand complete message
    - Existing "Start Hand" button already present - no changes needed
  - `frontend/src/styles/TableView.css`
    - Add `.winner-highlight` class for winner animation
    - Add `.hand-description` overlay styles
- **Tests to Write:**
  - `frontend/src/hooks/useWebSocket.test.ts`
    - `test('handles showdown_result event')`
    - `test('handles hand_complete event')`
    - `test('updates game state with winner info')`
  - `frontend/src/components/TableView.test.tsx`
    - `test('reveals hole cards on showdown')`
    - `test('highlights winner seat')`
    - `test('displays winning hand description')`
    - `test('shows split pot for ties')`
    - `test('updates stack values after showdown')`
    - `test('shows hand complete message')`
- **Steps:**
  1. Write all 9 failing tests
  2. Run tests to confirm failures (npm test)
  3. Add `showdown_result` handler in useWebSocket.ts - update state with winners, hand description
  4. Add `hand_complete` handler - display completion message
  5. Modify TableView.tsx to reveal all cards simultaneously when showdown state present
  6. Add winner highlight styling and conditionally apply to winner seats
  7. Display winning hand description as overlay or banner
  8. Update stack values from showdown result
  9. Show hand complete message
  10. Run tests to confirm all pass (npm test)
  11. Manual browser testing to verify full showdown flow

---

## Implementation Decisions

1. **Hand description format:** Simple names only - "Flush", "Straight", "Two Pair", etc.
2. **Odd chip in split pots:** First winner by seat order gets the extra chip (minimal code)
3. **Showdown card reveal:** All cards flip simultaneously (simpler, no sequencing)
4. **Next hand start:** Use existing "Start Hand" button - no auto-start, no timer
5. **Bust-out handling:** Clear seat immediately after showdown - player stays at table/lobby (simplest approach)

---

## Open Questions

None - all implementation decisions finalized based on user input.
