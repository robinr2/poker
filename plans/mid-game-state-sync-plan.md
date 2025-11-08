## Plan: Mid-Game State Synchronization

When a client joins or rejoins a table during an active hand, they should immediately see the complete game state including dealer position, blind positions, pot amount, and cards (their own hole cards + card backs for opponents). Additionally, session restoration should automatically navigate users back to their table, and the UI should clearly indicate when a hand is in progress.

**Phases: 7**

1. **Phase 1: Add Stack Field to TableStateSeat Struct**
    - **Objective:** Fix the struct mismatch where code tries to read `seat.Stack` but the field doesn't exist in the struct definition
    - **Files/Functions to Modify/Create:**
        - `internal/server/handlers.go` - `TableStateSeat` struct (lines 536-541)
        - `internal/server/handlers.go` - `SendTableState` function (lines 543-613)
    - **Tests to Write:**
        - `TestTableStateSeatIncludesStack` - Verify stack field is included in table_state payload
        - `TestTableStateSerializationWithStacks` - Verify JSON serialization includes stack values
    - **Steps:**
        1. Write tests that verify table_state includes stack information for seated players
        2. Run tests to confirm they fail (struct doesn't have Stack field)
        3. Add `Stack *int` field to `TableStateSeat` struct in handlers.go
        4. Update `SendTableState` to populate stack values from table seats
        5. Run tests to confirm they pass
        6. Run all existing tests to ensure no regressions

2. **Phase 2: Extend TableStatePayload with Game State Fields**
    - **Objective:** Add optional game state fields to table_state message so it can convey active hand information
    - **Files/Functions to Modify/Create:**
        - `internal/server/handlers.go` - `TableStatePayload` struct (around line 547)
        - `internal/server/handlers.go` - `SendTableState` function (lines 543-613)
    - **Tests to Write:**
        - `TestTableStateIncludesGameStateWhenHandActive` - Verify game state fields populated when hand is active
        - `TestTableStateOmitsGameStateWhenNoHand` - Verify fields are nil/zero when no hand active
        - `TestTableStateGameStateFields` - Verify dealer, blinds, pot are correctly set
    - **Steps:**
        1. Write tests that verify table_state includes game state when a hand is active
        2. Run tests to confirm they fail (struct doesn't have game state fields)
        3. Add fields to `TableStatePayload`: `DealerSeat *int`, `SmallBlindSeat *int`, `BigBlindSeat *int`, `Pot *int`, `HandInProgress bool`
        4. Update `SendTableState` to check if `table.CurrentHand != nil` and populate game state fields accordingly
        5. Run tests to confirm they pass
        6. Run all existing tests to ensure no regressions

3. **Phase 3: Include Personalized Hole Cards and Card Count in table_state**
    - **Objective:** When sending table_state to any player during an active hand, include their hole cards (if seated) and card counts for all players so spectators can see card backs
    - **Files/Functions to Modify/Create:**
        - `internal/server/handlers.go` - `TableStatePayload` struct (add `HoleCards []Card` field)
        - `internal/server/handlers.go` - `TableStateSeat` struct (add `CardCount *int` field)
        - `internal/server/handlers.go` - `SendTableState` function (extract client's seat and include their cards + card counts)
    - **Tests to Write:**
        - `TestTableStateIncludesHoleCardsForSeatedPlayer` - Verify seated player receives their hole cards in table_state
        - `TestTableStateOmitsHoleCardsForUnseatedPlayer` - Verify spectators don't receive hole cards but do see card counts
        - `TestTableStateHoleCardsPrivacy` - Verify player only receives their own cards, not opponents'
        - `TestTableStateCardCountsForSpectators` - Verify unseated players see how many cards each player has
    - **Steps:**
        1. Write tests that verify table_state includes hole cards for seated players and card counts for all during active hands
        2. Run tests to confirm they fail (fields don't exist)
        3. Add `HoleCards []Card` field to `TableStatePayload` struct and `CardCount *int` to `TableStateSeat`
        4. Modify `SendTableState` to accept client parameter to identify which seat they occupy
        5. If hand is active and client is seated, populate HoleCards from `CurrentHand.HoleCards[seatIndex]`
        6. For all seats with cards, populate CardCount so spectators can see card backs
        7. Run tests to confirm they pass
        8. Run all existing tests to ensure no regressions

4. **Phase 4: Update Frontend to Handle Game State in table_state Message**
    - **Objective:** Modify frontend to extract and apply game state fields from table_state messages, including card counts for rendering opponent card backs
    - **Files/Functions to Modify/Create:**
        - `frontend/src/hooks/useWebSocket.ts` - table_state message handler (lines 145-154)
        - `frontend/src/hooks/useWebSocket.ts` - Update interfaces to match new payload structure
        - `frontend/src/components/TableView.tsx` - Update card back rendering logic to use cardCount
    - **Tests to Write:**
        - `TestTableStateUpdatesGameState` - Verify gameState is updated when table_state includes game fields
        - `TestTableStateSetsHoleCards` - Verify holeCards are set from table_state message
        - `TestTableStateUpdatesCardCounts` - Verify card counts are tracked per seat
        - `TestTableStatePreservesExistingGameState` - Verify game state isn't cleared if table_state lacks game fields
    - **Steps:**
        1. Write tests that verify frontend processes game state and card counts from table_state messages
        2. Run tests to confirm they fail (handler doesn't process these fields)
        3. Update TypeScript interfaces to match extended TableStatePayload structure
        4. Modify table_state message handler in useWebSocket to extract game state fields if present
        5. Update gameState when table_state includes dealer, blinds, pot, and hole cards
        6. Track card counts per seat in tableState
        7. Update TableView to render card backs based on cardCount instead of only when local player has holeCards
        8. Run tests to confirm they pass
        9. Run all frontend tests to ensure no regressions

5. **Phase 5: Session Restoration Auto-Navigation**
    - **Objective:** When a client reconnects and receives session_restored with a tableID, automatically navigate to the table view and request updated table_state
    - **Files/Functions to Modify/Create:**
        - `frontend/src/App.tsx` - Add logic to handle tableID in session_restored
        - `frontend/src/hooks/useWebSocket.ts` - Expose tableID from session_restored payload
        - `internal/server/websocket.go` - Ensure table_state is sent after session_restored when player is at a table
    - **Tests to Write:**
        - `TestSessionRestoredAutoNavigatesToTable` - Verify frontend switches to table view when session_restored includes tableID
        - `TestSessionRestoredSendsTableState` - Verify backend sends table_state after session_restored
        - `TestSessionRestoredWithoutTableStaysInLobby` - Verify players not at a table stay in lobby
    - **Steps:**
        1. Write tests that verify session restoration triggers table view navigation and receives table_state
        2. Run tests to confirm they fail (auto-navigation not implemented)
        3. Update backend session restoration handler to send table_state when player is at a table
        4. Update frontend session_restored message handler to expose tableID and seatIndex
        5. Add logic in App.tsx to automatically set currentView to table when session_restored includes tableID
        6. Run tests to confirm they pass
        7. Run all tests to ensure no regressions

6. **Phase 6: Add Hand-In-Progress UI Indicator**
    - **Objective:** Add visual indicator in TableView when a hand is actively in progress
    - **Files/Functions to Modify/Create:**
        - `frontend/src/components/TableView.tsx` - Add indicator component/badge
        - `frontend/src/styles/TableView.css` - Style the indicator
    - **Tests to Write:**
        - `TestHandInProgressIndicatorShown` - Verify indicator displays when hand is active (pot > 0 or dealer assigned)
        - `TestHandInProgressIndicatorHidden` - Verify indicator hidden when no hand active
        - `TestHandInProgressIndicatorStyling` - Verify indicator is visible and well-positioned
    - **Steps:**
        1. Write tests that verify hand-in-progress indicator displays correctly
        2. Run tests to confirm they fail (indicator not implemented)
        3. Add conditional rendering in TableView for hand-in-progress badge
        4. Style the badge/indicator to be prominent but not intrusive
        5. Display text like "Hand in Progress" or "🎴 Hand Active"
        6. Run tests to confirm they pass
        7. Run all frontend tests to ensure no regressions

7. **Phase 7: Integration Testing with Late Join and Restoration Scenarios**
    - **Objective:** Verify complete end-to-end flow for late join, session restoration, and spectator scenarios
    - **Files/Functions to Modify/Create:**
        - `internal/server/websocket_integration_test.go` - Add comprehensive test scenarios
        - Manual testing with multiple browser windows
    - **Tests to Write:**
        - `TestLateJoinReceivesCompleteGameState` - Start hand with 2 players, join 3rd player, verify they see dealer/blinds/pot
        - `TestLateJoinSeatedPlayerSeesOwnCards` - Player with existing seat gets their hole cards when joining mid-hand
        - `TestLateJoinSpectatorSeesCardBacks` - Unseated late joiner sees card backs for all players
        - `TestSessionRestorationMidHandShowsState` - Refresh browser mid-hand, verify auto-navigation and state sync
        - `TestMultipleSpectatorsSeeConsistentState` - Multiple spectators join, all see same card backs
    - **Steps:**
        1. Write integration tests covering late join, spectator, and restoration scenarios
        2. Run tests to confirm they pass with all previous phases implemented
        3. Perform manual testing: start hand with 2 players in separate browsers, open 3rd browser and join as spectator
        4. Verify spectator sees: dealer badge, blind badges, pot, card backs for both players, hand-in-progress indicator
        5. Refresh one player's browser mid-hand, verify they auto-navigate to table and see their cards
        6. Run all backend and frontend tests for final verification
        7. Document any edge cases discovered during testing
