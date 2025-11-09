## Plan: Add Raise Option When Player Can Check

Fix the bug where players who have matched the current bet (can check) are only shown "check" and "fold" buttons, when they should also have the option to raise. This affects the big blind in unopened pots and any player on postflop streets.

**Phases: 3**

### **Phase 1: Add Failing Tests for Check-Raise Scenario**
- **Objective:** Write tests demonstrating that players who can check should also have raise as an available action
- **Files/Functions to Modify/Create:**
  - `internal/server/table_test.go` - Add tests for check-raise scenarios
- **Tests to Write:**
  - `TestGetValidActions_CanCheckAndRaise_Preflop` - BB in unopened pot should have check/fold/raise
  - `TestGetValidActions_CanCheckAndRaise_Postflop` - Player on flop should have check/fold/raise
  - `TestGetValidActions_CanCheckOnly_AllIn` - Player with insufficient chips can only check/fold
- **Steps:**
  1. Write test for BB in unopened pot (has matched bet, should see raise option)
  2. Write test for postflop player (all bets equal, should see raise option)
  3. Write test for player without enough chips to raise (should only see check/fold)
  4. Run tests - they should fail showing only check/fold returned
  5. Run linter and formatter

### **Phase 2: Fix GetValidActions to Include Raise**
- **Objective:** Update GetValidActions() to include "raise" when player can check (if they have sufficient chips)
- **Files/Functions to Modify/Create:**
  - `internal/server/table.go` - Modify `GetValidActions()` method
- **Tests to Write:**
  - None (using Phase 1 tests)
- **Steps:**
  1. In `GetValidActions()`, when callAmount == 0, calculate minRaise chips needed
  2. If player has enough chips for minRaise, return ["check", "fold", "raise"]
  3. Otherwise return ["check", "fold"]
  4. Run Phase 1 tests - they should now pass
  5. Run all tests to ensure no regressions
  6. Run linter and formatter

### **Phase 3: Integration Testing**
- **Objective:** Verify fix works in realistic scenarios including unopened pot BB and postflop action
- **Files/Functions to Modify/Create:**
  - `internal/server/table_test.go` - Add integration test
  - `internal/server/websocket_integration_test.go` - Add WebSocket test
- **Tests to Write:**
  - `TestHandFlow_BBCanRaiseUnopenedPot` - Full hand where BB raises unopened pot
  - `TestWebSocketFlow_CheckRaiseOnFlop` - WebSocket test with flop check-raise
- **Steps:**
  1. Write integration test for BB raising unopened pot preflop
  2. Write integration test for postflop check-raise scenario
  3. Write WebSocket test verifying BB gets raise option and can execute it
  4. Run all tests (should all pass)
  5. Run linter and formatter

---

**Implementation Notes:**

1. **Chip Calculation**: When player can check (callAmount == 0), they need `minRaise - PlayerBets[seatIndex]` chips to raise
2. **Consistent Logic**: Use same chip-checking logic as the callAmount > 0 branch
3. **All-In Edge Case**: Players without sufficient chips should only get check/fold
4. **Button Order**: Return actions in order ["check", "fold", "raise"] or ["check", "fold"] for UX consistency
