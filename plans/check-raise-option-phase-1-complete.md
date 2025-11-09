## Phase 1 Complete: Add Failing Tests for Check-Raise Scenario

Successfully added unit tests demonstrating that players who can check should also have raise as an available action. Tests initially failed (as expected), exposing the bug in GetValidActions.

**Files created/changed:**
- `internal/server/table_test.go` - Added 3 unit tests

**Functions created/changed:**
- None (Phase 1 added tests only, no production code changes)

**Tests created/changed:**
- `TestGetValidActions_CanCheckAndRaise_Preflop` (new) - Tests BB in unopened pot should get ["check", "fold", "raise"]
- `TestGetValidActions_CanCheckAndRaise_Postflop` (new) - Tests postflop player with matched bets should get ["check", "fold", "raise"]
- `TestGetValidActions_CanCheckOnly_AllIn` (new) - Tests player with insufficient chips should only get ["check", "fold"]

**Review Status:** APPROVED (as part of overall plan review)

**Git Commit Message:**
```
test: Add failing tests for check-raise option bug

- Add test for BB in unopened pot needing raise option
- Add test for postflop check-raise scenarios
- Add test for insufficient chips edge case (check/fold only)
- Tests demonstrate GetValidActions missing raise option when player can check
```
