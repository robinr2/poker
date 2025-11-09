## Phase 2 Complete: Street Progression Trigger Logic

Implemented automatic street progression that advances hands from preflop through river, dealing board cards and resetting betting state at each transition.

**Files created/changed:**
- internal/server/table.go
- internal/server/table_test.go

**Functions created/changed:**
- `Hand.AdvanceToNextStreet()` - Orchestrates street transitions by dealing appropriate board cards and resetting betting state
- Hand struct - Uses existing `Street` and `AdvanceStreet()` for state management

**Tests created/changed:**
- `TestHand_AdvanceToNextStreet_PreflopToFlop`
- `TestHand_AdvanceToNextStreet_FlopToTurn`
- `TestHand_AdvanceToNextStreet_TurnToRiver`
- `TestHand_AdvanceToNextStreet_RiverNoAdvance`
- `TestHand_AdvanceToNextStreet_ErrorIfDeckExhausted`

**Review Status:** APPROVED

**Git Commit Message:**
```
feat: add automatic street progression logic

- Implement AdvanceToNextStreet() method to handle preflop→flop→turn→river transitions
- Deal board cards automatically when streets advance
- Reset betting state at each transition using existing AdvanceStreet() method
- Add comprehensive tests for street progression and edge cases
```
