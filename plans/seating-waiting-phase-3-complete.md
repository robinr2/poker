## Phase 3 Complete: Leave Table and Disconnect Handling

Phase 3 successfully implements voluntary table leaving via leave_table messages and automatic seat cleanup on disconnect. Players can now leave tables to return to the lobby, and disconnecting players are automatically removed from seats with lobby state updates broadcast to remaining players.

**Files created/changed:**
- internal/server/handlers.go
- internal/server/server.go
- internal/server/websocket.go
- internal/server/websocket_integration_test.go

**Functions created/changed:**
- Client.HandleLeaveTable() - Processes leave_table message, clears seat, updates session
- Client.SendSeatCleared() - Sends seat_cleared confirmation to client
- Client.broadcastLobbyStateExcluding(*Client) - Broadcasts lobby state to all except one client
- Server.HandleDisconnect(token string) - Automatically clears seat on disconnect
- readPump message router - Added "leave_table" case
- readPump defer block - Calls HandleDisconnect before unregister

**Tests created/changed:**
- TestHandleLeaveTableSuccess - Player leaves table successfully
- TestHandleLeaveTableNotSeated - Error when player not seated
- TestLeaveTableBroadcastsLobbyState - Other clients receive lobby updates
- TestHandleDisconnectClearsSeat - Disconnect clears seat if seated
- TestHandleDisconnectNoSeat - Disconnect handles unseated player gracefully
- TestDisconnectBroadcastsLobbyState - Remaining clients get lobby updates

**Review Status:** APPROVED

**Git Commit Message:**
```
feat: implement leave table and disconnect handling

- Add HandleLeaveTable() for voluntary table leaving
- Add HandleDisconnect() for automatic seat cleanup on disconnect
- Add SendSeatCleared() confirmation message
- Add broadcastLobbyStateExcluding() to prevent duplicate messages
- Wire leave_table message in readPump router
- Call HandleDisconnect in readPump defer for disconnect cleanup
- 6 new integration tests, all passing
- Thread-safe with proper locking
```
