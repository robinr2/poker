## Phase 4 Complete: Frontend TableView UI

Phase 4 successfully implements the complete frontend UI for the seating system. Players can now see tables in the lobby, join tables to occupy seats, view their seat at the table, and leave to return to the lobby. The TableView component displays all 6 seats in a clean grid layout with the player's own seat highlighted. All message handling is properly wired for join_table, leave_table, seat_assigned, and seat_cleared WebSocket messages.

**Files created/changed:**
- frontend/src/components/TableView.tsx (NEW)
- frontend/src/components/TableView.test.tsx (NEW)
- frontend/src/styles/TableView.css (NEW)
- frontend/src/App.tsx
- frontend/src/App.test.tsx
- frontend/src/hooks/useWebSocket.ts
- frontend/src/components/LobbyView.tsx
- frontend/src/components/LobbyView.test.tsx

**Functions created/changed:**
- TableView component (NEW) - Displays 6 seats with player names and highlighting
- App view state management - Switches between lobby and table views
- App handleJoinTable() - Sends join_table WebSocket message
- App handleLeaveTable() - Sends leave_table WebSocket message
- useWebSocket seat message parsing - Handles seat_assigned and seat_cleared
- LobbyView join button - Wired to trigger join_table message

**Tests created/changed:**
- 16 new TableView tests (rendering, seat display, highlighting, layout, leave button)
- 5 new App tests (seat_assigned handling, seat_cleared handling, join/leave flow)
- 1 new LobbyView test (join button sends correct message)
- Fixed critical field name mismatch (tableID → tableId) in all tests

**Review Status:** APPROVED

**Git Commit Message:**
```
feat: implement frontend table view and seat display

- Add TableView component with 6-seat grid layout
- Add view state management in App (lobby/table switching)
- Wire join_table and leave_table WebSocket messages
- Handle seat_assigned message to show table view
- Handle seat_cleared message to return to lobby
- Highlight player's own seat with distinct styling
- Add Leave Table button for voluntary exit
- Fix field name mismatch (tableID → tableId)
- 22 new frontend tests, all passing
- Complete end-to-end join/leave flow working
```
