## Plan: Raises (Basic NL)

This plan implements min-raise computation, validation, and UI for no-limit poker raises while preventing side pot scenarios. Builds on completed Feature #5 (Call/Check/Fold) by adding raise actions with amount tracking, min/max validation, and a number input UI with presets. All logic supports both heads-up and multi-player games.

**Phases: 6**

### **Phase 1: Min-Raise Computation and Validation**
- **Objective:** Add backend logic to compute minimum valid raise amounts based on poker rules (min-raise = current bet + last raise increment)
- **Files/Functions to Modify/Create:**
  - `internal/server/table.go`: Add `LastRaise int` field to Hand struct
  - `internal/server/table.go`: Create `GetMinRaise(seatIndex int) int` method
  - `internal/server/table.go`: Update `NewHand()` to initialize `LastRaise = bigBlind`
  - `internal/server/table.go`: Update `AdvanceStreet()` to reset `LastRaise = 0` for post-flop streets
  - `internal/server/table_test.go`: New test file section for raise tests
- **Tests to Write:**
  - `TestGetMinRaise_Preflop` - BB=20, min-raise should be 40
  - `TestGetMinRaise_AfterRaise` - After raise to 60, min-raise should be 100
  - `TestGetMinRaise_AfterMultipleRaises` - Chain of raises maintains correct increments
  - `TestGetMinRaise_PostFlop` - FirstBet=30, min-raise=60
  - `TestGetMinRaise_HeadsUp` - Works correctly in heads-up scenario
  - `TestNewHand_InitializesLastRaise` - LastRaise starts at BB amount
  - `TestAdvanceStreet_ResetsLastRaise` - LastRaise resets to 0 after street change
- **Steps:**
  1. Write failing test `TestGetMinRaise_Preflop` expecting min-raise of 40 when BB=20
  2. Add `LastRaise int` field to Hand struct
  3. Update `NewHand()` to set `LastRaise = bigBlind`
  4. Implement `GetMinRaise()` returning `CurrentBet + LastRaise`
  5. Run test to confirm it passes
  6. Write failing test `TestAdvanceStreet_ResetsLastRaise`
  7. Update `AdvanceStreet()` to set `LastRaise = 0`
  8. Run test to confirm it passes
  9. Write remaining tests for multiple raises, post-flop, and heads-up scenarios
  10. Run all tests to confirm they pass

### **Phase 2: Max-Raise and Side Pot Prevention**
- **Objective:** Add logic to compute maximum valid raise (player's stack or smallest opponent stack to prevent side pots) and raise error when side pot would be created
- **Files/Functions to Modify/Create:**
  - `internal/server/table.go`: Create `GetMaxRaise(seatIndex int, seats []*Seat) int` method
  - `internal/server/table.go`: Create `ValidateRaise(seatIndex, amount, stack int, seats []*Seat) error` method
  - `internal/server/table_test.go`: Add max-raise and validation tests
- **Tests to Write:**
  - `TestGetMaxRaise_LimitedByPlayerStack` - Max raise = player's remaining stack
  - `TestGetMaxRaise_LimitedByOpponentStack` - Max raise limited to smallest opponent stack to avoid side pots
  - `TestGetMaxRaise_HeadsUp` - Heads-up allows full stack
  - `TestGetMaxRaise_MultiPlayer` - Multi-player correctly limits to smallest stack
  - `TestValidateRaise_BelowMinimum` - Returns error when amount < min-raise
  - `TestValidateRaise_AboveMaximum` - Returns error "raise would create side pot" when amount > max-raise
  - `TestValidateRaise_ValidAmount` - Returns nil for valid raise amount
  - `TestValidateRaise_AllInBelowMin` - Allows all-in even if below min-raise
  - `TestValidateRaise_HeadsUp` - Validation works correctly in heads-up
- **Steps:**
  1. Write failing test `TestGetMaxRaise_LimitedByPlayerStack`
  2. Implement `GetMaxRaise()` to return player's remaining stack
  3. Run test to confirm it passes
  4. Write failing test `TestGetMaxRaise_LimitedByOpponentStack`
  5. Update `GetMaxRaise()` to check all opponents' stacks and return minimum
  6. Run test to confirm it passes
  7. Write failing tests for `ValidateRaise()` including side pot error
  8. Implement `ValidateRaise()` checking min/max bounds, allowing all-in exception, raising error for side pot scenarios
  9. Write tests for heads-up and multi-player scenarios
  10. Run all tests to confirm they pass

### **Phase 3: Raise Action Processing**
- **Objective:** Update ProcessAction and GetValidActions to handle raise actions with amounts
- **Files/Functions to Modify/Create:**
  - `internal/server/table.go`: Update `GetValidActions(seatIndex int, seats []*Seat) []string` to include "raise"
  - `internal/server/table.go`: Update `ProcessAction(seatIndex int, action string, amount *int, seats []*Seat) error` signature to accept amount
  - `internal/server/table.go`: Add raise case in ProcessAction to update CurrentBet, LastRaise, PlayerBets
  - `internal/server/table_test.go`: Add ProcessAction raise tests
- **Tests to Write:**
  - `TestGetValidActions_IncludesRaise` - Returns ["fold", "call", "raise"] when player has enough chips
  - `TestGetValidActions_NoRaiseWhenInsufficient` - Excludes "raise" when player can only call/fold
  - `TestGetValidActions_HeadsUp` - Returns correct actions in heads-up
  - `TestProcessAction_RaiseUpdatesBets` - Raise updates CurrentBet, LastRaise, PlayerBets, Pot correctly
  - `TestProcessAction_RaiseInvalidAmount` - Returns error for invalid raise amount
  - `TestProcessAction_RaiseAllIn` - Handles all-in raise correctly
  - `TestProcessAction_MultipleRaises` - Chain of raises updates LastRaise increment correctly
  - `TestProcessAction_RaiseHeadsUp` - Raise works correctly in heads-up
- **Steps:**
  1. Write failing test `TestGetValidActions_IncludesRaise`
  2. Update `GetValidActions()` to check if player can raise (stack > callAmount + minRaise)
  3. Run test to confirm it passes
  4. Write failing test `TestProcessAction_RaiseUpdatesBets`
  5. Update `ProcessAction()` signature to accept `amount *int` parameter
  6. Add "raise" case that validates amount, updates CurrentBet/LastRaise/PlayerBets/Pot
  7. Run test to confirm it passes
  8. Update existing ProcessAction tests to pass `nil` for amount parameter
  9. Write remaining raise processing tests including heads-up scenarios
  10. Run all tests to confirm they pass

### **Phase 4: Handler Protocol Updates**
- **Objective:** Update WebSocket message handlers to support raise amounts in payloads
- **Files/Functions to Modify/Create:**
  - `internal/server/handlers.go`: Add `Amount *int` field to `PlayerActionPayload`
  - `internal/server/handlers.go`: Add `MinRaise int` and `MaxRaise int` fields to `ActionRequestPayload`
  - `internal/server/handlers.go`: Update `HandlePlayerAction()` to extract amount and pass to ProcessAction
  - `internal/server/handlers.go`: Update `BroadcastActionRequest()` to calculate and include MinRaise/MaxRaise
  - `internal/server/handlers_test.go`: Add handler tests for raise payloads
- **Tests to Write:**
  - `TestHandlePlayerAction_RaiseWithAmount` - Handler extracts amount and calls ProcessAction correctly
  - `TestHandlePlayerAction_RaiseMissingAmount` - Returns error when raise lacks amount
  - `TestBroadcastActionRequest_IncludesMinMaxRaise` - ActionRequest payload includes minRaise and maxRaise fields
  - `TestBroadcastActionRequest_MinMaxCalculation` - MinRaise and MaxRaise calculated correctly for multi-player and heads-up
- **Steps:**
  1. Write failing test `TestHandlePlayerAction_RaiseWithAmount`
  2. Add `Amount *int` field to `PlayerActionPayload` struct
  3. Update `HandlePlayerAction()` to extract amount from payload and pass to ProcessAction
  4. Update all ProcessAction calls in handlers to pass amount parameter
  5. Run test to confirm it passes
  6. Write failing test `TestBroadcastActionRequest_IncludesMinMaxRaise`
  7. Add `MinRaise int` and `MaxRaise int` fields to `ActionRequestPayload`
  8. Update `BroadcastActionRequest()` to call GetMinRaise/GetMaxRaise and include in payload
  9. Run test to confirm it passes
  10. Run all backend tests to confirm they pass

### **Phase 5: Frontend Protocol and State**
- **Objective:** Update frontend types and WebSocket handling to support raise amounts and min/max bounds
- **Files/Functions to Modify/Create:**
  - `frontend/src/hooks/useWebSocket.ts`: Add `amount?: number` to PlayerActionPayload interface
  - `frontend/src/hooks/useWebSocket.ts`: Add `minRaise?: number` and `maxRaise?: number` to ActionRequest interface
  - `frontend/src/hooks/useWebSocket.ts`: Update message parsing to extract minRaise/maxRaise from action_request
  - `frontend/src/hooks/useWebSocket.ts`: Update sendAction to accept optional amount parameter
  - `frontend/src/hooks/useWebSocket.test.ts`: Add tests for raise protocol
- **Tests to Write:**
  - `test('parses minRaise and maxRaise from action_request')` - State includes minRaise/maxRaise when present
  - `test('sendAction includes amount for raise')` - Raise action sends amount in payload
  - `test('sendAction omits amount for fold/check/call')` - Other actions don't include amount
- **Steps:**
  1. Write failing test for parsing minRaise/maxRaise from action_request
  2. Add `minRaise?: number` and `maxRaise?: number` to ActionRequest interface
  3. Update action_request message handler to extract and store minRaise/maxRaise in state
  4. Run test to confirm it passes
  5. Write failing test for sendAction with amount
  6. Add `amount?: number` to PlayerActionPayload interface
  7. Update `sendAction(action: string, amount?: number)` signature
  8. Include amount in WebSocket payload when provided
  9. Run test to confirm it passes
  10. Run all frontend tests to confirm they pass

### **Phase 6: Raise UI Components**
- **Objective:** Add raise button, number input, and preset buttons (Min/Pot/All-in) to frontend action bar
- **Files/Functions to Modify/Create:**
  - `frontend/src/components/TableView.tsx`: Add raise UI section to action bar
  - `frontend/src/components/TableView.tsx`: Add state for raise amount input
  - `frontend/src/components/TableView.tsx`: Add handlers for preset buttons (min, pot, all-in)
  - `frontend/src/components/TableView.tsx`: Update handleAction to pass amount for raises
  - `frontend/src/styles/TableView.css`: Add styles for raise controls
  - `frontend/src/components/TableView.test.tsx`: Add raise UI tests
- **Tests to Write:**
  - `test('shows Raise button when raise is valid action')` - Raise button appears when "raise" in validActions
  - `test('hides Raise button when raise not available')` - No raise button when only fold/check/call
  - `test('Min preset sets raise amount to minRaise')` - Min button sets input to minRaise value
  - `test('Pot preset calculates pot-sized raise')` - Pot button sets callAmount + pot (not including player's own bet)
  - `test('All-in preset sets raise to player stack')` - All-in button uses player's remaining stack
  - `test('Raise button disabled when amount invalid')` - Disabled if amount < min or > max
  - `test('Raise button sends action with amount')` - Click calls sendAction("raise", amount)
  - `test('amount input validates min/max bounds')` - Input shows error/warning for invalid amounts
- **Steps:**
  1. Write failing test for Raise button visibility
  2. Add Raise button to action bar, conditionally rendered when "raise" in validActions
  3. Run test to confirm it passes
  4. Write failing tests for preset buttons
  5. Add state for raiseAmount, number input field, and three preset buttons (Min/Pot/All-in)
  6. Implement preset button handlers: Min=minRaise, Pot=callAmount+pot, All-in=playerStack
  7. Run tests to confirm they pass
  8. Write failing test for Raise button sending amount
  9. Update handleAction to accept optional amount parameter and pass to sendAction
  10. Wire Raise button onClick to call handleAction("raise", raiseAmount)
  11. Run test to confirm it passes
  12. Add CSS styles for raise controls (input, buttons, layout)
  13. Run all frontend tests and do manual verification
  14. Run build to confirm no errors

**Design Decisions:**

1. **All-in below min-raise:** YES - all-in allowed at any amount (standard poker rules)
2. **Pot-sized calculation:** callAmount + existing pot (does NOT include player's own bet this round)
3. **UI input type:** Number input + preset buttons (no slider for now)
4. **Side pot rejection:** Raise error with message when raise would create side pot
5. **Heads-up support:** All logic must work correctly for both heads-up and multi-player games
