## Phase 1 Complete: Min-Raise Computation and Validation

Successfully implemented backend logic to compute minimum valid raise amounts based on poker rules (min-raise = current bet + last raise increment). The implementation correctly initializes LastRaise to the big blind amount at hand start and resets it to zero when advancing to new streets.

**Files created/changed:**
- internal/server/table.go
- internal/server/table_test.go

**Functions created/changed:**
- `Hand.LastRaise` (new field)
- `Hand.GetMinRaise()` (new method)
- `Hand.NewHand()` (updated to initialize LastRaise)
- `Hand.AdvanceStreet()` (updated to reset LastRaise)

**Tests created/changed:**
- `TestGetMinRaise_Preflop` (new)
- `TestGetMinRaise_AfterRaise` (new)
- `TestGetMinRaise_AfterMultipleRaises` (new)
- `TestGetMinRaise_PostFlop` (new)
- `TestGetMinRaise_HeadsUp` (new)
- `TestNewHand_InitializesLastRaise` (new)
- `TestAdvanceStreet_ResetsLastRaise` (new)

**Review Status:** APPROVED

**Git Commit Message:**
```
feat: add min-raise computation for poker raises

- Add LastRaise field to Hand struct to track raise increment
- Implement GetMinRaise() method returning CurrentBet + LastRaise
- Initialize LastRaise to big blind amount in NewHand()
- Reset LastRaise to zero in AdvanceStreet() for post-flop streets
- Add 7 comprehensive tests covering preflop, post-flop, multiple raises, and heads-up scenarios
```
