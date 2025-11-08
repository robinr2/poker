## Phase 3 Complete: Personalized Hole Cards & Card Counts

Successfully extended `table_state` messages to include personalized hole cards for seated players and card counts for all seats during active hands. Privacy is guaranteed: players only receive their own hole cards, never opponents'. Spectators receive card counts so they can render card backs.

**Files created/changed:**
- `internal/server/handlers.go`
- `internal/server/handlers_test.go`

**Functions created/changed:**
- `TableStatePayload` struct - Added `HoleCards map[int][]Card` field
- `TableStateSeat` struct - Added `CardCount *int` field
- `SendTableState()` - Now personalizes hole cards based on client's seat
- `filterHoleCardsForPlayer()` - New helper function to enforce privacy
- `broadcastTableState()` - Updated to use new personalized SendTableState

**Tests created/changed:**
- `TestTableStateIncludesHoleCardsForSeatedPlayer` - Verifies seated player receives their hole cards
- `TestTableStateOmitsHoleCardsForUnseatedPlayer` - Verifies spectators don't receive hole cards
- `TestTableStateHoleCardsPrivacy` - Verifies multi-player privacy (each sees only their own cards)
- `TestTableStateCardCountsForSpectators` - Verifies card counts populated for all seats

**Review Status:** APPROVED ✅

**Git Commit Message:**
```
feat: Add personalized hole cards and card counts to table_state

- Include player's own hole cards in table_state during active hands
- Add card counts for all seats so spectators can render card backs
- Implement filterHoleCardsForPlayer() to guarantee privacy
- Extend TableStatePayload with HoleCards map and TableStateSeat with CardCount
- Add 4 comprehensive tests covering seated players, spectators, and privacy
```
