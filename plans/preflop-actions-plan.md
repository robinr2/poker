## Plan: Preflop Actions - Call/Check/Fold Only

Implement the core betting action system for preflop play, enabling players to fold, check (when valid), or call to match the current bet. This system establishes turn-based action order (with heads-up and multi-player rules), validates actions, manages pot and bet tracking, and detects betting round closure—all without raises (Feature 6 will add raising).

**Phases: 5**

### 1. **Phase 1: Backend Turn Order & Action State**
   - **Objective:** Extend Hand struct with action tracking fields and implement turn order logic (heads-up vs multi-player, preflop rules)
   - **Files/Functions to Modify/Create:**
     - `internal/server/table.go`: Add new Hand fields (CurrentActor, CurrentBet, PlayerBets, FoldedPlayers, ActedPlayers, Street)
     - `internal/server/table.go`: Implement `GetFirstActor() int` (handles HU dealer-first, multi UTG-first)
     - `internal/server/table.go`: Implement `GetNextActiveSeat(fromSeat int) *int` (skips folded, wraps around)
   - **Tests to Write:**
     - `TestGetFirstActor_HeadsUp` - Verify dealer acts first preflop in HU
     - `TestGetFirstActor_MultiPlayer` - Verify first seat after BB acts first in 3+ player game
     - `TestGetNextActiveSeat` - Verify wrap-around and folded player skipping
   - **Steps:**
     1. Add new Hand struct fields with TDD (test first, fail, implement, pass)
     2. Implement GetFirstActor with HU/multi-player logic (test-driven)
     3. Implement GetNextActiveSeat with wrap-around and fold skip logic (test-driven)
     4. Run all existing tests to ensure no regressions
     5. Run `scripts/lint.sh` and fix any issues

### 2. **Phase 2: Backend Action Validation & Processing**
   - **Objective:** Implement action validation logic (fold/check/call rules) and action processing that updates game state (stacks, pots, bets)
   - **Files/Functions to Modify/Create:**
     - `internal/server/table.go`: Implement `GetValidActions(seatIndex int) []string`
     - `internal/server/table.go`: Implement `ProcessAction(seatIndex int, action string) error`
     - `internal/server/table.go`: Implement `GetCallAmount(seatIndex int) int`
   - **Tests to Write:**
     - `TestGetValidActions_CanCheck` - When current bet matches player bet, can check/fold
     - `TestGetValidActions_MustCall` - When behind current bet, can call/fold only
     - `TestProcessAction_Fold` - Marks player as folded, doesn't change pot/stacks
     - `TestProcessAction_Check` - Valid only when bet matched, no state change
     - `TestProcessAction_Call` - Moves chips from stack to pot, updates PlayerBets
     - `TestProcessAction_CallPartial` - Handle all-in when stack < call amount (auto all-in)
   - **Steps:**
     1. Implement GetValidActions with check/call logic (test-driven)
     2. Implement GetCallAmount helper (test-driven)
     3. Implement ProcessAction for fold (test-driven)
     4. Implement ProcessAction for check with validation (test-driven)
     5. Implement ProcessAction for call with pot/stack updates (test-driven)
     6. Handle all-in case when stack insufficient for full call - auto all-in for remaining stack (test-driven)
     7. Run all tests and lint

### 3. **Phase 3: Backend Betting Round Closure & Action Flow**
   - **Objective:** Implement betting round closure detection and action advancement logic to orchestrate the full preflop action sequence
   - **Files/Functions to Modify/Create:**
     - `internal/server/table.go`: Implement `IsBettingRoundComplete() bool`
     - `internal/server/table.go`: Implement `AdvanceAction() (*int, error)`
     - `internal/server/table.go`: Modify `StartHand()` to initialize action state and request first action
   - **Tests to Write:**
     - `TestIsBettingRoundComplete_NotAllActed` - Returns false when players haven't acted
     - `TestIsBettingRoundComplete_BetsNotMatched` - Returns false when bets unmatched
     - `TestIsBettingRoundComplete_AllMatched` - Returns true when all acted and matched
     - `TestIsBettingRoundComplete_AllFoldedButOne` - Returns true when only one player left
     - `TestAdvanceAction` - Moves to next active player, handles wrap-around
     - `TestStartHand_InitializesActionState` - Verify action fields initialized correctly
   - **Steps:**
     1. Implement IsBettingRoundComplete with all closure conditions (test-driven)
     2. Implement AdvanceAction to move turn to next player (test-driven)
     3. Modify StartHand to initialize action state and call GetFirstActor (test-driven)
     4. Add helper to request action from first player after hand start (test-driven)
     5. Send "betting_round_complete" message with outcome when round closes (test-driven)
     6. Run all tests and lint

### 4. **Phase 4: WebSocket Protocol & Handler**
   - **Objective:** Define WebSocket message types for actions and implement backend handler to receive player actions and broadcast results
   - **Files/Functions to Modify/Create:**
     - `internal/server/handlers.go`: Add ActionRequestPayload, PlayerActionPayload, ActionResultPayload structs
     - `internal/server/handlers.go`: Implement `HandlePlayerAction(c *Client, ...) error`
     - `internal/server/websocket.go`: Add "player_action" case in readPump message router
     - `internal/server/server.go`: Add `BroadcastActionRequest(tableID string, seatIndex int, ...)` and `BroadcastActionResult(...)`
   - **Tests to Write:**
     - `TestHandlePlayerAction_ValidCall` - Verify call action processes and broadcasts
     - `TestHandlePlayerAction_ValidCheck` - Verify check action processes and advances turn
     - `TestHandlePlayerAction_ValidFold` - Verify fold action marks player folded
     - `TestHandlePlayerAction_InvalidAction` - Verify error on invalid action (e.g., check when must call)
     - `TestHandlePlayerAction_OutOfTurn` - Verify error when not current actor
     - `TestActionSequence_Integration` - Full preflop action sequence with 3 players
   - **Steps:**
     1. Define payload structs in handlers.go (test-driven)
     2. Implement HandlePlayerAction with validation (test-driven)
     3. Add broadcast functions for action_request and action_result (test-driven)
     4. Wire up "player_action" case in websocket.go message router (test-driven)
     5. Write integration test for full action sequence (test-driven)
     6. Run all tests and lint

### 5. **Phase 5: Frontend Action Bar & Turn Indicator**
   - **Objective:** Build UI components for action buttons (Fold, Check, Call) and turn indicator, wire up WebSocket message handlers
   - **Files/Functions to Modify/Create:**
     - `frontend/src/hooks/useWebSocket.ts`: Add action_request and action_result message handlers
     - `frontend/src/components/TableView.tsx`: Add ActionBar component with Fold/Check/Call buttons
     - `frontend/src/components/TableView.tsx`: Add turn indicator highlighting for currentActor seat
     - `frontend/src/styles/TableView.css`: Add styles for action buttons and turn highlight
   - **Tests to Write:**
     - `TestUseWebSocket_ActionRequest` - Verify state updates on action_request message
     - `TestUseWebSocket_ActionResult` - Verify pot/stack updates on action_result message
     - `TestTableView_ActionButtonsVisible` - Verify buttons shown only for current actor
     - `TestTableView_CallButtonAmount` - Verify "Call X" label shows correct amount
     - `TestTableView_CheckVsCall` - Verify "Check" shown when call amount is 0
     - `TestTableView_TurnIndicator` - Verify current actor seat highlighted
   - **Steps:**
     1. Extend game state in useWebSocket with action fields (test-driven)
     2. Add action_request message handler to update currentActor/validActions (test-driven)
     3. Add action_result message handler to update pot/stacks/folded status (test-driven)
     4. Implement ActionBar component with conditional button rendering - show only valid button (test-driven)
     5. Add sendAction helper to send player_action message (test-driven)
     6. Add turn indicator CSS class to highlight current actor's seat (test-driven)
     7. Wire up button clicks to sendAction (test-driven)
     8. Run all frontend tests and lint

**Implementation Notes:**
1. **All-in short-call:** When stack < call amount, automatically all-in for remaining stack in call action
2. **Betting round completion:** Send "betting_round_complete" message with outcome when round closes, then pause (Feature 7 will handle street progression)
3. **Check vs Call button:** Show only the valid button dynamically ("Check" if can check, "Call X" if must call)
