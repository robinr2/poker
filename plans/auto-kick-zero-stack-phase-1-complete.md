## Phase 1 Complete: Backend - Implement Bust-Out Notification System

Successfully replaced silent `handleBustOutsLocked()` with proper notification system that sends `seat_cleared` messages, updates sessions, and broadcasts state changes to all clients.

**Files created/changed:**
- `internal/server/table.go`
- `internal/server/handlers.go`
- `internal/server/table_test.go`

**Functions created/changed:**
- `handleBustOutsWithNotificationsLocked()` (new) - Collects busted player tokens before clearing seats
- `handleBustOutNotifications()` (new) - Sends notifications, updates sessions, broadcasts state
- `HandleShowdown()` (modified) - Calls new methods at both bust-out points

**Tests created/changed:**
- `TestHandleBustOutsWithNotificationsLocked_SinglePlayerBusted` - Verify single bust-out detected and cleared
- `TestHandleBustOutsWithNotificationsLocked_MultiplePlayersBusted` - Verify multiple simultaneous bust-outs
- `TestHandleBustOutsWithNotificationsLocked_NoBustOuts` - Verify no false positives

**Review Status:** APPROVED

**Implementation Highlights:**
- ✅ Full TDD approach - tests written first, saw them fail, then implemented to pass
- ✅ Thread-safe design - tokens collected under lock, notifications sent after unlock
- ✅ Robust error handling - gracefully handles nil clients and disconnected players
- ✅ Follows existing patterns - matches `HandleLeaveTable` flow
- ✅ Zero regressions - all 277 backend tests pass (3 new tests added)

**Git Commit Message:**
```
feat: Auto-kick players with zero stack after showdown

- Add handleBustOutsWithNotificationsLocked() to collect busted player tokens
- Add handleBustOutNotifications() to send seat_cleared and update sessions
- Modify HandleShowdown() to call new notification methods at both bust-out points
- Send seat_cleared WebSocket messages to busted clients
- Update sessions to clear TableID and SeatIndex for busted players
- Broadcast table_state and lobby_state after bust-outs
- Handle edge cases: nil clients, disconnected players
- Add 3 TDD tests for bust-out notification system
- All 277 backend tests passing
```
