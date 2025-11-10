## Phase 2 Complete: Clear Board Cards on New Hand Start (Optimistic Update)

Successfully implemented clearing of boardCards array when start_hand action is sent optimistically, preventing old board cards from persisting into the new hand.

**Files created/changed:**
- `frontend/src/hooks/useWebSocket.ts`
- `frontend/src/hooks/useWebSocket.test.ts`

**Functions created/changed:**
- `useWebSocket` - Modified `sendAction` function (line 614) to add `updated.boardCards = []` in optimistic update for `start_hand` action

**Tests created/changed:**
- `sendAction clears boardCards array when start_hand is sent` - Verifies boardCards is cleared to empty array
- `boardCards is empty array (not undefined) after start_hand action` - Verifies array semantics preserved

**Implementation Details:**

When `start_hand` action is sent, the optimistic update now clears:
```javascript
if (action === 'start_hand') {
  setGameState((prev) => {
    const updated = { ...prev };
    delete updated.showdown;      // Hide winner overlay
    delete updated.handComplete;  // Hide hand complete message
    delete updated.street;        // Clear street label
    updated.boardCards = [];      // Clear board cards immediately
    return updated;
  });
}
```

This ensures board card placeholders are empty immediately, preventing old flop/turn/river cards from displaying during the new hand setup.

**Why This Works:**

The boardCards array is used to track both:
1. The display of board card placeholders (always visible as empty slots)
2. The actual card values (displayed when available)

By setting `boardCards = []` in the optimistic update:
- Placeholders become empty immediately ✓
- No stale cards from previous hand display ✓
- Server confirmation (Phase 3) keeps the array empty ✓

**Review Status:** APPROVED ✅

**Git Commit Message:**
```
feat: Clear board cards on start_hand optimistic update

- Add boardCards clearing to start_hand action optimistic update
- Write 2 tests verifying board cards clear and array semantics preserved
- Prevents stale board cards from previous hand displaying during new hand setup
```
