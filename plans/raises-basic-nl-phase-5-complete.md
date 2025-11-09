## Phase 5 Complete: Frontend Protocol and State

Successfully implemented frontend protocol parsing and state management for raise actions. The implementation extends the useWebSocket hook to parse minRaise and maxRaise from action_request messages, store raise amounts in GameState, and provide a sendAction helper function for sending raise actions with amounts to the backend.

**Files created/changed:**
- frontend/src/hooks/useWebSocket.ts
- frontend/src/hooks/useWebSocket.test.ts

**Functions created/changed:**
- `GameState` interface (added minRaise and maxRaise fields)
- `ActionRequestPayload` interface (added minRaise and maxRaise fields)
- `UseWebSocketReturn` interface (added sendAction function)
- `useWebSocket()` hook (added playerSeatIndex state tracking)
- Message handler for 'action_request' (updated to parse and store raise bounds)
- `sendAction()` memoized callback (new - sends player_action with optional amount)

**Tests created/changed:**
- `TestUseWebSocket_ParseMinMaxRaiseFromActionRequest` (new)
- `TestUseWebSocket_SendActionWithRaiseAmount` (new)
- `TestUseWebSocket_TrackPlayerSeatIndex` (new)
- `TestUseWebSocket_SendActionWithoutAmount` (updated - now tests both raise and non-raise)

**Review Status:** APPROVED

**Git Commit Message:**
```
feat: Add raise amount protocol to frontend WebSocket handling

- Extend ActionRequestPayload interface with minRaise and maxRaise fields
- Add playerSeatIndex state tracking in useWebSocket hook
- Update action_request handler to parse and store raise bounds in GameState
- Implement sendAction() helper for sending player_action with optional amount
- Update seat_assigned/seat_cleared handlers to track player's seat index
- Add 4 comprehensive tests covering protocol parsing and action sending
```
