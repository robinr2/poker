## Phase 1 Complete: Write Failing Tests for IsBettingRoundComplete

Successfully created comprehensive test suite that exposes the all-in betting round completion bug across 2-player, 3-player, and multi-player scenarios.

**Files created/changed:**
- internal/server/table_test.go

**Functions created/changed:**
- None (test-only phase)

**Tests created/changed:**
- TestIsBettingRoundComplete_TwoPlayerBothAllInUnequalStacks (FAILS - exposes bug)
- TestIsBettingRoundComplete_TwoPlayerOneAllInOneMatched (PASSES - control test)
- TestIsBettingRoundComplete_ThreePlayerTwoAllInOneActive (FAILS - exposes bug)
- TestIsBettingRoundComplete_ThreePlayerAllDifferentStacks (FAILS - exposes bug)
- TestIsBettingRoundComplete_MultiPlayerSomeAllInSomeFolded (FAILS - exposes bug)
- TestIsBettingRoundComplete_AllPlayersAllIn (FAILS - exposes bug)

**Review Status:** APPROVED

**Git Commit Message:**
```
test: add failing tests for all-in betting round completion bug

- Add 6 test cases exposing bug where IsBettingRoundComplete doesn't skip all-in players
- Cover 2-player scenarios with unequal all-in stacks (900 vs 1000)
- Cover 3-player scenarios with mixed all-in and active players
- Cover multi-player scenarios with 5+ players in various states
- Cover edge case where all remaining players are all-in
- Tests fail as expected: function returns false when all-in players can't match higher bets
```
