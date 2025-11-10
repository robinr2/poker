# Future TODOs

This document tracks features and improvements that need to be implemented later, but are not part of current active work.

## Poker Gameplay Features

### Side Pots
- **Status:** Not yet implemented
- **Priority:** High (required for proper all-in scenarios with 3+ players)
- **Description:** When multiple players go all-in with different stack sizes, create side pots so that each player can only win chips they were able to match
- **Context:** Current pot accounting fix (pot-accounting-fix-plan.md) does not address side pots. The basic pot accounting model needs to be correct first before implementing side pot logic.
- **Implementation Notes:**
  - Side pots need to be calculated when players go all-in for different amounts
  - At showdown, each side pot is awarded separately starting from smallest to largest
  - Players can only win from pots they contributed to
- **Related Files:**
  - `internal/server/table.go` - Pot and PlayerBets tracking
  - Showdown settlement logic

### Other Future Items
(Add more items here as they come up)
