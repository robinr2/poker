## Phase 2 Complete: Fix IsBettingRoundComplete

Successfully fixed the `IsBettingRoundComplete()` function to correctly handle all-in players by skipping players with stack = 0 when checking if all bets have been matched.

**Files created/changed:**
- internal/server/table.go

**Functions created/changed:**
- `IsBettingRoundComplete()` (lines 1552-1565)

**Tests created/changed:**
- No new tests (Phase 1 tests used for verification)
- All 5 failing tests from Phase 1 now pass:
  - TestIsBettingRoundComplete_TwoPlayerBothAllInUnequalStacks
  - TestIsBettingRoundComplete_ThreePlayerTwoAllInOneActive
  - TestIsBettingRoundComplete_ThreePlayerAllDifferentStacks
  - TestIsBettingRoundComplete_MultiPlayerSomeAllInSomeFolded
  - TestIsBettingRoundComplete_AllPlayersAllIn

**Review Status:** APPROVED

**Git Commit Message:**
```
fix: Skip all-in players in IsBettingRoundComplete check

- Add stack == 0 check to skip all-in players when validating bet matching
- Fixes infinite loop bug when players go all-in with unequal stacks
- All 5 Phase 1 test cases now pass with no regressions
```
