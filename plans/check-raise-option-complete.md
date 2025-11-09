## Plan Complete: Add Raise Option When Player Can Check

Successfully fixed the bug where players who have matched the current bet (can check) only see "check" and "fold" buttons. Now players correctly see "raise" as an option when they have sufficient chips, enabling proper check-raise scenarios both preflop (BB in unopened pots) and postflop.

**Phases Completed:** 3 of 3
1. ✅ Phase 1: Add Failing Tests for Check-Raise Scenario
2. ✅ Phase 2: Fix GetValidActions to Include Raise
3. ✅ Phase 3: Integration Testing

**All Files Created/Modified:**
- `internal/server/table.go` - Modified GetValidActions() method
- `internal/server/table_test.go` - Added 5 unit tests, fixed 1 existing test
- `internal/server/websocket_integration_test.go` - Added 1 WebSocket integration test

**Key Functions/Classes Added:**
- Modified `GetValidActions()` in `internal/server/table.go` to properly check player chips when callAmount == 0 and include "raise" option

**Test Coverage:**
- Total tests written: 6 (3 unit + 2 integration + 1 WebSocket)
- Fixed tests: 1 (TestGetValidActions_CanCheck)
- All tests passing: ✅ (241 backend tests)

**Bug Fixed:**
The root cause was in `GetValidActions()` at lines 853-854. When `callAmount == 0` (player can check), the function immediately returned `["check", "fold"]` without checking if the player had chips to raise. Now the function calculates `chipsNeeded = minRaise - PlayerBets[seatIndex]` and includes "raise" when the player has sufficient chips.

**Recommendations for Next Steps:**
- Manual browser testing to verify UI correctly displays and handles the raise button in check-raise scenarios
- Consider adding frontend tests for the raise button appearing when player can check
- Test both preflop BB unopened pot and postflop check-raise flows in the browser
