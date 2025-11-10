## Plan: Auto-Kick Players with Zero Stack

Automatically remove players from table when their stack reaches zero, with proper session cleanup and client notifications (same behavior as manual leave).

**Phases: 3**

### **Phase 1: Backend - Implement Bust-Out Notification System**

- **Objective:** Replace silent `handleBustOutsLocked()` with proper notification system that sends `seat_cleared` messages and updates sessions
- **Files/Functions to Modify/Create:**
  - `internal/server/table.go`:
    - Create new method `handleBustOutsWithNotifications()` (after line 313)
    - Replace calls to `handleBustOutsLocked()` at lines 198 and 256
  - `internal/server/server.go`:
    - May need to add helper method to get client by token if not exists
- **Tests to Write:**
  - `TestHandleBustOutsWithNotifications_SinglePlayerBusted` - Verify seat cleared, session updated, seat_cleared sent, table_state broadcast
  - `TestHandleBustOutsWithNotifications_MultiplePlayersBusted` - Verify multiple simultaneous bust-outs handled correctly
  - `TestHandleBustOutsWithNotifications_NoClients` - Verify no crash when busted player has no active client connection
- **Steps:**
  1. Write tests first (TDD): Create test cases that expect `seat_cleared` messages sent to busted players, sessions updated, and table_state broadcast to remaining players
  2. Run tests to see them fail (no implementation yet)
  3. Implement `handleBustOutsWithNotifications()` method in `table.go`:
     - Collect tokens of players with stack == 0 BEFORE clearing seats
     - Call `handleBustOutsLocked()` to clear seats locally
     - For each busted token: find client, send `seat_cleared`, update session, log bust-out
     - Broadcast `table_state` to remaining players at table
     - Broadcast `lobby_state` to all clients
  4. Replace `handleBustOutsLocked()` calls at lines 198 and 256 with new method
  5. Run tests to confirm they pass
  6. Run full backend test suite to verify no regressions

### **Phase 2: Backend - Add Comprehensive Integration Tests**

- **Objective:** Verify auto-kick works in real showdown scenarios with pot distribution
- **Files/Functions to Modify/Create:**
  - `internal/server/table_test.go`:
    - Add integration test for all-in player losing at showdown
    - Add test for multiple all-in players busting simultaneously
- **Tests to Write:**
  - `TestShowdown_AllInPlayerBustsOut` - Player goes all-in, loses, verify auto-kicked
  - `TestShowdown_MultiplePlayersBustOut` - Multiple losers with zero stacks, verify all kicked
  - `TestShowdown_WinnerWithZeroStackNotKicked` - Edge case: player starts with 0 stack but wins (shouldn't be kicked)
- **Steps:**
  1. Write integration tests that simulate full hand with pot distribution and verify bust-outs
  2. Run tests to see them fail or pass (should pass if Phase 1 implementation is correct)
  3. If tests fail, debug and fix the bust-out notification logic
  4. Run full backend test suite to verify all 274+ tests still pass

### **Phase 3: Integration - Verify Frontend Handles Bust-Out Gracefully**

- **Objective:** Verify frontend correctly handles receiving `seat_cleared` mid-game without crashes
- **Files/Functions to Modify/Create:**
  - `frontend/src/hooks/useWebSocket.test.ts`:
    - Add test for receiving `seat_cleared` during active hand
  - `frontend/src/App.test.tsx`:
    - Add test for bust-out flow (player at table → busts → returns to lobby)
- **Tests to Write:**
  - `test_seat_cleared_during_active_hand` - Verify receiving `seat_cleared` mid-game clears state and returns to lobby
  - `test_bust_out_flow_complete` - Full flow: join table → play hand → bust out → verify lobby view
- **Steps:**
  1. Write frontend tests for bust-out scenarios
  2. Run tests to see if they pass (frontend already handles `seat_cleared`, should pass)
  3. If tests fail, update `useWebSocket.ts` or `App.tsx` to handle bust-out edge cases
  4. Run full frontend test suite to verify all 228+ tests pass
  5. Manual verification: Start local server, join table, simulate bust-out scenario

---

## Open Questions - RESOLVED

1. **Session cleanup timing:** Should we broadcast `lobby_state` after bust-outs? ✅ **YES** - Match `HandleLeaveTable` behavior
2. **Disconnected players:** Handle gracefully if busted player has no active connection? ✅ **YES** - Log warning and continue
3. **Multiple bust-outs:** Send individual `seat_cleared` messages or batch? ✅ **Individual** - For simplicity and consistency
