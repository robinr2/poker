## Phase 1 Complete: Backend Board Card Storage & Dealing

Added board card state management to the Hand struct with methods to deal flop (3 cards), turn (1 card), and river (1 card), following standard poker burn card rules.

**Files created/changed:**
- internal/server/table.go
- internal/server/table_test.go

**Functions created/changed:**
- `Hand.DealFlop()` - Burns 1 card and deals 3 cards to the board
- `Hand.DealTurn()` - Burns 1 card and deals 1 card to the board
- `Hand.DealRiver()` - Burns 1 card and deals 1 card to the board
- `Hand.StartHand()` - Modified to initialize BoardCards slice
- Hand struct - Added `BoardCards []Card` field

**Tests created/changed:**
- `TestHand_BoardCards_InitiallyEmpty`
- `TestHand_DealFlop_DealsThreeCards`
- `TestHand_DealFlop_BurnsCardBeforeDealing`
- `TestHand_DealFlop_ErrorsIfDeckExhausted`
- `TestHand_DealTurn_DealsOneCard`
- `TestHand_DealTurn_BurnsCardBeforeDealing`
- `TestHand_DealRiver_DealsOneCard`
- `TestHand_DealRiver_BurnsCardBeforeDealing`

**Review Status:** APPROVED

**Git Commit Message:**
```
feat: add backend board card storage and dealing methods

- Add BoardCards []Card field to Hand struct
- Implement DealFlop/DealTurn/DealRiver methods with burn card handling
- Add comprehensive tests for board card dealing and burn card verification
- Initialize BoardCards as empty slice in StartHand()
```
