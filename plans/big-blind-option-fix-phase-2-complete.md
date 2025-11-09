## Phase 2 Complete: Update Betting Round Completion Logic

Updated IsBettingRoundComplete() to respect the BigBlindHasOption flag, ensuring the big blind gets their option to close preflop betting in an unopened pot. All tests pass with no regressions.

**Files created/changed:**
- internal/server/table.go
- internal/server/table_test.go

**Functions created/changed:**
- IsBettingRoundComplete (modified to check BigBlindHasOption flag)

**Tests created/changed:**
- TestIsBettingRoundComplete_PreflopUnopenedPot_BBHasOption (new)
- TestIsBettingRoundComplete_PreflopAfterBBChecks_RoundComplete (new)
- TestIsBettingRoundComplete_PreflopAfterRaise_BBOptionGone (new)
- TestIsBettingRoundComplete_PostflopNoSpecialCase (new)

**Review Status:** APPROVED

**Git Commit Message:**
```
fix: Add big blind option check to round completion logic

- Update IsBettingRoundComplete to check BigBlindHasOption flag
- Prevent round from completing when BB has the option in unopened pots
- Add 4 comprehensive tests for BB option completion scenarios
- All 241 backend tests passing with no regressions
```
