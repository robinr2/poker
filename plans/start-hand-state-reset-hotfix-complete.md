## Start Hand State Reset Hotfix - COMPLETE ✅

This hotfix successfully fixed two bugs in the start hand state management: (1) street label persisting after starting a new hand, and (2) Start Hand button remaining visible during active hand play. Both bugs were caused by incomplete state clearing in the optimistic update.

**Summary of All Phases**

### Phase 1: Clear Street Label on Hand Start ✅
- Added `delete updated.street` to start_hand optimistic update
- Added `street: undefined` to hand_started backend confirmation handler
- Verified street remains undefined until board_dealt is received
- All 222 frontend tests passing

### Phase 2: Fix Start Hand Button Visibility ✅
- Removed `updated.pot = 0` from start_hand optimistic update
- Button now correctly hides during active hand (pot > 0, no handComplete)
- Button shows on first hand (pot === 0) or after winner determined (handComplete set)
- Added 3 new tests for button visibility, updated 2 existing tests
- Board cards still hide correctly (handled by `boardCards = []`)
- All 225 frontend tests passing

**Files Modified (Total: 3)**
- `frontend/src/hooks/useWebSocket.ts` - Optimistic update state clearing
- `frontend/src/components/TableView.test.tsx` - Button visibility tests
- `frontend/src/hooks/useWebSocket.test.ts` - Updated pot value expectations

**Tests Added/Updated (Total: 8 new/updated)**
- Phase 1: 3 tests for street clearing
- Phase 2: 3 tests for button visibility + 2 updated tests for pot behavior

**Root Cause Analysis**

The start_hand optimistic update was over-clearing state:
- ✅ Correctly: Deleted `showdown` state (winner overlay)
- ✅ Correctly: Deleted `handComplete` state (hand complete message)
- ✅ Correctly: Deleted `street` field (street label)
- ✅ Correctly: Reset `boardCards` to [] (hide board)
- ❌ Incorrectly: Reset `pot` to 0 (made button think first hand)

This caused the button visibility logic to fail:
```javascript
const showStartHandButton = pot === 0 || handComplete;
```

During hand with pot > 0, the incorrectly-reset pot would make the condition `pot === 0` true, showing the button when it should be hidden.

**Impact**

- ✅ Street label no longer persists after starting new hand
- ✅ Start Hand button no longer visible during active hand play
- ✅ Board cards still hide immediately when starting hand
- ✅ All existing functionality preserved
- ✅ Zero regressions in test suite

**Review Status:** APPROVED ✅

**Total Test Coverage:** 225 frontend tests, all passing

**Next Steps:** Ready for integration or next feature work
