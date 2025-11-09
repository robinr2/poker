## Phase 2 Complete: Max-Raise and Side Pot Prevention

Successfully implemented logic to compute maximum valid raise amounts and validate raises to prevent side pot creation. The implementation correctly limits max raise to the minimum of the player's stack or the smallest opponent's stack, and provides clear error messages when raises would create side pots.

**Files created/changed:**
- internal/server/table.go
- internal/server/table_test.go

**Functions created/changed:**
- `Hand.GetMaxRaise()` (new method)
- `Hand.ValidateRaise()` (new method)

**Tests created/changed:**
- `TestGetMaxRaise_LimitedByPlayerStack` (new)
- `TestGetMaxRaise_LimitedByOpponentStack` (new)
- `TestGetMaxRaise_HeadsUp` (new)
- `TestGetMaxRaise_MultiPlayer` (new)
- `TestValidateRaise_BelowMinimum` (new)
- `TestValidateRaise_AboveMaximum` (new)
- `TestValidateRaise_ValidAmount` (new)
- `TestValidateRaise_AllInBelowMin` (new)
- `TestValidateRaise_HeadsUp` (new)

**Review Status:** APPROVED

**Git Commit Message:**
```
feat: add max-raise validation and side pot prevention

- Implement GetMaxRaise() to limit raises to smallest opponent stack
- Implement ValidateRaise() with min/max bounds checking
- Allow all-in at any amount (bypasses minimum raise requirement)
- Return error "raise would create side pot" when max exceeded
- Add 9 comprehensive tests covering heads-up and multi-player scenarios
```
