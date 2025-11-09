## Phase 3 Complete: Raise Action Processing

Successfully extended GetValidActions and ProcessAction to handle raise actions with amounts. The implementation correctly validates raise amounts, updates game state (CurrentBet, LastRaise, PlayerBets, Pot), and supports both heads-up and multi-player scenarios with proper all-in handling.

**Files created/changed:**
- internal/server/table.go
- internal/server/table_test.go
- internal/server/handlers.go
- internal/server/handlers_test.go

**Functions created/changed:**
- `Hand.GetValidActions()` (updated signature and logic to include "raise")
- `Hand.ProcessAction()` (extended to handle "raise" action with variadic amount parameter)
- Updated calls to GetValidActions in handlers.go

**Tests created/changed:**
- `TestGetValidActions_IncludesRaise` (new)
- `TestGetValidActions_NoRaiseWhenInsufficient` (new)
- `TestGetValidActions_HeadsUp` (new)
- `TestProcessAction_RaiseUpdatesBets` (new)
- `TestProcessAction_RaiseInvalidAmount` (new)
- `TestProcessAction_RaiseAllIn` (new)
- `TestProcessAction_MultipleRaises` (new)
- `TestProcessAction_RaiseHeadsUp` (new)

**Review Status:** APPROVED

**Git Commit Message:**
```
feat: add raise action processing to game engine

- Extend GetValidActions() to include "raise" when player can afford min-raise
- Update ProcessAction() to handle raise with variadic amount parameter
- Correctly update CurrentBet, LastRaise increment, PlayerBets, and Pot
- Support all-in raises below minimum raise amount
- Add 8 comprehensive tests covering validation and state updates
- Update handler calls to pass new parameters
```
