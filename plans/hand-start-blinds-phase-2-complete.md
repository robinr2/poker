## Phase 2 Complete: Dealer Button & Blind Position Logic

Successfully implemented dealer button rotation and blind position calculation with proper heads-up rules. The dealer rotates clockwise through active players, and blind positions are calculated correctly for both heads-up (dealer=SB) and normal (3+) games.

**Files created/changed:**
- internal/server/table.go
- internal/server/table_test.go

**Functions created/changed:**
- Table.DealerSeat (new field, *int type)
- Table.NextDealer() (new method)
- Table.GetBlindPositions() (new method)

**Tests created/changed:**
- TestNextDealerFirstHand (new test)
- TestNextDealerRotation (new test)
- TestNextDealerSkipsWaiting (new test)
- TestGetBlindPositionsNormal (new test)
- TestGetBlindPositionsHeadsUp (new test)
- TestGetBlindPositionsInsufficientPlayers (new test)
- TestGetBlindPositionsScatteredSeats (new test)
- TestGetBlindPositionsInvalidDealer (new test)

**Review Status:** APPROVED (after revision)

**Initial Review Findings:**
- Critical Issue: GetBlindPositions() didn't validate dealer seat was active
- Major Issue: Missing test coverage for scattered active seats and invalid dealer
- Fixes Applied: Added validation to return error for inactive dealer seat, added 2 comprehensive tests

**Git Commit Message:**
```
feat: implement dealer rotation and blind position logic

- Add DealerSeat field to Table struct for tracking current dealer
- Implement NextDealer() to rotate dealer clockwise through active players
- Implement GetBlindPositions() with heads-up exception (dealer=SB for 2 players)
- Handle normal case (3+ players): SB after dealer, BB after SB
- Skip non-active seats during rotation
- Validate dealer seat is active before calculating blind positions
- Add comprehensive tests for rotation, heads-up, scattered seats, edge cases
```
