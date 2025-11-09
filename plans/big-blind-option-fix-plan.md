## Plan: Big Blind Option Fix

Fix the preflop betting bug where the big blind doesn't get their option to check or raise when facing an unopened pot (when small blind calls with no raises). This implements the standard poker rule that the big blind closes preflop action.

**Phases: 3**

### **Phase 1: Add Big Blind Option State Tracking**
- **Objective:** Add state to track when the big blind has the option to close preflop betting, and set/clear this flag appropriately.
- **Files/Functions to Modify/Create:**
  - `internal/server/table.go` - Add `BigBlindHasOption bool` field to Hand struct
  - `internal/server/table.go` - Set flag in `StartHand()` method
  - `internal/server/table.go` - Clear flag in `ProcessAction()` when BB acts or on any raise
  - `internal/server/table.go` - Clear flag in `AdvanceStreet()` when moving past preflop
  - `internal/server/table_test.go` - Add tests for flag state management
- **Tests to Write:**
  - `TestHand_BigBlindHasOption_InitiallyTrue`
  - `TestHand_BigBlindHasOption_ClearedWhenBBChecks`
  - `TestHand_BigBlindHasOption_ClearedWhenBBRaises`
  - `TestHand_BigBlindHasOption_ClearedOnAnyRaise`
  - `TestHand_BigBlindHasOption_ClearedOnStreetAdvance`
- **Steps:**
  1. Write failing tests for BigBlindHasOption flag behavior
  2. Add `BigBlindHasOption bool` field to Hand struct
  3. Set `BigBlindHasOption = true` in `StartHand()` method (only on preflop)
  4. In `ProcessAction()`, clear flag when: (a) actor is BigBlindSeat, (b) action is "raise"
  5. In `AdvanceStreet()`, ensure flag is cleared when leaving preflop
  6. Run tests to verify flag management works correctly
  7. Run linter and formatter

### **Phase 2: Update Betting Round Completion Logic**
- **Objective:** Modify IsBettingRoundComplete() to respect the big blind option, preventing round completion until BB has acted on preflop when they have the option.
- **Files/Functions to Modify/Create:**
  - `internal/server/table.go` - Modify `IsBettingRoundComplete()` to check BigBlindHasOption flag
  - `internal/server/table_test.go` - Add tests for round completion with BB option
- **Tests to Write:**
  - `TestIsBettingRoundComplete_PreflopUnopenedPot_BBHasOption`
  - `TestIsBettingRoundComplete_PreflopAfterBBChecks_RoundComplete`
  - `TestIsBettingRoundComplete_PreflopAfterRaise_BBOptionGone`
  - `TestIsBettingRoundComplete_PostflopNoSpecialCase`
- **Steps:**
  1. Write failing tests demonstrating BB option prevents round completion
  2. In `IsBettingRoundComplete()`, add check: If `BigBlindHasOption == true` AND `ActedPlayers[BigBlindSeat] == false`, return false
  3. Ensure this check only applies on preflop (flag should only be true on preflop anyway)
  4. Run tests to verify round completion logic respects BB option
  5. Run linter and formatter

### **Phase 3: Integration Testing & Validation**
- **Objective:** Add end-to-end tests covering complete scenarios where BB gets their option, and verify the fix works in realistic game flows.
- **Files/Functions to Modify/Create:**
  - `internal/server/table_test.go` - Add integration tests for BB option scenarios
  - `internal/server/websocket_integration_test.go` - Add WebSocket test for BB option flow
- **Tests to Write:**
  - `TestHandFlow_PreflopSBCallsBBChecks_FlopDealt`
  - `TestHandFlow_PreflopSBCallsBBRaises_ActionContinues`
  - `TestHandFlow_HeadsUpSBCallsBBOption`
  - `TestWebSocketFlow_BBGetsActionAfterSBCalls`
- **Steps:**
  1. Write failing integration tests for complete BB option scenarios
  2. Run tests - they should pass with Phase 1 & 2 implementations
  3. If any fail, debug and fix the specific edge case
  4. Manually test in browser: SB calls, verify BB gets action with check/raise options
  5. Run all tests (backend + frontend) to verify no regressions
  6. Run linter and formatter

---

**Implementation Decisions:**

1. **BigBlindHasOption Flag**: Use explicit boolean flag (cleanest, most readable)
2. **Flag Scope**: Only true on preflop, cleared on BB action or any raise
3. **Heads-Up**: Works automatically since BB is still tracked as BigBlindSeat
4. **Postflop**: Flag never set, no impact on flop/turn/river betting
