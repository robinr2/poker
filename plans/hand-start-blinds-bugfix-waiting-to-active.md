## Bugfix: Player Status Transition from "waiting" to "active"

### Problem
After completing all 6 phases of the Hand Start & Blinds feature, the "Start Hand" button appeared in the UI but clicking it produced the error:
```
"insufficient active players to start hand: 0 active, need at least 2"
```

### Root Cause
Players joining tables were correctly assigned Status="waiting" (per Seating & Waiting plan Phase 2), but **no code existed** to transition them to Status="active". The `StartHand()` method only counted players with Status="active", resulting in 0 eligible players.

**Evidence:**
- `AssignSeat()` in table.go line 111: Sets Status="waiting" ✅
- `StartHand()` in table.go line 364: Only counts "active" players ❌
- **Missing**: Transition logic from "waiting" → "active"

### Solution Implemented
Modified `StartHand()` to transition all "waiting" players to "active" status at the beginning of hand start:

**File: `internal/server/table.go`**

1. **StartHand() - Added Step 0**: Transition "waiting" → "active"
   - Before: Started validation immediately
   - After: Loops through seats and sets Status="active" for any Status="waiting" players
   - Location: Lines 357-366 (new Step 0 before validation)

2. **CanStartHand() - Updated validation**: Check for "waiting" OR "active" players
   - Before: Only counted Status="active" players
   - After: Counts both "waiting" and "active" players
   - Rationale: `CanStartHand()` is called BEFORE `StartHand()` transitions players, so it needs to check for both statuses to determine if a hand CAN start

### Design Rationale
This fix aligns with the original plan intent:
- **Seating-Waiting Plan (Phase 2)**: Players start in "waiting" status when they sit down
- **Hand-Start-Blinds Plan (Phase 4, line 136)**: Plan mentions "status becomes relevant in Feature 4 when dealing cards"
- **Proper separation**: Players are "waiting" until they participate in a hand, then become "active"

This design allows for future features like:
- Spectator mode (players who join but don't play)
- Sit-out functionality (active players who want to skip hands)
- Tournament late registration (players waiting for next hand)

### Testing
- **All existing tests pass**: 80+ backend tests, 128 frontend tests
- **No test modifications needed**: Tests that set Status="active" directly still work correctly
- **Integration tests pass**: WebSocket broadcasts and card privacy work as expected

### Files Modified
- `internal/server/table.go`:
  - `StartHand()`: Added Step 0 to transition "waiting" → "active" (lines 357-366)
  - `CanStartHand()`: Updated to count "waiting" OR "active" players (lines 333-343)

### Next Steps
1. Manual testing: Verify "Start Hand" button now works with 2+ seated players
2. Test gameplay: Confirm dealer rotation, card dealing, and stack updates work correctly
3. If successful, commit this bugfix
4. Continue to Feature #5 (Betting Actions)
