## Plan: Critical Gameplay Bugs Fix

Fixes four critical bugs preventing proper poker gameplay: missing showdown trigger on river, incorrect minimum raise calculation postflop, missing folded player visual indicators, and incorrect action order when big blind raises preflop.

**Phases: 5 phases**

### 1. **Phase 1: Fix Missing Showdown Trigger (Backend)**
   - **Objective:** Add showdown trigger when betting completes on river with multiple active players
   - **Files/Functions to Modify/Create:** 
     - `internal/server/handlers.go` - `HandleAction()` method (around lines 1356-1401)
   - **Tests to Write:** 
     - `TestHandleAction_RiverBettingCompleteTriggersShowdown` - Verify showdown is called when betting completes on river
     - `TestHandleAction_RiverNoShowdownIfNotComplete` - Verify showdown is NOT called if betting incomplete
   - **Steps:**
     1. Write failing tests for river showdown trigger logic
     2. Add else branch in `HandleAction()` to trigger showdown when `currentStreet == "river"` and betting round complete
     3. Unlock mutex before calling `table.HandleShowdown()` and re-lock after
     4. Run tests to verify they pass
     5. Lint and format code

### 2. **Phase 2: Fix Minimum Raise Calculation (Backend)**
   - **Objective:** Preserve big blind as minimum raise increment when advancing to postflop streets
   - **Files/Functions to Modify/Create:**
     - `internal/server/table.go` - `AdvanceStreet()` method (around lines 1466-1473)
     - `internal/server/table_test.go` - `TestAdvanceStreet_ResetsLastRaise` (needs update)
   - **Tests to Write:**
     - `TestAdvanceStreet_PreservesMinimumRaisePostflop` - Verify LastRaise equals 20 (big blind) after advancing to flop/turn/river
     - `TestGetMinRaise_PostflopFirstAction` - Verify minimum raise is 40 when no raises yet postflop
   - **Steps:**
     1. Write failing tests for postflop minimum raise preservation
     2. Modify `AdvanceStreet()` to set `h.LastRaise = 20` when advancing to postflop streets (hardcode value for simplicity)
     3. Update existing test `TestAdvanceStreet_ResetsLastRaise` to expect LastRaise=20 postflop
     4. Run tests to verify they pass
     5. Lint and format code

### 3. **Phase 3: Remove Auto-Clear and Show Start Hand Button After Showdown (Frontend)**
   - **Objective:** Keep showdown overlay visible until player clicks "Start Hand" button, remove 5-second auto-clear timeout
   - **Files/Functions to Modify/Create:**
     - `frontend/src/hooks/useWebSocket.ts` - Remove timeout logic (lines 557-570)
     - `frontend/src/components/TableView.tsx` - Show "Start Hand" button when showdown is displayed
     - `frontend/src/styles/TableView.css` - Style start hand button overlay
   - **Tests to Write:**
     - `TestUseWebSocket_ShowdownPersists` - Verify showdown state is not auto-cleared after timeout
     - `TestTableView_StartHandButtonShownAfterShowdown` - Verify start hand button appears when showdown state exists
   - **Steps:**
     1. Write failing tests for persistent showdown display
     2. Remove `handCompleteTimeoutRef` and all timeout logic from useWebSocket
     3. Modify TableView to show "Start Hand" button when `gameState.showdown` exists
     4. Update sendMessage to clear local showdown state when "start_hand" is sent
     5. Run tests to verify they pass
     6. Lint and format code

### 4. **Phase 4: Add Folded Player Visual Indicator (Frontend)**
   - **Objective:** Apply visual styling to clearly show which players are folded vs active
   - **Files/Functions to Modify/Create:**
     - `frontend/src/components/TableView.tsx` - Seat rendering (around lines 166-233)
     - `frontend/src/styles/TableView.css` - Add `.folded` class styles
   - **Tests to Write:**
     - `TestTableView_FoldedPlayerStyling` - Verify folded class is applied to folded players
     - `TestTableView_ActivePlayerNoFoldedClass` - Verify active players don't have folded class
   - **Steps:**
     1. Write failing tests for folded player styling
     2. Add conditional `folded` class to seat className based on `gameState?.foldedPlayers?.includes(seat.index)`
     3. Add CSS rule `.seat.folded` with opacity 0.5, gray background, and muted border
     4. Run tests to verify they pass
     5. Lint and format code

### 5. **Phase 5: Debug Big Blind Extra Action Bug (Backend)**
   - **Objective:** Investigate and fix incorrect action order when BB raises preflop and everyone calls
   - **Files/Functions to Modify/Create:**
     - `internal/server/handlers.go` - `HandleAction()` method
     - `internal/server/table.go` - `IsBettingRoundComplete()`, `GetFirstActor()`, or `AdvanceAction()` methods
   - **Tests to Write:**
     - `TestBigBlindRaisePreflopClosesAction` - Verify BB doesn't get action postflop after raising preflop when all call
     - `TestActionOrder_PostflopAfterBBRaise` - Verify first player after BB acts first postflop
   - **Steps:**
     1. Write comprehensive test reproducing exact scenario: BB raises preflop → all call → verify SB acts first on flop
     2. Run test to confirm it fails and identify root cause
     3. Fix the identified issue (likely in action completion or street advancement logic)
     4. Run test to verify it passes
     5. Run full test suite to ensure no regressions
     6. Lint and format code

### 6. **Phase 6: Integration Testing & Manual Verification**
   - **Objective:** Verify all fixes work together in a complete hand flow
   - **Files/Functions to Modify/Create:**
     - New integration test or add to existing `internal/server/websocket_integration_test.go`
   - **Tests to Write:**
     - `TestCompleteHandFlow_WithShowdown` - Full hand from deal through river to showdown with multiple players
     - `TestCompleteHandFlow_MinRaisePostflop` - Verify min raise calculation correct postflop
   - **Steps:**
     1. Write integration test covering full hand flow with all fixed scenarios
     2. Run test to verify it passes
     3. Build frontend and backend
     4. Manual test: Start server, play complete hand, verify showdown displays and persists
     5. Manual test: Verify "Start Hand" button appears and starts next hand
     6. Manual test: Verify min raise shows 40 (2x BB) after checks on flop
     7. Manual test: Verify folded players are visually grayed out
     8. Manual test: BB raises preflop, all call, verify action order correct on flop
