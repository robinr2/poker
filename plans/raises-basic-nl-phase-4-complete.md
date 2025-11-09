## Phase 4 Complete: Handler Protocol Updates

Successfully updated WebSocket message handlers to support raise amounts in payloads. The implementation extends PlayerActionPayload to include optional Amount field, adds MinRaise and MaxRaise to ActionRequestPayload, and updates BroadcastActionRequest to calculate and include raise bounds.

**Files created/changed:**
- internal/server/handlers.go
- internal/server/handlers_test.go
- internal/server/server.go

**Functions created/changed:**
- `PlayerActionPayload` struct (added Amount field)
- `ActionRequestPayload` struct (added MinRaise and MaxRaise fields)
- `Server.BroadcastActionRequest()` (updated to calculate and include min/max raise)
- `Server.HandlePlayerAction()` (updated signature to accept variadic amount parameter)
- `Client.HandlePlayerActionMessage()` (updated to extract and pass amount from payload)

**Tests created/changed:**
- `TestHandlePlayerAction_RaiseWithAmount` (new)
- `TestBroadcastActionRequest_IncludesRaiseAmounts` (new)
- `TestHandlePlayerActionMessage_ExtractsRaiseAmount` (new)

**Review Status:** APPROVED

**Git Commit Message:**
```
feat: Update WebSocket handlers to support raise amounts in payloads

- Add Amount field to PlayerActionPayload for raise actions
- Add MinRaise and MaxRaise fields to ActionRequestPayload
- Update BroadcastActionRequest to calculate raise bounds before sending
- Extend HandlePlayerAction to accept optional amount parameter via variadic args
- Update HandlePlayerActionMessage to extract amount from payload
- Add 3 comprehensive tests covering raise amount handling
```
