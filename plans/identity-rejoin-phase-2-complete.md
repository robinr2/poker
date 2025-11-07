## Phase 2 Complete: WebSocket Session Protocol

Extended WebSocket protocol to handle token-based authentication and session creation via JSON messages. Clients can now connect with existing tokens to restore sessions, or connect without a token and send set_name to create a new session. Comprehensive testing validates all protocol flows including edge cases.

**Files created/changed:**
- internal/server/handlers.go (NEW)
- internal/server/websocket.go (MODIFIED)
- internal/server/websocket_integration_test.go (NEW)

**Functions created/changed:**
- WebSocketMessage struct - Generic message envelope with type and payload
- SetNamePayload, SessionCreatedPayload, SessionRestoredPayload, ErrorPayload structs
- HandleWebSocket() - Enhanced to extract and validate token from query parameter
- readPump() - Updated to parse JSON messages and route by type
- Client.HandleSetName() - Creates new session from client name
- Client.SendSessionRestored() - Sends restored session info to rejoining player
- Client.SendError() - Sends error messages to client

**Tests created/changed:**
- TestHandleWebSocket_WithoutToken
- TestHandleWebSocket_WithValidToken
- TestHandleWebSocket_WithInvalidToken
- TestSetNameMessage
- TestSessionCreatedMessage
- TestSessionRestoredMessage
- TestSessionRestoredMessageWithoutTableSeat
- TestMultipleConnectionsSameToken
- TestInvalidSetNameMessage
- TestInvalidJSONMessage

**Review Status:** APPROVED ✅

**Test Coverage:** 91.4% (28/28 tests passing, no race conditions)

**Git Commit Message:**
```
feat: Add WebSocket session protocol with JSON messages

- Define JSON message types: WebSocketMessage envelope with type/payload structure
- Add message handlers: SetNamePayload, SessionCreatedPayload, SessionRestoredPayload, ErrorPayload
- Extract token from WebSocket query parameter (?token=uuid)
- Validate token on connection and send session_restored for valid tokens
- Implement set_name message handler to create new sessions
- Send session_created response with generated UUID token
- Parse incoming WebSocket messages as JSON and route by type
- Add graceful error handling for invalid JSON and malformed messages
- Add 10 comprehensive integration tests covering all protocol flows
- Test edge cases: invalid tokens, concurrent connections, empty names
```
