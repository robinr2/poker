# Bug Fixes Session Summary

This document summarizes all bugs fixed in this debugging session.

---

## Bug 1: Pot Displays as 0 During Preflop ✅ FIXED

**Problem**: The pot displayed as 0 during preflop betting, only updating on the first postflop action.

**Root Cause**: The `Pot` field only reflected committed bets from previous streets. During preflop, all bets were in `PlayerBets` but not yet added to `Pot`.

**Solution**: 
- Added `GetDisplayPot()` method to `Table` that returns `Pot + sum(PlayerBets)`
- Updated all broadcast locations to use `GetDisplayPot()` instead of raw `Pot`
- This ensures the UI always shows the total amount in play

**Files Modified**:
- `internal/server/table.go` - Added `GetDisplayPot()` method
- `internal/server/server.go` - Updated broadcasts to use `GetDisplayPot()`
- `internal/server/handlers.go` - Updated broadcasts to use `GetDisplayPot()`
- `internal/server/table_test.go` - Updated tests
- `internal/server/handlers_test.go` - Updated tests

**Tests**: ✅ All 189 tests passing

---

## Bug 2: Can't Raise After Large Preflop Raise ✅ FIXED

**Problem**: After raising to 600+ preflop, players could only check/fold on the flop (no raise option).

**Root Cause**: `AdvanceStreet()` incorrectly preserved `LastRaise` from preflop to postflop streets. This caused the raise validation logic to think raises were still happening from the 600 level.

**Solution**: 
- Modified `AdvanceStreet()` to always reset `LastRaise=20` (big blind) when advancing to postflop streets
- Preflop raise amounts no longer affect postflop minimum raise calculations
- Postflop streets now always start with minimum raise = big blind

**Files Modified**:
- `internal/server/table.go` - Updated `AdvanceStreet()` method (line ~570)
- `internal/server/table_test.go` - Fixed/updated tests to expect new behavior

**Tests**: ✅ All 189 tests passing

---

## Bug 3: Bust-Out Timing and Notifications ✅ FIXED

**Problem**: Players who lost all their chips were kicked immediately and couldn't see the final board state.

**Desired Behavior**: 
1. Player loses chips → sees showdown → stays seated
2. Next player clicks "start hand" → busted player gets kicked
3. Busted player receives notifications and returns to lobby
4. New hand starts with remaining players

**Solution Implemented**:

### Part 3A: Delay Kick Until Next Hand Start
- Modified `handleBustOutsWithNotificationsLocked()` to collect bust-out tokens but NOT clear seats or send notifications
- **Removed** immediate `handleBustOutNotifications()` calls from showdown code (lines 224, 294)
- Busted players remain seated with Stack==0 after showdown completes
- Modified `StartHand()` to kick players with Stack==0 at the beginning (lines 747-757)
- Kicking only happens when the next hand starts (when "start hand" button is clicked)

### Part 3B: Send Notifications to Busted Players (only at next hand start)
- Modified `handleBustOutNotifications()` to:
  1. Find the busted player's WebSocket client
  2. Send `seat_cleared` message to busted player (tells client they're no longer seated)
  3. Send `lobby_state` message to busted player (shows them the lobby again)
  4. Clear the player's session (remove table/seat info)
  5. Broadcast table state to remaining players (shows empty seats)
  6. Broadcast lobby state to all clients (updates table info)
- This function is now ONLY called from `StartHand()`, never from showdown

**Files Modified**:
- `internal/server/table.go` - Removed immediate notification calls from showdown, added kick logic at start of `StartHand()`
- `internal/server/handlers.go` - Completely rewrote `handleBustOutNotifications()` to send proper messages
- `internal/server/table_test.go` - Fixed 11+ tests to expect new behavior
- `internal/server/card_distribution_test.go` - Fixed test to give players chips

**Tests**: ✅ All 189 tests passing

---

## Bug 4: Start Hand Button Shows When Only One Player Remains ✅ FIXED

**Problem**: 
1. When two players went all-in and one lost all their money, the "Start Hand" button was still shown even though only one player remained
2. The button didn't show when players first joined the table
3. After a hand completed (e.g., A goes all-in, B calls, C folds, B wins), the button didn't appear even though B and C both had money

**Root Causes**:
1. Initial calculation only checked "active" players, not "waiting" players (new joiners)
2. After hand completion, no `table_state` message was sent to update the `canStartHand` flag on the frontend

**Desired Behavior**: 
- "Start Hand" button should only show when there are at least 2 players with chips (waiting or active status)
- After a bust-out that leaves only 1 player, no button should appear
- When players first join the table (before first hand), button should appear
- After any hand completes, button should appear if 2+ players remain

**Solution Implemented**:

### Backend: Add `canStartHand` Field
- Added `canStartHand` boolean field to `TableStatePayload` struct
- Calculates if hand can start: `playersWithChips >= 2 && table.CurrentHand == nil`
- Counts **both "waiting" and "active"** players (waiting = just joined, active = played at least one hand)
- Updated both `SendTableState()` and `sendPersonalizedTableState()` to populate this field
- **Added `broadcastTableState()` calls after hand completion** to send updated `canStartHand` to frontend

### Frontend: Use `canStartHand` Field
- Added `canStartHand?: boolean` to `GameState` interface in both `TableView.tsx` and `useWebSocket.ts`
- Updated button visibility logic: `showStartHandButton = isSeated && (!handInProgress || isHandComplete) && canStartHand`
- Updated `useWebSocket.ts` to extract and set `canStartHand` from `table_state` messages
- Button only appears when server confirms enough players are present

**Files Modified**:
- `internal/server/handlers.go` - Added `canStartHand` field to `TableStatePayload`, updated calculation to include "waiting" players
- `internal/server/table.go` - Added `broadcastTableState()` calls after all showdown completion paths (3 locations: early winner, multi-player showdown, no winners)
- `frontend/src/components/TableView.tsx` - Added `canStartHand` to `GameState`, updated button visibility logic
- `frontend/src/hooks/useWebSocket.ts` - Added `canStartHand` to `GameState` interface and extraction from `table_state` payload
- `frontend/src/components/TableView.test.tsx` - Updated 7 tests to include `canStartHand: true` in mock data

**Tests**: 
- ✅ Backend: All 189 tests passing
- ✅ Frontend: All 250 tests passing
- ✅ Build succeeds (both frontend and backend)

---

## Summary

All four bugs have been fixed:
1. ✅ Pot now displays correctly during preflop betting
2. ✅ Players can raise normally after large preflop raises
3. ✅ Busted players see final board state and receive proper kick notifications
4. ✅ "Start Hand" button only shows when 2+ players with chips are present

**Test Status**: 
- ✅ Backend: All 189 tests passing
- ✅ Frontend: All 250 tests passing

**Build Status**: ✅ Build succeeds

The game now handles bust-outs gracefully with proper timing, notifications, and UI state management.
