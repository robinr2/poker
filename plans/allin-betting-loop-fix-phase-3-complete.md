## Phase 3 Complete: Fix GetValidActions and Test All Streets

Successfully fixed `GetValidActions()` to return an empty array when player stack = 0, preventing all-in players from receiving action prompts. Added comprehensive tests across all betting streets.

**Files created/changed:**
- internal/server/table.go
- internal/server/table_test.go

**Functions created/changed:**
- `GetValidActions()` (lines 1152-1156) - Added early return for stack = 0

**Tests created/changed:**
- TestGetValidActions_AllInPlayerZeroStackPreflop - Preflop all-in scenario
- TestGetValidActions_AllInPlayerZeroStackFlop - Flop all-in scenario
- TestGetValidActions_AllInPlayerZeroStackTurn - Turn all-in scenario
- TestGetValidActions_AllInPlayerZeroStackRiver - River all-in scenario
- TestGetValidActions_AllInPlayerWithCallAmount - All-in with call pending
- TestGetValidActions_AllInPlayerWithRaise - All-in with raise available

**Review Status:** APPROVED

**Git Commit Message:**
```
fix: Return empty actions for all-in players in GetValidActions

- Add early return when playerStack == 0 to prevent action prompts
- Prevents all-in players from receiving call/check/fold/raise options
- Add 6 comprehensive tests covering all streets and edge cases
- All 323 backend tests pass with no regressions
```
