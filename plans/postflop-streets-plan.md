## Plan: Postflop Streets - Flop/Turn/River

This feature enables the progression of poker hands through the flop (3 cards), turn (1 card), and river (1 card) community card streets. After the preflop betting round completes, the game will automatically deal board cards and continue betting rounds until showdown or all players fold.

**Phases: 5**

### **Phase 1: Backend Board Card Storage & Dealing**
- **Objective:** Add board card state to the Hand struct and implement methods to deal flop/turn/river with proper deck management (burn cards discarded, not stored).
- **Files/Functions to Modify/Create:**
  - `internal/server/table.go` - Add `BoardCards []Card` field to Hand struct
  - `internal/server/table.go` - Implement `Hand.DealFlop()`, `Hand.DealTurn()`, `Hand.DealRiver()` methods
  - `internal/server/table_test.go` - Add tests for board card dealing
  - `internal/server/card_distribution_test.go` - Add tests for burn cards
- **Tests to Write:**
  - `TestHand_DealFlop_DealsThreeCards`
  - `TestHand_DealFlop_BurnsCardBeforeDealing`
  - `TestHand_DealTurn_DealsOneCard`
  - `TestHand_DealTurn_BurnsCardBeforeDealing`
  - `TestHand_DealRiver_DealsOneCard`
  - `TestHand_DealRiver_BurnsCardBeforeDealing`
  - `TestHand_BoardCards_InitiallyEmpty`
  - `TestHand_DealFlop_ErrorsIfDeckExhausted`
- **Steps:**
  1. Write failing tests for board card storage and dealing methods
  2. Add `BoardCards []Card` field to Hand struct in `table.go`
  3. Implement `DealFlop()` method (burn 1, deal 3 cards to board)
  4. Implement `DealTurn()` method (burn 1, deal 1 card to board)
  5. Implement `DealRiver()` method (burn 1, deal 1 card to board)
  6. Run tests to verify they pass
  7. Run linter and formatter

### **Phase 2: Street Progression Trigger Logic**
- **Objective:** Automatically detect when preflop/flop/turn betting rounds complete and advance to the next street by dealing board cards.
- **Files/Functions to Modify/Create:**
  - `internal/server/table.go` - Modify `Hand.ProcessAction()` to detect round completion and advance streets
  - `internal/server/table.go` - Add helper method `Hand.AdvanceToNextStreet()` to orchestrate street transitions
  - `internal/server/table_test.go` - Add integration tests for street progression
- **Tests to Write:**
  - `TestHand_PreflopComplete_AdvancesToFlop`
  - `TestHand_FlopComplete_AdvancesToTurn`
  - `TestHand_TurnComplete_AdvancesToRiver`
  - `TestHand_RiverComplete_DoesNotAdvance`
  - `TestHand_AdvanceToNextStreet_DealsBoardCards`
  - `TestHand_AdvanceToNextStreet_ResetsBettingState`
  - `TestHand_AllFoldedPreflop_DoesNotDealBoard`
- **Steps:**
  1. Write failing tests for automatic street progression
  2. Implement `AdvanceToNextStreet()` helper method that checks current street and calls appropriate deal method
  3. Modify `ProcessAction()` to call `AdvanceToNextStreet()` when `IsBettingRoundComplete()` returns true
  4. Ensure betting state is reset via existing `AdvanceStreet()` method
  5. Add edge case handling for all players folding before river
  6. Run tests to verify they pass
  7. Run linter and formatter

### **Phase 3: WebSocket Board Card Broadcasting**
- **Objective:** Broadcast board card events to all players at the table when flop/turn/river are dealt.
- **Files/Functions to Modify/Create:**
  - `internal/server/handlers.go` - Add `broadcastBoardDealt()` function
  - `internal/server/table.go` - Call `broadcastBoardDealt()` after each street's cards are dealt
  - `internal/server/handlers_test.go` - Add tests for board broadcast payloads
  - `internal/server/websocket_integration_test.go` - Add integration test for board event flow
- **Tests to Write:**
  - `TestBroadcastBoardDealt_SendsToAllTablePlayers`
  - `TestBroadcastBoardDealt_IncludesCorrectBoardCards`
  - `TestBroadcastBoardDealt_IncludesStreetIndicator`
  - `TestWebSocketFlow_FlopBroadcast_AfterPreflopComplete`
  - `TestWebSocketFlow_TurnBroadcast_AfterFlopComplete`
  - `TestWebSocketFlow_RiverBroadcast_AfterTurnComplete`
- **Steps:**
  1. Write failing tests for board broadcast events
  2. Define WebSocket message structure for `board_dealt` event (type, tableID, boardCards, street)
  3. Implement `broadcastBoardDealt()` function in `handlers.go`
  4. Hook `broadcastBoardDealt()` calls into `DealFlop()`, `DealTurn()`, `DealRiver()` methods
  5. Add integration test covering full preflop → flop transition with WebSocket messages
  6. Run tests to verify they pass
  7. Run linter and formatter

### **Phase 4: Frontend Board Display Component**
- **Objective:** Display community cards in the table UI and update the display when new board cards are dealt.
- **Files/Functions to Modify/Create:**
  - `frontend/src/components/TableView.tsx` - Add board card display section
  - `frontend/src/hooks/useWebSocket.ts` - Handle `board_dealt` event and update board state
  - `frontend/src/styles/TableView.css` - Add styles for board card display
  - `frontend/src/components/TableView.test.tsx` - Add tests for board rendering
  - `frontend/src/hooks/useWebSocket.test.ts` - Add tests for board event handling
- **Tests to Write:**
  - `TestTableView_DisplaysBoardCards_WhenPresent`
  - `TestTableView_EmptyBoard_PreflopState`
  - `TestTableView_DisplaysThreeCards_AfterFlop`
  - `TestTableView_DisplaysFourCards_AfterTurn`
  - `TestTableView_DisplaysFiveCards_AfterRiver`
  - `TestUseWebSocket_HandlesBoardDealtEvent`
  - `TestUseWebSocket_UpdatesBoardState_Incrementally`
- **Steps:**
  1. Write failing tests for board card display
  2. Add `boardCards` state field to TableView component
  3. Implement board card rendering section in TableView JSX (5 card slots, show cards as dealt)
  4. Add `board_dealt` event handler in `useWebSocket.ts`
  5. Update board state when event is received
  6. Add CSS styling for board card layout (centered, horizontal row)
  7. Run tests to verify they pass
  8. Run linter and formatter

### **Phase 5: Street Indicator & Flow Integration**
- **Objective:** Display the current street name (Preflop/Flop/Turn/River) in a separate game info section and ensure smooth action flow across all streets.
- **Files/Functions to Modify/Create:**
  - `frontend/src/components/TableView.tsx` - Add game info section with street indicator
  - `frontend/src/styles/TableView.css` - Style game info section and street indicator
  - `frontend/src/components/TableView.test.tsx` - Test street indicator updates
  - `internal/server/table_test.go` - Add end-to-end test for full hand progression
- **Tests to Write:**
  - `TestTableView_DisplaysStreetName_Preflop`
  - `TestTableView_DisplaysStreetName_Flop`
  - `TestTableView_DisplaysStreetName_Turn`
  - `TestTableView_DisplaysStreetName_River`
  - `TestTableView_StreetIndicator_UpdatesWhenBoardDealt`
  - `TestHand_FullHandProgression_PreflopToRiver`
  - `TestHand_ActionFlow_ContinuesAcrossStreets`
- **Steps:**
  1. Write failing tests for street indicator display
  2. Add game info section to TableView UI with street name display
  3. Include street field in `board_dealt` event payload
  4. Update street indicator when board events are received
  5. Add end-to-end backend test that progresses a hand from preflop through river with multiple betting rounds
  6. Verify action bar remains functional across street transitions
  7. Run all tests to verify integration
  8. Run linter and formatter

---

**Implementation Decisions:**

1. **Burn Cards**: Discarded silently, not stored or logged
2. **Street Indicator**: Displayed in separate game info section (not above/below board)
3. **Board Reveal**: Instant reveal, no animations
4. **All-In Run-Out**: One street at a time (simpler, reuses existing progression logic)
