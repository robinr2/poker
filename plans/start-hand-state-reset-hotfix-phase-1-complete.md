## Phase 1 Complete: Clear Street Label on Hand Start

Successfully implemented clearing of the street field when starting a new hand, preventing "River" or other street labels from persisting into the next hand.

**Files created/changed:**
- `frontend/src/hooks/useWebSocket.ts`
- `frontend/src/hooks/useWebSocket.test.ts`

**Functions created/changed:**
- `useWebSocket` - Modified `sendAction` function (line 613) to add `delete updated.street;` in optimistic update
- `useWebSocket` - Modified `hand_started` handler (line 307) to add `street: undefined,` in backend confirmation

**Tests created/changed:**
- `sendAction clears street when start_hand is sent` - Verifies optimistic clearing
- `hand_started message clears street field` - Verifies backend confirmation clearing
- `street remains undefined until board_dealt is received` - Verifies lifecycle integrity

**Review Status:** APPROVED ✅

**Git Commit Message:**
```
fix: Clear street label when starting new hand

- Add street field clearing to start_hand optimistic update
- Add street field clearing to hand_started backend confirmation
- Write 3 tests verifying street clears and remains undefined until board_dealt
- All 222 frontend tests passing
```
