## Phase 3 Complete: WebSocket Board Card Broadcasting

Implemented WebSocket broadcasting of community cards when flop, turn, and river are dealt, and integrated street advancement into the game action flow.

**Files created/changed:**
- internal/server/handlers.go
- internal/server/table.go
- internal/server/handlers_test.go
- internal/server/websocket_integration_test.go

**Functions created/changed:**
- `broadcastBoardDealt()` - Broadcasts board cards and street indicator to all players at table
- `Table.AdvanceToNextStreetWithBroadcast()` - Wrapper method combining street advancement with broadcasting
- `HandlePlayerAction()` - Modified to advance streets after betting rounds complete (2 locations)
- `BoardDealtPayload` struct - Message format for board_dealt events

**Tests created/changed:**
- `TestBroadcastBoardDealt_SendsToAllTablePlayers`
- `TestBroadcastBoardDealt_IncludesCorrectBoardCards`
- `TestBroadcastBoardDealt_IncludesStreetIndicator`
- `TestWebSocketFlow_FlopBroadcast_AfterPreflopComplete`
- `TestWebSocketFlow_TurnBroadcast_AfterFlopComplete`
- `TestWebSocketFlow_RiverBroadcast_AfterTurnComplete`
- `TestHandlePlayerAction_AdvancesStreetAfterRoundComplete` - Critical integration test
- `TestHandlePlayerAction_ValidCheck` - Updated for new street advancement behavior

**Review Status:** APPROVED

**Git Commit Message:**
```
feat: add board card broadcasting and street advancement integration

- Implement broadcastBoardDealt() to broadcast community cards via WebSocket
- Add AdvanceToNextStreetWithBroadcast() wrapper method for table street transitions
- Integrate automatic street advancement into HandlePlayerAction after betting rounds complete
- Handle both normal round completion and all-but-one-fold scenarios
- Add comprehensive tests for broadcast payloads and integration flow
- Fix critical TODO: game now properly advances preflop→flop→turn→river
```
