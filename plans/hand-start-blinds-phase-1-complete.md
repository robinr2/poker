## Phase 1 Complete: Game State Structures

Successfully added foundational data structures for tracking poker game state: Card type with 52-card deck generation, Hand struct for managing dealer/blinds/pot/cards, and Stack field on Seat for chip tracking (1000 starting stack).

**Files created/changed:**
- internal/server/table.go
- internal/server/table_test.go

**Functions created/changed:**
- Card (new struct with Rank, Suit fields)
- Card.String() (new method)
- NewDeck() (new function)
- Hand (new struct)
- Seat.Stack (new field)
- Table.AssignSeat() (modified to initialize Stack to 1000)
- Table.ClearSeat() (modified to reset Stack to 0)

**Tests created/changed:**
- TestCardString (new test)
- TestNewDeck (new test)
- TestHandInitialization (new test)
- TestSeatWithStack (new test)
- TestTableClearSeatResetStack (new test)

**Review Status:** APPROVED (after revision)

**Initial Review Finding:**
- Issue: ClearSeat() didn't reset Stack field to 0
- Fix Applied: Added Stack = 0 to ClearSeat() method
- Additional Test: TestTableClearSeatResetStack verifies Stack reset behavior

**Git Commit Message:**
```
feat: add game state structures for hand start and blinds

- Add Card type with 52-card deck generation (NewDeck function)
- Add Hand struct to track dealer position, blinds, pot, deck, and hole cards
- Add Stack field to Seat struct (1000 starting chips)
- Initialize Stack to 1000 when assigning seats
- Reset Stack to 0 when clearing seats
- Add comprehensive tests for all new structures
```
