## Plan: Identity & Rejoin System

Implement player identity and session management with UUID tokens, name registration, and automatic rejoin capability for poker game. This is the foundational feature enabling all multiplayer gameplay.

**Phases: 3 phases**

1. **Phase 1: Backend Session Management**
    - **Objective:** Create session storage, UUID token generation, and session lifecycle management
    - **Files/Functions to Create:**
      - `internal/server/session.go` - Session struct, SessionManager with Create/Get/Update/Remove methods
      - `internal/server/session_test.go` - Tests for session creation, retrieval, validation, and cleanup
    - **Files/Functions to Modify:**
      - `internal/server/server.go` - Add SessionManager to Server struct
      - `internal/server/websocket.go` - Add Token field to Client struct; modify Hub to track sessions
      - `go.mod` - Add `github.com/google/uuid` dependency
    - **Tests to Write:**
      - `TestSessionManager_CreateSession` - Test session creation with name validation
      - `TestSessionManager_GetSession` - Test session retrieval by token
      - `TestSessionManager_UpdateSession` - Test session updates (table, seat)
      - `TestSessionManager_RemoveSession` - Test session removal
      - `TestSessionManager_ConcurrentAccess` - Test thread safety
      - `TestSessionManager_InvalidNames` - Test name validation (empty, too long, invalid chars)
    - **Steps:**
        1. Write tests for SessionManager operations (should fail)
        2. Add `github.com/google/uuid` dependency to go.mod
        3. Create Session struct with Token, Name, TableID, SeatIndex, CreatedAt fields
        4. Implement SessionManager with thread-safe map and mutex
        5. Implement CreateSession with UUID generation and name validation (1-20 chars, alphanumeric + space/dash/underscore)
        6. Implement GetSession, UpdateSession, RemoveSession methods
        7. Add SessionManager to Server struct in server.go
        8. Add Token string field to Client struct in websocket.go
        9. Run tests to verify they pass

2. **Phase 2: WebSocket Session Protocol**
    - **Objective:** Extend WebSocket protocol to handle token-based authentication and session creation
    - **Files/Functions to Modify:**
      - `internal/server/websocket.go` - Modify HandleWebSocket to extract token; add message handling for set_name
      - `internal/server/handlers.go` - Add message type definitions and handlers
      - `internal/server/server_test.go` - Add integration tests for session flow
    - **Tests to Write:**
      - `TestHandleWebSocket_WithValidToken` - Test connection with existing valid token
      - `TestHandleWebSocket_WithInvalidToken` - Test connection with invalid/expired token
      - `TestHandleWebSocket_WithoutToken` - Test new connection without token
      - `TestSetNameMessage` - Test set_name message handling
      - `TestSessionCreatedMessage` - Test session_created response format
      - `TestSessionRestoredMessage` - Test session_restored for rejoin with table/seat
      - `TestMultipleConnectionsSameToken` - Test concurrent connections with same token
    - **Steps:**
        1. Write tests for WebSocket session protocol (should fail)
        2. Define JSON message types: MessageType string + Payload interface{}
        3. Modify HandleWebSocket to extract token from query parameter (?token=uuid)
        4. If token exists, validate with SessionManager; if valid, attach to Client
        5. If no token or invalid, create pending session state
        6. Implement message handler for "set_name" type
        7. On set_name: create session, send "session_created" message with token
        8. On valid token connection: send "session_restored" message with name and optional table/seat
        9. Update readPump to parse JSON messages and route by type
        10. Run tests to verify protocol works correctly

3. **Phase 3: Frontend Identity & Rejoin UI**
    - **Objective:** Implement name prompt, token storage, and automatic rejoin on page load
    - **Files/Functions to Create:**
      - `frontend/src/services/SessionService.ts` - localStorage management for token
      - `frontend/src/services/SessionService.test.ts` - Tests for SessionService
      - `frontend/src/components/NamePrompt.tsx` - Name entry modal component
      - `frontend/src/components/NamePrompt.test.tsx` - Tests for NamePrompt component
    - **Files/Functions to Modify:**
      - `frontend/src/services/WebSocketService.ts` - Accept token parameter, append to URL, handle session messages
      - `frontend/src/services/WebSocketService.test.ts` - Add tests for token handling
      - `frontend/src/App.tsx` - Add name prompt logic, auto-rejoin, session state management
      - `frontend/src/App.css` - Add styles for name prompt modal
    - **Tests to Write:**
      - `TestSessionService_GetToken` - Test token retrieval from localStorage
      - `TestSessionService_SetToken` - Test token storage
      - `TestSessionService_ClearToken` - Test token removal
      - `TestNamePrompt_Render` - Test modal renders with input field
      - `TestNamePrompt_Validation` - Test name validation (length, characters)
      - `TestNamePrompt_Submit` - Test form submission sends set_name message
      - `TestWebSocketService_TokenInURL` - Test token appended to WebSocket URL
      - `TestWebSocketService_SessionMessages` - Test handling of session_created/session_restored
      - `TestApp_NoToken_ShowsPrompt` - Test name prompt shown when no token in localStorage
      - `TestApp_ValidToken_AutoConnects` - Test auto-connection with existing token
      - `TestApp_SessionRestored_WithTable` - Test navigation to table if seated
    - **Steps:**
        1. Write tests for SessionService and NamePrompt (should fail)
        2. Create SessionService with getToken, setToken, clearToken methods using localStorage key "poker_session_token"
        3. Create NamePrompt component with form, validation, and onSubmit callback
        4. Modify WebSocketService constructor to accept optional token parameter
        5. Append token to WebSocket URL if provided: `ws://host/ws?token=${token}`
        6. Add message handlers for "session_created" and "session_restored" types
        7. Update App.tsx to check localStorage for token on mount
        8. If no token: show NamePrompt modal
        9. On name submit: send "set_name" message via WebSocket
        10. On "session_created": save token to localStorage, hide prompt
        11. On "session_restored" with tableID: navigate to table view (placeholder for now)
        12. Add CSS for modal overlay and form styling
        13. Run tests to verify frontend flow works

**Open Questions:**
1. Should we allow anonymous connections initially, or require name before any connection? **Recommendation: Require name before game actions, but allow connection first**
2. Should we add session expiry/TTL for MVP? **Recommendation: No expiry for MVP; in-memory clears on restart**
3. HTTP endpoint for session creation vs WebSocket-only? **Recommendation: WebSocket-only for simplicity**
