## Phase 1 Complete: Add Big Blind Option State Tracking

Successfully added state tracking for the big blind option privilege. The BigBlindHasOption flag is now properly managed throughout the hand lifecycle, setting the foundation for Phase 2 to utilize this state in betting round completion logic.

**Files created/changed:**
- `internal/server/table.go`
- `internal/server/table_test.go`

**Functions created/changed:**
- Hand struct - Added `BigBlindHasOption bool` field (line 51)
- `StartHand()` - Sets flag to true when hand starts (line 483)
- `ProcessAction()` - Clears flag when BB acts or on any raise (lines 979, 995, 1021, 1070)
- `AdvanceStreet()` - Clears flag when advancing to next street (line 1177)

**Tests created/changed:**
- `TestHand_BigBlindHasOption_InitiallyTrue` (new)
- `TestHand_BigBlindHasOption_ClearedWhenBBChecks` (new)
- `TestHand_BigBlindHasOption_ClearedWhenBBRaises` (new)
- `TestHand_BigBlindHasOption_ClearedOnAnyRaise` (new)
- `TestHand_BigBlindHasOption_ClearedOnStreetAdvance` (new)

**Review Status:** APPROVED ✅

**Git Commit Message:**
```
feat: Add big blind option state tracking

- Add BigBlindHasOption bool field to Hand struct to track BB preflop privilege
- Set flag to true in StartHand() for every new hand
- Clear flag when BB acts with any action (fold, check, call, raise)
- Clear flag when any player raises (BB loses option on any raise)
- Clear flag in AdvanceStreet() when moving to postflop streets
- Add 5 comprehensive tests covering all flag state transitions
```
