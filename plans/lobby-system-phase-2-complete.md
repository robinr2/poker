## Phase 2 Complete: Lobby State and WebSocket Events

Successfully implemented lobby state messaging over WebSocket. Clients now receive the current state of all 4 tables (with seat counts) when they connect or restore their session. Infrastructure is ready for real-time updates when seat changes occur.

**Files created/changed:**
- internal/server/handlers.go
- internal/server/websocket.go
- internal/server/handlers_test.go
- internal/server/websocket_integration_test.go
- frontend/src/App.tsx
- frontend/src/App.test.tsx
- frontend/src/hooks/useWebSocket.ts

**Backend Functions created/changed:**
- `Server.GetLobbyState() []TableInfo` - Thread-safe method (RLock) that builds array of table info from all 4 tables
- `Server.broadcastLobbyState() error` - Helper to broadcast lobby state to all connected clients via hub
- `Client.SendLobbyState(server, logger) error` - Sends lobby_state message to individual client
- `Client.HandleSetName()` - Modified to send lobby_state after session_created
- `HandleWebSocket()` - Modified to send lobby_state after session_restored
- `Client.readPump()` - Updated signature to accept Server parameter

**Frontend Functions created/changed:**
- `useWebSocket()` - Added optional `UseWebSocketOptions` parameter with `onMessage` callback
- `useWebSocket()` - Modified to call callback immediately on message receipt (before state update)
- `App.handleMessage()` - New callback that processes WebSocket messages with useCallback and refs
- `App.tsx` - Added `playerNameRef` and `showPromptRef` to maintain fresh values in callback

**Backend Structs created:**
- `TableInfo` - ID, Name, SeatsOccupied, MaxSeats (JSON: camelCase tags)

**Frontend Interfaces created:**
- `UseWebSocketOptions` - Interface with optional `onMessage?: (message: string) => void` callback

**Message Protocol Extended:**
- Added `lobby_state` message type: `{"type": "lobby_state", "payload": [...]}`
- Sent automatically after `session_created` (new users)
- Sent automatically after `session_restored` (returning users)

**Tests created:**
- `TestGetLobbyState` - Verifies empty tables return correct data (0 occupied seats)
- `TestGetLobbyStateWithOccupiedSeats` - Verifies seat counts with 2 and 4 occupied seats
- `TestGetLobbyStateThreadSafety` - Concurrent reads/writes (10 goroutines, 100 ops each)
- `TestWebSocketSendsLobbyStateOnConnect` - Client receives lobby_state after new connection
- `TestLobbyStateMessageFormat` - Validates JSON structure and all required fields
- `TestWebSocketSendsLobbyStateOnRestore` - Sends lobby_state after session restore

**Frontend Changes:**
- Add handling for `lobby_state` message in App.tsx message handler
- Fixed React 18 state batching issue where rapid `session_created` + `lobby_state` messages prevented modal close
- Solution: Added optional `onMessage` callback to useWebSocket hook that processes messages synchronously before state batching
- Updated App.tsx to use `handleMessage` callback with refs to avoid stale closures
- Fix test expectation for WebSocket connection count (now correctly creates 1 connection with token)
- All 63 frontend tests passing

**Technical Note - React 18 State Batching Fix:**
The server sends two messages rapidly: `session_created` → `lobby_state`. React 18's automatic batching caused only the last message to trigger the useEffect, so the modal never closed. We fixed this by adding an optional callback parameter to useWebSocket that executes synchronously on each message arrival, bypassing React's batching for critical message handling while still maintaining state updates for other consumers.

**Review Status:** APPROVED

**Git Commit Message:**
```
feat: Add lobby state WebSocket messaging

Backend:
- Create TableInfo struct with ID, Name, SeatsOccupied, MaxSeats fields
- Implement Server.GetLobbyState() with thread-safe RLock access to build table snapshots
- Implement Server.broadcastLobbyState() for pushing updates to all connected clients
- Implement Client.SendLobbyState() for individual client lobby state messages
- Send lobby_state automatically after session_created (new users)
- Send lobby_state automatically after session_restored (returning users)
- Add 6 comprehensive tests: empty state, occupied seats, thread safety, WebSocket integration, message format

Frontend:
- Add onMessage callback parameter to useWebSocket hook to handle rapid messages before React batching
- Implement App.handleMessage callback with refs to process session_created and lobby_state synchronously
- Fix React 18 batching issue where rapid messages prevented modal from closing after name submission
- Add handling for lobby_state message type in App component
- Fix WebSocket connection test expectation (1 connection with token)
- All 63 tests passing
```
