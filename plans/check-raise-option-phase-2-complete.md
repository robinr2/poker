## Phase 2 Complete: Fix GetValidActions to Include Raise

Updated GetValidActions() to include "raise" when player can check (if they have sufficient chips). This fixes the bug where players who have matched the current bet only see check/fold options.

**Files created/changed:**
- `internal/server/table.go` - Modified GetValidActions method
- `internal/server/table_test.go` - Fixed existing test

**Functions created/changed:**
- `GetValidActions()` - Added chip-checking logic when callAmount == 0 to include raise option

**Tests created/changed:**
- Fixed `TestGetValidActions_CanCheck` - Updated to properly represent insufficient chips scenario

**Review Status:** APPROVED

All Phase 1 tests now pass:
- ✅ TestGetValidActions_CanCheckAndRaise_Preflop
- ✅ TestGetValidActions_CanCheckAndRaise_Postflop  
- ✅ TestGetValidActions_CanCheckOnly_AllIn
- ✅ All 228 backend tests passing with no regressions

**Git Commit Message:**
```
fix: Add raise option when player can check

- Update GetValidActions to check chips and include raise when callAmount == 0
- Calculate chipsNeeded = minRaise - PlayerBets[seatIndex]
- Return ["check", "fold", "raise"] if player has sufficient chips
- Fix TestGetValidActions_CanCheck to properly test insufficient chips scenario
- Fixes bug where BB in unopened pot and postflop players only see check/fold
```
