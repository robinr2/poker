## Phase 3 Complete: Remove Auto-Clear and Show Start Hand Button After Showdown (Frontend)

Removed 5-second auto-clear timeout for showdown overlay and added persistent "Start Hand" button to allow players to manually start the next hand. Showdown results now remain visible until player explicitly clicks "Start Hand".

**Files created/changed:**
- frontend/src/hooks/useWebSocket.ts
- frontend/src/components/TableView.tsx
- frontend/src/hooks/useWebSocket.test.ts
- frontend/src/components/TableView.test.tsx

**Functions created/changed:**
- `useWebSocket()` - Removed handCompleteTimeoutRef and all timeout logic, added state clearing when start_hand action sent
- `TableView` component - Updated button visibility to show during showdown, uses sendAction prop
- `handleStartHand()` - Now clears showdown state when clicked during showdown

**Tests created/changed:**
- `should not auto-clear showdown state after timeout` - Verifies showdown persists indefinitely
- `should clear showdown and handComplete when start_hand action is sent` - Verifies state cleanup on new hand
- `should display "Start Hand" button when showdown exists` - Verifies button appears and works during showdown

**Review Status:** APPROVED

**Git Commit Message:**
```
feat: persist showdown overlay until manual hand start

- Remove 5-second auto-clear timeout from showdown display
- Show "Start Hand" button during showdown for manual progression
- Clear showdown state locally when starting new hand
- Improves UX by letting players view results as long as needed
- Add 3 tests verifying showdown persistence and button behavior
```
