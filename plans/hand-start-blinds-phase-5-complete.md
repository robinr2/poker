## Phase 5 Complete: WebSocket Protocol Extension & Broadcast Integration

Successfully implemented WebSocket broadcast functions with card privacy and integrated them into StartHand() to automatically notify all players when a hand begins, blinds are posted, and cards are dealt.

**Files created/changed:**
- internal/server/table.go
- internal/server/handlers.go
- internal/server/websocket.go
- internal/server/server.go
- internal/server/table_test.go
- internal/server/websocket_integration_test.go

**Functions created/changed:**
- `Table.Server` (field added) - reference to Server for broadcast capability
- `NewTable()` - updated signature to accept Server parameter
- `StartHand()` - integrated broadcasts: hand_started, blind_posted (x2), cards_dealt
- `broadcastHandStarted()` - broadcasts dealer and blind positions to all table clients
- `broadcastBlindPosted()` - broadcasts blind posting events with updated stacks
- `broadcastCardsDealt()` - broadcasts hole cards with per-player privacy filtering
- `filterHoleCardsForPlayer()` - helper function ensuring players only see their own cards
- `HandStartedPayload`, `BlindPostedPayload`, `CardsDealtPayload` (types added)

**Tests created/changed:**
- `TestStartHandBroadcastsMessages` - verifies StartHand() broadcasts all four events in correct order
- `TestStartHandBroadcastsCardPrivacy` - verifies each player only receives their own hole cards
- Updated all table_test.go test cases to pass Server reference

**Review Status:** APPROVED

**Git Commit Message:**
feat: Integrate WebSocket broadcasts into hand start flow

- Add Server reference to Table struct for broadcast capability
- Integrate broadcasts into StartHand(): hand_started, blind_posted, cards_dealt
- Implement broadcastHandStarted() to notify clients of dealer and blind positions
- Implement broadcastBlindPosted() to notify clients of blind posting events
- Implement broadcastCardsDealt() with per-player card privacy filtering
- Add filterHoleCardsForPlayer() to ensure players only see their own hole cards
- Release lock before broadcasting to prevent blocking network operations
- Add error handling with state reversion on broadcast failures
- Add 2 integration tests verifying broadcast flow and card privacy
