package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// readMessage reads a JSON message from the WebSocket connection
func readMessage(t *testing.T, ws *websocket.Conn) WebSocketMessage {
	t.Helper()
	var msg WebSocketMessage
	err := ws.ReadJSON(&msg)
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	return msg
}

// sendMessage sends a JSON message to the WebSocket connection
func sendMessage(t *testing.T, ws *websocket.Conn, msgType string, payload interface{}) {
	t.Helper()
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	msg := WebSocketMessage{
		Type:    msgType,
		Payload: json.RawMessage(payloadBytes),
	}
	err = ws.WriteJSON(msg)
	if err != nil {
		t.Fatalf("failed to send message: %v", err)
	}
}

// TestHandleWebSocket_WithoutToken tests new connection without token
func TestHandleWebSocket_WithoutToken(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")

	// Connect without token
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Should be able to connect without token
	if ws == nil {
		t.Fatal("expected websocket connection to be established")
	}

	// Send set_name message
	sendMessage(t, ws, "set_name", SetNamePayload{Name: "Alice"})

	// Should receive session_created message
	msg := readMessage(t, ws)
	if msg.Type != "session_created" {
		t.Errorf("expected message type 'session_created', got %q", msg.Type)
	}

	// Parse payload
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload SessionCreatedPayload
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		t.Fatalf("failed to parse session_created payload: %v", err)
	}

	if payload.Name != "Alice" {
		t.Errorf("expected name 'Alice', got %q", payload.Name)
	}

	if payload.Token == "" {
		t.Error("expected token to be set, got empty string")
	}
}

// TestHandleWebSocket_WithValidToken tests connection with existing valid token
func TestHandleWebSocket_WithValidToken(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create a session first
	session, err := server.sessionManager.CreateSession("Bob")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session.Token

	// Connect with valid token
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Should receive session_restored message automatically
	msg := readMessage(t, ws)
	if msg.Type != "session_restored" {
		t.Errorf("expected message type 'session_restored', got %q", msg.Type)
	}

	// Parse payload
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload SessionRestoredPayload
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		t.Fatalf("failed to parse session_restored payload: %v", err)
	}

	if payload.Name != "Bob" {
		t.Errorf("expected name 'Bob', got %q", payload.Name)
	}
}

// TestHandleWebSocket_WithInvalidToken tests connection with invalid/expired token
func TestHandleWebSocket_WithInvalidToken(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Convert http:// to ws://
	invalidToken := "invalid-uuid-token"
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + invalidToken

	// Connect with invalid token
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Give the server a moment to send the error message
	time.Sleep(10 * time.Millisecond)

	// Should receive error message
	msg := readMessage(t, ws)
	if msg.Type != "error" {
		t.Errorf("expected message type 'error', got %q", msg.Type)
	}

	// Parse payload
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload ErrorPayload
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		t.Fatalf("failed to parse error payload: %v", err)
	}

	if payload.Message == "" {
		t.Error("expected error message to be set, got empty string")
	}
}

// TestSetNameMessage tests set_name message handling
func TestSetNameMessage(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")

	// Connect without token
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Send set_name message
	sendMessage(t, ws, "set_name", SetNamePayload{Name: "Charlie"})

	// Should receive session_created message
	msg := readMessage(t, ws)
	if msg.Type != "session_created" {
		t.Errorf("expected message type 'session_created', got %q", msg.Type)
	}

	// Verify the session was created in the SessionManager
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload SessionCreatedPayload
	json.Unmarshal(payloadBytes, &payload)

	retrievedSession, err := server.sessionManager.GetSession(payload.Token)
	if err != nil {
		t.Errorf("failed to retrieve session: %v", err)
	}

	if retrievedSession.Name != "Charlie" {
		t.Errorf("expected name 'Charlie' in retrieved session, got %q", retrievedSession.Name)
	}
}

// TestSessionCreatedMessage tests session_created response format
func TestSessionCreatedMessage(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")

	// Connect without token
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Send set_name message
	sendMessage(t, ws, "set_name", SetNamePayload{Name: "David"})

	// Should receive session_created message with correct format
	msg := readMessage(t, ws)
	if msg.Type != "session_created" {
		t.Errorf("expected message type 'session_created', got %q", msg.Type)
	}

	// Verify payload has required fields
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload SessionCreatedPayload
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		t.Fatalf("failed to parse session_created payload: %v", err)
	}

	if payload.Token == "" {
		t.Error("expected token in session_created payload")
	}

	if payload.Name == "" {
		t.Error("expected name in session_created payload")
	}
}

// TestSessionRestoredMessage tests session_restored for rejoin with table/seat
func TestSessionRestoredMessage(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create a session and set table/seat info
	session, err := server.sessionManager.CreateSession("Eve")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	tableID := "table-1"
	seatIndex := 3
	_, err = server.sessionManager.UpdateSession(session.Token, &tableID, &seatIndex)
	if err != nil {
		t.Fatalf("failed to update session: %v", err)
	}

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session.Token

	// Connect with valid token
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Should receive session_restored message with table/seat info
	msg := readMessage(t, ws)
	if msg.Type != "session_restored" {
		t.Errorf("expected message type 'session_restored', got %q", msg.Type)
	}

	// Parse payload
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload SessionRestoredPayload
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		t.Fatalf("failed to parse session_restored payload: %v", err)
	}

	if payload.Name != "Eve" {
		t.Errorf("expected name 'Eve', got %q", payload.Name)
	}

	if payload.TableID == nil || *payload.TableID != "table-1" {
		t.Errorf("expected tableID 'table-1', got %v", payload.TableID)
	}

	if payload.SeatIndex == nil || *payload.SeatIndex != 3 {
		t.Errorf("expected seatIndex 3, got %v", payload.SeatIndex)
	}
}

// TestSessionRestoredMessageWithoutTableSeat tests session_restored without table/seat
func TestSessionRestoredMessageWithoutTableSeat(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create a session without table/seat info
	session, err := server.sessionManager.CreateSession("Frank")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session.Token

	// Connect with valid token
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Should receive session_restored message without table/seat info
	msg := readMessage(t, ws)
	if msg.Type != "session_restored" {
		t.Errorf("expected message type 'session_restored', got %q", msg.Type)
	}

	// Parse payload
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload SessionRestoredPayload
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		t.Fatalf("failed to parse session_restored payload: %v", err)
	}

	if payload.Name != "Frank" {
		t.Errorf("expected name 'Frank', got %q", payload.Name)
	}

	if payload.TableID != nil {
		t.Errorf("expected tableID to be nil, got %v", payload.TableID)
	}

	if payload.SeatIndex != nil {
		t.Errorf("expected seatIndex to be nil, got %v", payload.SeatIndex)
	}
}

// TestMultipleConnectionsSameToken tests concurrent connections with same token
func TestMultipleConnectionsSameToken(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create a session
	session, err := server.sessionManager.CreateSession("Grace")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session.Token

	// Connect first client with valid token
	dialer := websocket.Dialer{}
	ws1, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect first client: %v", err)
	}
	defer ws1.Close()

	// Receive session_restored message for first client
	msg1 := readMessage(t, ws1)
	if msg1.Type != "session_restored" {
		t.Errorf("expected first client to receive 'session_restored', got %q", msg1.Type)
	}

	// Connect second client with same token
	ws2, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect second client: %v", err)
	}
	defer ws2.Close()

	// Receive session_restored message for second client
	msg2 := readMessage(t, ws2)
	if msg2.Type != "session_restored" {
		t.Errorf("expected second client to receive 'session_restored', got %q", msg2.Type)
	}

	// Both should have the same name
	payloadBytes1, _ := json.Marshal(msg1.Payload)
	var payload1 SessionRestoredPayload
	json.Unmarshal(payloadBytes1, &payload1)

	payloadBytes2, _ := json.Marshal(msg2.Payload)
	var payload2 SessionRestoredPayload
	json.Unmarshal(payloadBytes2, &payload2)

	if payload1.Name != payload2.Name {
		t.Errorf("expected both clients to have same name, got %q and %q", payload1.Name, payload2.Name)
	}
}

// TestInvalidSetNameMessage tests set_name with invalid name
func TestInvalidSetNameMessage(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")

	// Connect without token
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Send set_name message with empty name
	sendMessage(t, ws, "set_name", SetNamePayload{Name: ""})

	// Should receive error message
	msg := readMessage(t, ws)
	if msg.Type != "error" {
		t.Errorf("expected message type 'error', got %q", msg.Type)
	}

	// Parse payload
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload ErrorPayload
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		t.Fatalf("failed to parse error payload: %v", err)
	}

	if payload.Message == "" {
		t.Error("expected error message to be set, got empty string")
	}
}

// TestInvalidJSONMessage tests handling of invalid JSON
func TestInvalidJSONMessage(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")

	// Connect without token
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Send invalid JSON message
	err = ws.WriteMessage(websocket.TextMessage, []byte("invalid json {"))
	if err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	// Should receive error message
	msg := readMessage(t, ws)
	if msg.Type != "error" {
		t.Errorf("expected message type 'error', got %q", msg.Type)
	}
}

// TestWebSocketSendsLobbyStateOnConnect verifies client receives lobby_state after connection
func TestWebSocketSendsLobbyStateOnConnect(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")

	// Connect without token
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Send set_name message
	sendMessage(t, ws, "set_name", SetNamePayload{Name: "TestPlayer"})

	// Receive session_created message
	msg1 := readMessage(t, ws)
	if msg1.Type != "session_created" {
		t.Errorf("expected first message to be 'session_created', got %q", msg1.Type)
	}

	// Receive lobby_state message
	msg2 := readMessage(t, ws)
	if msg2.Type != "lobby_state" {
		t.Errorf("expected second message to be 'lobby_state', got %q", msg2.Type)
	}

	// Payload is double-encoded: first unmarshal to string, then parse JSON
	var payloadStr string
	err = json.Unmarshal(msg2.Payload, &payloadStr)
	if err != nil {
		t.Fatalf("failed to parse lobby_state payload as string: %v", err)
	}

	var lobbyState []interface{}
	err = json.Unmarshal([]byte(payloadStr), &lobbyState)
	if err != nil {
		t.Fatalf("failed to parse lobby_state array: %v", err)
	}

	if len(lobbyState) != 4 {
		t.Errorf("expected 4 tables in lobby state, got %d", len(lobbyState))
	}
}

// TestLobbyStateMessageFormat verifies JSON structure of lobby_state message
func TestLobbyStateMessageFormat(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")

	// Connect without token
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Send set_name message
	sendMessage(t, ws, "set_name", SetNamePayload{Name: "TestPlayer"})

	// Read session_created
	_ = readMessage(t, ws)

	// Read lobby_state message
	msg := readMessage(t, ws)
	if msg.Type != "lobby_state" {
		t.Errorf("expected message type 'lobby_state', got %q", msg.Type)
	}

	// Payload is double-encoded: first unmarshal to string, then parse JSON
	var payloadStr string
	err = json.Unmarshal(msg.Payload, &payloadStr)
	if err != nil {
		t.Fatalf("failed to parse lobby_state payload as string: %v", err)
	}

	// Parse as array of table info
	var tables []map[string]interface{}
	err = json.Unmarshal([]byte(payloadStr), &tables)
	if err != nil {
		t.Fatalf("failed to parse lobby_state tables array: %v", err)
	}

	if len(tables) != 4 {
		t.Errorf("expected 4 tables, got %d", len(tables))
	}

	// Verify first table has required fields
	firstTable := tables[0]
	if id, ok := firstTable["id"].(string); !ok || id == "" {
		t.Error("expected table to have 'id' field")
	}

	if name, ok := firstTable["name"].(string); !ok || name == "" {
		t.Error("expected table to have 'name' field")
	}

	if maxSeats, ok := firstTable["max_seats"].(float64); !ok || maxSeats != 6 {
		t.Error("expected table to have 'max_seats' field with value 6")
	}

	if seatsOccupied, ok := firstTable["seats_occupied"].(float64); !ok {
		t.Error("expected table to have 'seats_occupied' field")
	} else if seatsOccupied < 0 || seatsOccupied > 6 {
		t.Errorf("expected seats_occupied to be between 0 and 6, got %v", seatsOccupied)
	}
}

// TestWebSocketSendsLobbyStateOnRestore verifies lobby_state sent after session_restored
func TestWebSocketSendsLobbyStateOnRestore(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create a session
	session, err := server.sessionManager.CreateSession("Player")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session.Token

	// Connect with valid token
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Receive session_restored message
	msg1 := readMessage(t, ws)
	if msg1.Type != "session_restored" {
		t.Errorf("expected first message to be 'session_restored', got %q", msg1.Type)
	}

	// Receive lobby_state message
	msg2 := readMessage(t, ws)
	if msg2.Type != "lobby_state" {
		t.Errorf("expected second message to be 'lobby_state', got %q", msg2.Type)
	}

	// Payload is double-encoded: first unmarshal to string, then parse JSON
	var payloadStr string
	err = json.Unmarshal(msg2.Payload, &payloadStr)
	if err != nil {
		t.Fatalf("failed to parse lobby_state payload as string: %v", err)
	}

	var lobbyState []interface{}
	err = json.Unmarshal([]byte(payloadStr), &lobbyState)
	if err != nil {
		t.Fatalf("failed to parse lobby_state array: %v", err)
	}

	if len(lobbyState) != 4 {
		t.Errorf("expected 4 tables in lobby state, got %d", len(lobbyState))
	}
}

// TestHandleJoinTableSuccess tests successful join_table with seat assignment
func TestHandleJoinTableSuccess(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create a session
	session, err := server.sessionManager.CreateSession("Player")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session.Token

	// Connect with valid token
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Receive session_restored message
	_ = readMessage(t, ws)

	// Receive lobby_state message
	_ = readMessage(t, ws)

	// Send join_table message
	sendMessage(t, ws, "join_table", JoinTablePayload{TableId: "table-1"})

	// Receive seat_assigned message
	msg := readMessage(t, ws)
	if msg.Type != "seat_assigned" {
		t.Errorf("expected message type 'seat_assigned', got %q", msg.Type)
	}

	// Parse seat_assigned payload
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload SeatAssignedPayload
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		t.Fatalf("failed to parse seat_assigned payload: %v", err)
	}

	if payload.TableId != "table-1" {
		t.Errorf("expected tableId 'table-1', got %q", payload.TableId)
	}

	if payload.SeatIndex != 0 {
		t.Errorf("expected seatIndex 0, got %d", payload.SeatIndex)
	}

	if payload.Status != "waiting" {
		t.Errorf("expected status 'waiting', got %q", payload.Status)
	}

	// Verify session was updated with table and seat info
	updatedSession, err := server.sessionManager.GetSession(session.Token)
	if err != nil {
		t.Fatalf("failed to retrieve session: %v", err)
	}

	if updatedSession.TableID == nil || *updatedSession.TableID != "table-1" {
		t.Errorf("expected session TableID 'table-1', got %v", updatedSession.TableID)
	}

	if updatedSession.SeatIndex == nil || *updatedSession.SeatIndex != 0 {
		t.Errorf("expected session SeatIndex 0, got %v", updatedSession.SeatIndex)
	}
}

// TestHandleJoinTableFull tests join_table returns error when table is full
func TestHandleJoinTableFull(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Fill all 6 seats on table-1 with dummy tokens
	for i := 0; i < 6; i++ {
		token := "dummy-token-" + string(rune('0'+i))
		_, _ = server.tables[0].AssignSeat(&token)
	}

	// Create a session for our player
	session, err := server.sessionManager.CreateSession("Player")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session.Token

	// Connect with valid token
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Receive session_restored message
	_ = readMessage(t, ws)

	// Receive lobby_state message
	_ = readMessage(t, ws)

	// Send join_table message for full table
	sendMessage(t, ws, "join_table", JoinTablePayload{TableId: "table-1"})

	// Should receive error message
	msg := readMessage(t, ws)
	if msg.Type != "error" {
		t.Errorf("expected message type 'error', got %q", msg.Type)
	}

	// Parse error payload
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload ErrorPayload
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		t.Fatalf("failed to parse error payload: %v", err)
	}

	if payload.Message != "table_full" {
		t.Errorf("expected error message 'table_full', got %q", payload.Message)
	}
}

// TestHandleJoinTableAlreadySeated tests join_table returns error when player already seated
func TestHandleJoinTableAlreadySeated(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create a session
	session, err := server.sessionManager.CreateSession("Player")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Manually seat player at table-1
	_, _ = server.tables[0].AssignSeat(&session.Token)
	server.sessionManager.UpdateSession(session.Token, &server.tables[0].ID, nil)

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session.Token

	// Connect with valid token
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Receive session_restored message
	_ = readMessage(t, ws)

	// Receive lobby_state message
	_ = readMessage(t, ws)

	// Send join_table message for another table while already seated
	sendMessage(t, ws, "join_table", JoinTablePayload{TableId: "table-2"})

	// Should receive error message
	msg := readMessage(t, ws)
	if msg.Type != "error" {
		t.Errorf("expected message type 'error', got %q", msg.Type)
	}

	// Parse error payload
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload ErrorPayload
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		t.Fatalf("failed to parse error payload: %v", err)
	}

	if payload.Message != "already_seated" {
		t.Errorf("expected error message 'already_seated', got %q", payload.Message)
	}
}

// TestHandleJoinTableInvalidTableID tests join_table returns error for non-existent table
func TestHandleJoinTableInvalidTableID(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create a session
	session, err := server.sessionManager.CreateSession("Player")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session.Token

	// Connect with valid token
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Receive session_restored message
	_ = readMessage(t, ws)

	// Receive lobby_state message
	_ = readMessage(t, ws)

	// Send join_table message for non-existent table
	sendMessage(t, ws, "join_table", JoinTablePayload{TableId: "table-999"})

	// Should receive error message
	msg := readMessage(t, ws)
	if msg.Type != "error" {
		t.Errorf("expected message type 'error', got %q", msg.Type)
	}

	// Parse error payload
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload ErrorPayload
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		t.Fatalf("failed to parse error payload: %v", err)
	}

	if payload.Message != "invalid_table" {
		t.Errorf("expected error message 'invalid_table', got %q", payload.Message)
	}
}

// TestSeatAssignedMessage tests seat_assigned message format
func TestSeatAssignedMessage(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create a session
	session, err := server.sessionManager.CreateSession("Player")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session.Token

	// Connect with valid token
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Receive session_restored message
	_ = readMessage(t, ws)

	// Receive lobby_state message
	_ = readMessage(t, ws)

	// Send join_table message
	sendMessage(t, ws, "join_table", JoinTablePayload{TableId: "table-2"})

	// Receive seat_assigned message
	msg := readMessage(t, ws)
	if msg.Type != "seat_assigned" {
		t.Fatalf("expected message type 'seat_assigned', got %q", msg.Type)
	}

	// Verify payload has all required fields
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload SeatAssignedPayload
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		t.Fatalf("failed to parse seat_assigned payload: %v", err)
	}

	if payload.TableId == "" {
		t.Error("expected tableId in seat_assigned payload")
	}

	if payload.SeatIndex < 0 || payload.SeatIndex > 5 {
		t.Errorf("expected seatIndex between 0-5, got %d", payload.SeatIndex)
	}

	if payload.Status == "" {
		t.Error("expected status in seat_assigned payload")
	}
}

// TestJoinTableUpdatesSession tests that session is updated with TableID and SeatIndex
func TestJoinTableUpdatesSession(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create a session
	session, err := server.sessionManager.CreateSession("Player")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Verify session initially has no table/seat
	initialSession, _ := server.sessionManager.GetSession(session.Token)
	if initialSession.TableID != nil {
		t.Fatal("expected initial session to have nil TableID")
	}

	if initialSession.SeatIndex != nil {
		t.Fatal("expected initial session to have nil SeatIndex")
	}

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session.Token

	// Connect with valid token
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Receive session_restored message
	_ = readMessage(t, ws)

	// Receive lobby_state message
	_ = readMessage(t, ws)

	// Send join_table message
	sendMessage(t, ws, "join_table", JoinTablePayload{TableId: "table-3"})

	// Receive seat_assigned message
	_ = readMessage(t, ws)

	// Verify session was updated
	updatedSession, _ := server.sessionManager.GetSession(session.Token)
	if updatedSession.TableID == nil || *updatedSession.TableID != "table-3" {
		t.Errorf("expected session TableID to be 'table-3', got %v", updatedSession.TableID)
	}

	if updatedSession.SeatIndex == nil || *updatedSession.SeatIndex < 0 || *updatedSession.SeatIndex > 5 {
		t.Errorf("expected session SeatIndex to be 0-5, got %v", updatedSession.SeatIndex)
	}
}

// TestJoinTableBroadcastsLobbyState tests that lobby_state is broadcast after join
func TestJoinTableBroadcastsLobbyState(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create two sessions
	session1, err := server.sessionManager.CreateSession("Player1")
	if err != nil {
		t.Fatalf("failed to create session1: %v", err)
	}

	session2, err := server.sessionManager.CreateSession("Player2")
	if err != nil {
		t.Fatalf("failed to create session2: %v", err)
	}

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Connect first client
	wsURL1 := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session1.Token
	dialer := websocket.Dialer{}
	ws1, _, err := dialer.Dial(wsURL1, nil)
	if err != nil {
		t.Fatalf("failed to connect ws1: %v", err)
	}
	defer ws1.Close()

	// Receive session_restored and lobby_state
	_ = readMessage(t, ws1)
	_ = readMessage(t, ws1)

	// Connect second client
	wsURL2 := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session2.Token
	ws2, _, err := dialer.Dial(wsURL2, nil)
	if err != nil {
		t.Fatalf("failed to connect ws2: %v", err)
	}
	defer ws2.Close()

	// Receive session_restored and lobby_state for second client
	_ = readMessage(t, ws2)
	_ = readMessage(t, ws2)

	// Player 1 joins a table
	sendMessage(t, ws1, "join_table", JoinTablePayload{TableId: "table-1"})

	// Player 1 receives seat_assigned
	msg1 := readMessage(t, ws1)
	if msg1.Type != "seat_assigned" {
		t.Fatalf("expected seat_assigned for player1, got %q", msg1.Type)
	}

	// Player 2 should receive lobby_state broadcast
	msg2 := readMessage(t, ws2)
	if msg2.Type != "lobby_state" {
		t.Fatalf("expected lobby_state for player2, got %q", msg2.Type)
	}

	// Verify the lobby_state shows updated seat count
	var payloadStr string
	err = json.Unmarshal(msg2.Payload, &payloadStr)
	if err != nil {
		t.Fatalf("failed to parse lobby_state payload as string: %v", err)
	}

	var tables []map[string]interface{}
	err = json.Unmarshal([]byte(payloadStr), &tables)
	if err != nil {
		t.Fatalf("failed to parse lobby_state tables array: %v", err)
	}

	// Find table-1 in the lobby state
	var table1 map[string]interface{}
	for _, table := range tables {
		if id, ok := table["id"].(string); ok && id == "table-1" {
			table1 = table
			break
		}
	}

	if table1 == nil {
		t.Fatal("table-1 not found in lobby_state")
	}

	seatsOccupied, ok := table1["seats_occupied"].(float64)
	if !ok || int(seatsOccupied) != 1 {
		t.Errorf("expected seats_occupied to be 1 for table-1, got %v", seatsOccupied)
	}
}

// TestHandleLeaveTableSuccess tests successful leave_table with seat clearing
func TestHandleLeaveTableSuccess(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create a session
	session, err := server.sessionManager.CreateSession("Player")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session.Token

	// Connect with valid token
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Receive session_restored message
	_ = readMessage(t, ws)

	// Receive lobby_state message
	_ = readMessage(t, ws)

	// Send join_table message
	sendMessage(t, ws, "join_table", JoinTablePayload{TableId: "table-1"})

	// Receive seat_assigned message
	msg := readMessage(t, ws)
	if msg.Type != "seat_assigned" {
		t.Errorf("expected message type 'seat_assigned', got %q", msg.Type)
	}

	// Receive table_state message (sent after join)
	_ = readMessage(t, ws)

	// Verify session has table and seat
	updatedSession, _ := server.sessionManager.GetSession(session.Token)
	if updatedSession.TableID == nil || *updatedSession.TableID != "table-1" {
		t.Fatalf("session TableID should be set")
	}

	// Send leave_table message
	sendMessage(t, ws, "leave_table", struct{}{})

	// Receive seat_cleared message
	msg = readMessage(t, ws)
	if msg.Type != "seat_cleared" {
		t.Errorf("expected message type 'seat_cleared', got %q", msg.Type)
	}

	// Verify session was updated (TableID and SeatIndex should be nil)
	updatedSession, _ = server.sessionManager.GetSession(session.Token)
	if updatedSession.TableID != nil {
		t.Errorf("expected session TableID to be nil after leave_table, got %v", updatedSession.TableID)
	}
	if updatedSession.SeatIndex != nil {
		t.Errorf("expected session SeatIndex to be nil after leave_table, got %v", updatedSession.SeatIndex)
	}

	// Verify seat is actually cleared on the table
	table := server.tables[0] // table-1 is the first table
	seat, found := table.GetSeatByToken(&session.Token)
	if found {
		t.Errorf("expected player to be removed from table, but found in seat %d", seat.Index)
	}
}

// TestHandleLeaveTableNotSeated tests leave_table returns error when player not seated
func TestHandleLeaveTableNotSeated(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create a session
	session, err := server.sessionManager.CreateSession("Player")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session.Token

	// Connect with valid token
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Receive session_restored message
	_ = readMessage(t, ws)

	// Receive lobby_state message
	_ = readMessage(t, ws)

	// Send leave_table message without being seated
	sendMessage(t, ws, "leave_table", struct{}{})

	// Should receive error message
	msg := readMessage(t, ws)
	if msg.Type != "error" {
		t.Errorf("expected message type 'error', got %q", msg.Type)
	}

	// Parse error payload
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload ErrorPayload
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		t.Fatalf("failed to parse error payload: %v", err)
	}

	if payload.Message != "not_seated" {
		t.Errorf("expected error message 'not_seated', got %q", payload.Message)
	}
}

// TestLeaveTableBroadcastsLobbyState tests that lobby_state is broadcast after leave
func TestLeaveTableBroadcastsLobbyState(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create two sessions
	session1, err := server.sessionManager.CreateSession("Player1")
	if err != nil {
		t.Fatalf("failed to create session1: %v", err)
	}

	session2, err := server.sessionManager.CreateSession("Player2")
	if err != nil {
		t.Fatalf("failed to create session2: %v", err)
	}

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Connect first client
	wsURL1 := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session1.Token
	dialer := websocket.Dialer{}
	ws1, _, err := dialer.Dial(wsURL1, nil)
	if err != nil {
		t.Fatalf("failed to connect ws1: %v", err)
	}
	defer ws1.Close()

	// Receive session_restored and lobby_state
	_ = readMessage(t, ws1)
	_ = readMessage(t, ws1)

	// Connect second client
	wsURL2 := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session2.Token
	ws2, _, err := dialer.Dial(wsURL2, nil)
	if err != nil {
		t.Fatalf("failed to connect ws2: %v", err)
	}
	defer ws2.Close()

	// Receive session_restored and lobby_state for second client
	_ = readMessage(t, ws2)
	_ = readMessage(t, ws2)

	// Player 1 joins a table
	sendMessage(t, ws1, "join_table", JoinTablePayload{TableId: "table-1"})

	// Player 1 receives seat_assigned
	_ = readMessage(t, ws1)

	// Player 1 receives table_state
	_ = readMessage(t, ws1)

	// Player 2 receives lobby_state broadcast
	_ = readMessage(t, ws2)

	// Player 1 leaves the table
	sendMessage(t, ws1, "leave_table", struct{}{})

	// Player 1 receives seat_cleared
	msg1 := readMessage(t, ws1)
	if msg1.Type != "seat_cleared" {
		t.Fatalf("expected seat_cleared for player1, got %q", msg1.Type)
	}

	// Player 2 should receive lobby_state broadcast
	msg2 := readMessage(t, ws2)
	if msg2.Type != "lobby_state" {
		t.Fatalf("expected lobby_state for player2, got %q", msg2.Type)
	}

	// Verify the lobby_state shows updated seat count (back to 0)
	var payloadStr string
	err = json.Unmarshal(msg2.Payload, &payloadStr)
	if err != nil {
		t.Fatalf("failed to parse lobby_state payload as string: %v", err)
	}

	var tables []map[string]interface{}
	err = json.Unmarshal([]byte(payloadStr), &tables)
	if err != nil {
		t.Fatalf("failed to parse lobby_state tables array: %v", err)
	}

	// Find table-1 in the lobby state
	var table1 map[string]interface{}
	for _, table := range tables {
		if id, ok := table["id"].(string); ok && id == "table-1" {
			table1 = table
			break
		}
	}

	if table1 == nil {
		t.Fatal("table-1 not found in lobby_state")
	}

	seatsOccupied, ok := table1["seats_occupied"].(float64)
	if !ok || int(seatsOccupied) != 0 {
		t.Errorf("expected seats_occupied to be 0 for table-1 after leave, got %v", seatsOccupied)
	}
}

// TestHandleDisconnectClearsSeat tests that disconnect clears seat if player was seated
func TestHandleDisconnectClearsSeat(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create a session
	session, err := server.sessionManager.CreateSession("Player")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session.Token

	// Connect with valid token
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Receive session_restored message
	_ = readMessage(t, ws)

	// Receive lobby_state message
	_ = readMessage(t, ws)

	// Send join_table message
	sendMessage(t, ws, "join_table", JoinTablePayload{TableId: "table-1"})

	// Receive seat_assigned message
	_ = readMessage(t, ws)

	// Verify player is seated
	seat := server.FindPlayerSeat(&session.Token)
	if seat == nil {
		t.Fatal("expected player to be seated before disconnect")
	}

	// Close the connection
	ws.Close()

	// Give server time to process disconnect
	time.Sleep(50 * time.Millisecond)

	// Verify seat is cleared
	seat = server.FindPlayerSeat(&session.Token)
	if seat != nil {
		t.Errorf("expected player seat to be cleared on disconnect, but found in seat %d", seat.Index)
	}

	// Verify session was updated (TableID and SeatIndex should be nil)
	updatedSession, _ := server.sessionManager.GetSession(session.Token)
	if updatedSession.TableID != nil {
		t.Errorf("expected session TableID to be nil after disconnect, got %v", updatedSession.TableID)
	}
	if updatedSession.SeatIndex != nil {
		t.Errorf("expected session SeatIndex to be nil after disconnect, got %v", updatedSession.SeatIndex)
	}
}

// TestHandleDisconnectNoSeat tests that disconnect doesn't error when player has no seat
func TestHandleDisconnectNoSeat(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create a session
	session, err := server.sessionManager.CreateSession("Player")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session.Token

	// Connect with valid token
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Receive session_restored message
	_ = readMessage(t, ws)

	// Receive lobby_state message
	_ = readMessage(t, ws)

	// Verify player is not seated
	seat := server.FindPlayerSeat(&session.Token)
	if seat != nil {
		t.Fatal("expected player to not be seated before disconnect")
	}

	// Close the connection (this should not error)
	ws.Close()

	// Give server time to process disconnect
	time.Sleep(50 * time.Millisecond)

	// Just verify the test didn't panic by reaching this point
	if seat := server.FindPlayerSeat(&session.Token); seat != nil {
		t.Error("expected player to remain unseated")
	}
}

// TestDisconnectBroadcastsLobbyState tests that remaining clients receive updated lobby_state on disconnect
func TestDisconnectBroadcastsLobbyState(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create two sessions
	session1, err := server.sessionManager.CreateSession("Player1")
	if err != nil {
		t.Fatalf("failed to create session1: %v", err)
	}

	session2, err := server.sessionManager.CreateSession("Player2")
	if err != nil {
		t.Fatalf("failed to create session2: %v", err)
	}

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Connect first client
	wsURL1 := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session1.Token
	dialer := websocket.Dialer{}
	ws1, _, err := dialer.Dial(wsURL1, nil)
	if err != nil {
		t.Fatalf("failed to connect ws1: %v", err)
	}
	defer ws1.Close()

	// Receive session_restored and lobby_state
	_ = readMessage(t, ws1)
	_ = readMessage(t, ws1)

	// Connect second client
	wsURL2 := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session2.Token
	ws2, _, err := dialer.Dial(wsURL2, nil)
	if err != nil {
		t.Fatalf("failed to connect ws2: %v", err)
	}
	defer ws2.Close()

	// Receive session_restored and lobby_state for second client
	_ = readMessage(t, ws2)
	_ = readMessage(t, ws2)

	// Player 1 joins a table
	sendMessage(t, ws1, "join_table", JoinTablePayload{TableId: "table-1"})

	// Player 1 receives seat_assigned
	_ = readMessage(t, ws1)

	// Player 2 receives lobby_state broadcast
	_ = readMessage(t, ws2)

	// Player 1 disconnects
	ws1.Close()

	// Give server time to process disconnect
	time.Sleep(50 * time.Millisecond)

	// Player 2 should receive lobby_state broadcast showing table is empty
	msg := readMessage(t, ws2)
	if msg.Type != "lobby_state" {
		t.Fatalf("expected lobby_state for player2 on player1 disconnect, got %q", msg.Type)
	}

	// Verify the lobby_state shows updated seat count (back to 0)
	var payloadStr string
	err = json.Unmarshal(msg.Payload, &payloadStr)
	if err != nil {
		t.Fatalf("failed to parse lobby_state payload as string: %v", err)
	}

	var tables []map[string]interface{}
	err = json.Unmarshal([]byte(payloadStr), &tables)
	if err != nil {
		t.Fatalf("failed to parse lobby_state tables array: %v", err)
	}

	// Find table-1 in the lobby state
	var table1 map[string]interface{}
	for _, table := range tables {
		if id, ok := table["id"].(string); ok && id == "table-1" {
			table1 = table
			break
		}
	}

	if table1 == nil {
		t.Fatal("table-1 not found in lobby_state")
	}

	seatsOccupied, ok := table1["seats_occupied"].(float64)
	if !ok || int(seatsOccupied) != 0 {
		t.Errorf("expected seats_occupied to be 0 for table-1 after disconnect, got %v", seatsOccupied)
	}
}

// TestHandleLeaveTableReceivesLobbyState tests that the leaving client receives updated lobby_state
func TestHandleLeaveTableReceivesLobbyState(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create a session for the player
	session, err := server.sessionManager.CreateSession("Player")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session.Token

	// Connect with valid token
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Receive session_restored message
	_ = readMessage(t, ws)

	// Receive initial lobby_state message
	_ = readMessage(t, ws)

	// Send join_table message
	sendMessage(t, ws, "join_table", JoinTablePayload{TableId: "table-1"})

	// Receive seat_assigned message
	msg := readMessage(t, ws)
	if msg.Type != "seat_assigned" {
		t.Errorf("expected message type 'seat_assigned', got %q", msg.Type)
	}

	// Receive table_state message (sent after join)
	_ = readMessage(t, ws)

	// Send leave_table message
	sendMessage(t, ws, "leave_table", struct{}{})

	// Receive seat_cleared message
	msg = readMessage(t, ws)
	if msg.Type != "seat_cleared" {
		t.Errorf("expected first message after leave_table to be 'seat_cleared', got %q", msg.Type)
	}

	// Receive lobby_state message - THIS IS THE FIX WE'RE TESTING
	msg = readMessage(t, ws)
	if msg.Type != "lobby_state" {
		t.Errorf("expected second message after leave_table to be 'lobby_state', got %q", msg.Type)
	}

	// Verify the lobby_state shows table has 0 occupied seats
	var payloadStr string
	err = json.Unmarshal(msg.Payload, &payloadStr)
	if err != nil {
		t.Fatalf("failed to parse lobby_state payload as string: %v", err)
	}

	var tables []map[string]interface{}
	err = json.Unmarshal([]byte(payloadStr), &tables)
	if err != nil {
		t.Fatalf("failed to parse lobby_state tables array: %v", err)
	}

	// Find table-1 in the lobby state
	var table1 map[string]interface{}
	for _, table := range tables {
		if id, ok := table["id"].(string); ok && id == "table-1" {
			table1 = table
			break
		}
	}

	if table1 == nil {
		t.Fatal("table-1 not found in lobby_state after leave_table")
	}

	// Verify seats_occupied is 0 (was 1 before the player left)
	seatsOccupied, ok := table1["seats_occupied"].(float64)
	if !ok || int(seatsOccupied) != 0 {
		t.Errorf("expected seats_occupied to be 0 in lobby_state after leave_table, got %v", seatsOccupied)
	}
}

// TestGetPlayerName tests GetPlayerName method returns correct player name
func TestSessionManager_GetPlayerName(t *testing.T) {
	logger := slog.Default()
	sm := NewSessionManager(logger)

	// Create a session
	session, err := sm.CreateSession("Alice")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Get player name by token
	name, err := sm.GetPlayerName(session.Token)
	if err != nil {
		t.Errorf("GetPlayerName failed: %v", err)
	}

	if name != "Alice" {
		t.Errorf("expected name 'Alice', got %q", name)
	}
}

// TestGetPlayerName_NotFound tests GetPlayerName returns error for non-existent token
func TestSessionManager_GetPlayerName_NotFound(t *testing.T) {
	logger := slog.Default()
	sm := NewSessionManager(logger)

	_, err := sm.GetPlayerName("nonexistent-token")
	if err == nil {
		t.Fatal("expected error for non-existent token, got nil")
	}
}

// TestGetClientsAtTable tests GetClientsAtTable returns clients at specific table
func TestServer_GetClientsAtTable(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create two sessions
	session1, err := server.sessionManager.CreateSession("Player1")
	if err != nil {
		t.Fatalf("failed to create session1: %v", err)
	}

	session2, err := server.sessionManager.CreateSession("Player2")
	if err != nil {
		t.Fatalf("failed to create session2: %v", err)
	}

	// Manually seat both players at table-1
	_, _ = server.tables[0].AssignSeat(&session1.Token)
	server.sessionManager.UpdateSession(session1.Token, &server.tables[0].ID, nil)

	_, _ = server.tables[0].AssignSeat(&session2.Token)
	server.sessionManager.UpdateSession(session2.Token, &server.tables[0].ID, nil)

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Connect first client
	wsURL1 := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session1.Token
	dialer := websocket.Dialer{}
	ws1, _, err := dialer.Dial(wsURL1, nil)
	if err != nil {
		t.Fatalf("failed to connect ws1: %v", err)
	}
	defer ws1.Close()

	// Connect second client
	wsURL2 := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session2.Token
	ws2, _, err := dialer.Dial(wsURL2, nil)
	if err != nil {
		t.Fatalf("failed to connect ws2: %v", err)
	}
	defer ws2.Close()

	// Give the hub a moment to register the clients
	time.Sleep(50 * time.Millisecond)

	// Get clients at table-1
	clients := server.GetClientsAtTable(server.tables[0].ID)

	if len(clients) != 2 {
		t.Errorf("expected 2 clients at table-1, got %d", len(clients))
	}

	// Verify clients have correct tokens
	tokens := make(map[string]bool)
	for _, client := range clients {
		tokens[client.Token] = true
	}

	if !tokens[session1.Token] {
		t.Errorf("expected to find session1 token in clients")
	}

	if !tokens[session2.Token] {
		t.Errorf("expected to find session2 token in clients")
	}
}

// TestTableStatePayload tests table_state message contains all seat information
func TestTableStateMessage(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create a session for a player
	session1, err := server.sessionManager.CreateSession("Player1")
	if err != nil {
		t.Fatalf("failed to create session1: %v", err)
	}

	session2, err := server.sessionManager.CreateSession("Player2")
	if err != nil {
		t.Fatalf("failed to create session2: %v", err)
	}

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Connect first client
	wsURL1 := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session1.Token
	dialer := websocket.Dialer{}
	ws1, _, err := dialer.Dial(wsURL1, nil)
	if err != nil {
		t.Fatalf("failed to connect ws1: %v", err)
	}
	defer ws1.Close()

	// Receive session_restored
	_ = readMessage(t, ws1)
	// Receive lobby_state
	_ = readMessage(t, ws1)

	// Connect second client and have it join the table
	wsURL2 := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session2.Token
	ws2, _, err := dialer.Dial(wsURL2, nil)
	if err != nil {
		t.Fatalf("failed to connect ws2: %v", err)
	}
	defer ws2.Close()

	// Receive session_restored
	_ = readMessage(t, ws2)
	// Receive lobby_state
	_ = readMessage(t, ws2)

	// Player 1 joins table-1
	sendMessage(t, ws1, "join_table", JoinTablePayload{TableId: "table-1"})
	_ = readMessage(t, ws1) // seat_assigned
	_ = readMessage(t, ws1) // table_state
	_ = readMessage(t, ws2) // lobby_state broadcast (Player1 joined, so lobby updated)

	// Player 2 joins the same table
	sendMessage(t, ws2, "join_table", JoinTablePayload{TableId: "table-1"})

	// Player 2 should receive seat_assigned
	msg := readMessage(t, ws2)
	if msg.Type != "seat_assigned" {
		t.Fatalf("expected seat_assigned for player2, got %q", msg.Type)
	}

	// Player 2 should receive table_state
	msg = readMessage(t, ws2)
	if msg.Type != "table_state" {
		t.Fatalf("expected table_state for player2, got %q", msg.Type)
	}

	// Parse the table_state payload
	var tableStatePayload map[string]interface{}
	err = json.Unmarshal(msg.Payload, &tableStatePayload)
	if err != nil {
		t.Fatalf("failed to parse table_state payload: %v", err)
	}

	// Verify tableId is correct
	if tableID, ok := tableStatePayload["tableId"].(string); !ok || tableID != "table-1" {
		t.Errorf("expected tableId 'table-1', got %v", tableStatePayload["tableId"])
	}

	// Verify seats array exists
	seatsArray, ok := tableStatePayload["seats"].([]interface{})
	if !ok {
		t.Fatalf("expected seats array in table_state, got %T", tableStatePayload["seats"])
	}

	if len(seatsArray) != 6 {
		t.Errorf("expected 6 seats, got %d", len(seatsArray))
	}

	// Verify seat structure
	for i, seatInterface := range seatsArray {
		seat, ok := seatInterface.(map[string]interface{})
		if !ok {
			t.Errorf("seat %d is not a map, got %T", i, seatInterface)
			continue
		}

		// Each seat should have index, playerName, and status
		if index, ok := seat["index"].(float64); !ok || int(index) != i {
			t.Errorf("seat %d has wrong index: %v", i, seat["index"])
		}

		// Check that seat has playerName and status fields
		if _, ok := seat["playerName"]; !ok {
			t.Errorf("seat %d missing playerName field", i)
		}

		if _, ok := seat["status"]; !ok {
			t.Errorf("seat %d missing status field", i)
		}
	}

	// Verify player 1 is in seat 0 with their name
	seat0 := seatsArray[0].(map[string]interface{})
	if playerName := seat0["playerName"]; playerName != "Player1" {
		t.Errorf("expected seat 0 playerName 'Player1', got %v", playerName)
	}

	// Verify player 2 is in seat 1 with their name
	seat1 := seatsArray[1].(map[string]interface{})
	if playerName := seat1["playerName"]; playerName != "Player2" {
		t.Errorf("expected seat 1 playerName 'Player2', got %v", playerName)
	}
}

// TestTableStateBroadcastOnJoin tests that table_state is broadcast to all players at table when someone joins
func TestTableStateBroadcastOnJoin(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create two sessions
	session1, err := server.sessionManager.CreateSession("Player1")
	if err != nil {
		t.Fatalf("failed to create session1: %v", err)
	}

	session2, err := server.sessionManager.CreateSession("Player2")
	if err != nil {
		t.Fatalf("failed to create session2: %v", err)
	}

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Connect first client
	wsURL1 := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session1.Token
	dialer := websocket.Dialer{}
	ws1, _, err := dialer.Dial(wsURL1, nil)
	if err != nil {
		t.Fatalf("failed to connect ws1: %v", err)
	}
	defer ws1.Close()

	// Receive session_restored and lobby_state
	_ = readMessage(t, ws1)
	_ = readMessage(t, ws1)

	// Player 1 joins table-1
	sendMessage(t, ws1, "join_table", JoinTablePayload{TableId: "table-1"})

	// Player 1 receives seat_assigned
	_ = readMessage(t, ws1)

	// Player 1 should receive table_state
	msg := readMessage(t, ws1)
	if msg.Type != "table_state" {
		t.Fatalf("expected table_state for player1 after join, got %q", msg.Type)
	}

	// Parse and verify first table_state
	var firstTableState map[string]interface{}
	err = json.Unmarshal(msg.Payload, &firstTableState)
	if err != nil {
		t.Fatalf("failed to parse table_state: %v", err)
	}

	seatsArray := firstTableState["seats"].([]interface{})
	seat0 := seatsArray[0].(map[string]interface{})
	if seat0["playerName"] != "Player1" {
		t.Errorf("expected seat 0 to have Player1, got %v", seat0["playerName"])
	}

	// Now connect second client
	wsURL2 := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session2.Token
	ws2, _, err := dialer.Dial(wsURL2, nil)
	if err != nil {
		t.Fatalf("failed to connect ws2: %v", err)
	}
	defer ws2.Close()

	// Receive session_restored and lobby_state
	_ = readMessage(t, ws2)
	_ = readMessage(t, ws2)

	// Player 2 joins table-1
	sendMessage(t, ws2, "join_table", JoinTablePayload{TableId: "table-1"})

	// Player 2 receives seat_assigned
	_ = readMessage(t, ws2)

	// Player 2 receives table_state
	msg = readMessage(t, ws2)
	if msg.Type != "table_state" {
		t.Fatalf("expected table_state for player2, got %q", msg.Type)
	}

	// Verify player2's table_state shows both players
	var player2TableState map[string]interface{}
	err = json.Unmarshal(msg.Payload, &player2TableState)
	if err != nil {
		t.Fatalf("failed to parse player2 table_state: %v", err)
	}

	seatsArray = player2TableState["seats"].([]interface{})
	seat0 = seatsArray[0].(map[string]interface{})
	seat1 := seatsArray[1].(map[string]interface{})

	if seat0["playerName"] != "Player1" {
		t.Errorf("expected seat 0 to have Player1, got %v", seat0["playerName"])
	}

	if seat1["playerName"] != "Player2" {
		t.Errorf("expected seat 1 to have Player2, got %v", seat1["playerName"])
	}

	// Player 1 should also receive table_state broadcast
	msg = readMessage(t, ws1)
	if msg.Type != "table_state" {
		t.Fatalf("expected table_state broadcast for player1, got %q", msg.Type)
	}

	// Verify player1's updated table_state shows both players
	var player1UpdatedTableState map[string]interface{}
	err = json.Unmarshal(msg.Payload, &player1UpdatedTableState)
	if err != nil {
		t.Fatalf("failed to parse updated player1 table_state: %v", err)
	}

	seatsArray = player1UpdatedTableState["seats"].([]interface{})
	seat0 = seatsArray[0].(map[string]interface{})
	seat1 = seatsArray[1].(map[string]interface{})

	if seat0["playerName"] != "Player1" {
		t.Errorf("expected seat 0 to have Player1 in broadcast, got %v", seat0["playerName"])
	}

	if seat1["playerName"] != "Player2" {
		t.Errorf("expected seat 1 to have Player2 in broadcast, got %v", seat1["playerName"])
	}
}

// TestLeaveTableBroadcastsTableStateToRemaining tests that remaining players receive table_state when someone leaves
func TestLeaveTableBroadcastsTableStateToRemaining(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create two sessions
	session1, err := server.sessionManager.CreateSession("Player1")
	if err != nil {
		t.Fatalf("failed to create session1: %v", err)
	}

	session2, err := server.sessionManager.CreateSession("Player2")
	if err != nil {
		t.Fatalf("failed to create session2: %v", err)
	}

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Connect first client
	wsURL1 := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session1.Token
	dialer := websocket.Dialer{}
	ws1, _, err := dialer.Dial(wsURL1, nil)
	if err != nil {
		t.Fatalf("failed to connect ws1: %v", err)
	}
	defer ws1.Close()

	// Receive session_restored and lobby_state
	_ = readMessage(t, ws1)
	_ = readMessage(t, ws1)

	// Connect second client
	wsURL2 := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session2.Token
	ws2, _, err := dialer.Dial(wsURL2, nil)
	if err != nil {
		t.Fatalf("failed to connect ws2: %v", err)
	}
	defer ws2.Close()

	// Receive session_restored and lobby_state for second client
	_ = readMessage(t, ws2)
	_ = readMessage(t, ws2)

	// Player 1 joins table-1
	sendMessage(t, ws1, "join_table", JoinTablePayload{TableId: "table-1"})

	// Player 1 receives seat_assigned
	_ = readMessage(t, ws1)

	// Player 1 receives table_state
	_ = readMessage(t, ws1)

	// Player 2 receives lobby_state broadcast
	_ = readMessage(t, ws2)

	// Player 2 joins table-1
	sendMessage(t, ws2, "join_table", JoinTablePayload{TableId: "table-1"})

	// Player 2 receives seat_assigned
	_ = readMessage(t, ws2)

	// Player 2 receives table_state showing both players
	_ = readMessage(t, ws2)

	// Player 1 receives table_state broadcast
	_ = readMessage(t, ws1)

	// Now Player 2 leaves the table
	sendMessage(t, ws2, "leave_table", struct{}{})

	// Player 2 receives seat_cleared
	msg := readMessage(t, ws2)
	if msg.Type != "seat_cleared" {
		t.Fatalf("expected seat_cleared for player2, got %q", msg.Type)
	}

	// Player 1 should receive table_state broadcast showing only themselves
	msg = readMessage(t, ws1)
	if msg.Type != "table_state" {
		t.Fatalf("expected table_state for player1 after player2 leaves, got %q", msg.Type)
	}

	// Parse and verify the table_state shows only player1
	var tableStatePayload map[string]interface{}
	err = json.Unmarshal(msg.Payload, &tableStatePayload)
	if err != nil {
		t.Fatalf("failed to parse table_state: %v", err)
	}

	seatsArray := tableStatePayload["seats"].([]interface{})
	seat0 := seatsArray[0].(map[string]interface{})
	seat1 := seatsArray[1].(map[string]interface{})

	// Seat 0 should have Player1
	if seat0["playerName"] != "Player1" {
		t.Errorf("expected seat 0 to have Player1, got %v", seat0["playerName"])
	}

	// Seat 1 should be empty (nil)
	if seat1["playerName"] != nil {
		t.Errorf("expected seat 1 to be empty after Player2 left, got %v", seat1["playerName"])
	}
}

// TestDisconnectBroadcastsTableStateToRemaining tests that remaining players receive table_state when someone disconnects
func TestDisconnectBroadcastsTableStateToRemaining(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create two sessions
	session1, err := server.sessionManager.CreateSession("Player1")
	if err != nil {
		t.Fatalf("failed to create session1: %v", err)
	}

	session2, err := server.sessionManager.CreateSession("Player2")
	if err != nil {
		t.Fatalf("failed to create session2: %v", err)
	}

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Connect first client
	wsURL1 := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session1.Token
	dialer := websocket.Dialer{}
	ws1, _, err := dialer.Dial(wsURL1, nil)
	if err != nil {
		t.Fatalf("failed to connect ws1: %v", err)
	}
	defer ws1.Close()

	// Receive session_restored and lobby_state
	_ = readMessage(t, ws1)
	_ = readMessage(t, ws1)

	// Connect second client
	wsURL2 := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session2.Token
	ws2, _, err := dialer.Dial(wsURL2, nil)
	if err != nil {
		t.Fatalf("failed to connect ws2: %v", err)
	}
	defer ws2.Close()

	// Receive session_restored and lobby_state for second client
	_ = readMessage(t, ws2)
	_ = readMessage(t, ws2)

	// Player 1 joins table-1
	sendMessage(t, ws1, "join_table", JoinTablePayload{TableId: "table-1"})

	// Player 1 receives seat_assigned
	_ = readMessage(t, ws1)

	// Player 1 receives table_state
	_ = readMessage(t, ws1)

	// Player 2 receives lobby_state broadcast
	_ = readMessage(t, ws2)

	// Player 2 joins table-1
	sendMessage(t, ws2, "join_table", JoinTablePayload{TableId: "table-1"})

	// Player 2 receives seat_assigned
	_ = readMessage(t, ws2)

	// Player 2 receives table_state showing both players
	_ = readMessage(t, ws2)

	// Player 1 receives table_state broadcast
	_ = readMessage(t, ws1)

	// Player 2 disconnects (closes connection)
	ws2.Close()

	// Give server time to process disconnect
	time.Sleep(50 * time.Millisecond)

	// Player 1 should receive table_state broadcast showing only themselves
	msg := readMessage(t, ws1)
	if msg.Type != "table_state" {
		t.Fatalf("expected table_state for player1 after player2 disconnects, got %q", msg.Type)
	}

	// Parse and verify the table_state shows only player1
	var tableStatePayload map[string]interface{}
	err = json.Unmarshal(msg.Payload, &tableStatePayload)
	if err != nil {
		t.Fatalf("failed to parse table_state: %v", err)
	}

	seatsArray := tableStatePayload["seats"].([]interface{})
	seat0 := seatsArray[0].(map[string]interface{})
	seat1 := seatsArray[1].(map[string]interface{})

	// Seat 0 should have Player1
	if seat0["playerName"] != "Player1" {
		t.Errorf("expected seat 0 to have Player1, got %v", seat0["playerName"])
	}

	// Seat 1 should be empty (nil)
	if seat1["playerName"] != nil {
		t.Errorf("expected seat 1 to be empty after Player2 disconnected, got %v", seat1["playerName"])
	}
}

// TestStartHandBroadcastsMessages verifies StartHand() broadcasts hand_started, blind_posted, and cards_dealt
func TestStartHandBroadcastsMessages(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create two sessions
	session1, err := server.sessionManager.CreateSession("Player1")
	if err != nil {
		t.Fatalf("failed to create session1: %v", err)
	}

	session2, err := server.sessionManager.CreateSession("Player2")
	if err != nil {
		t.Fatalf("failed to create session2: %v", err)
	}

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Connect first client
	wsURL1 := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session1.Token
	dialer := websocket.Dialer{}
	ws1, _, err := dialer.Dial(wsURL1, nil)
	if err != nil {
		t.Fatalf("failed to connect ws1: %v", err)
	}
	defer ws1.Close()

	// Receive session_restored and lobby_state
	_ = readMessage(t, ws1)
	_ = readMessage(t, ws1)

	// Connect second client
	wsURL2 := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session2.Token
	ws2, _, err := dialer.Dial(wsURL2, nil)
	if err != nil {
		t.Fatalf("failed to connect ws2: %v", err)
	}
	defer ws2.Close()

	// Receive session_restored and lobby_state for second client
	_ = readMessage(t, ws2)
	_ = readMessage(t, ws2)

	// Both players join table-1
	sendMessage(t, ws1, "join_table", JoinTablePayload{TableId: "table-1"})
	_ = readMessage(t, ws1) // seat_assigned
	_ = readMessage(t, ws1) // table_state

	_ = readMessage(t, ws2) // lobby_state broadcast

	sendMessage(t, ws2, "join_table", JoinTablePayload{TableId: "table-1"})
	_ = readMessage(t, ws2) // seat_assigned
	_ = readMessage(t, ws2) // table_state

	_ = readMessage(t, ws1) // table_state broadcast

	// Mark both players as active to allow hand start
	table := server.tables[0]
	table.mu.Lock()
	table.Seats[0].Status = "active"
	table.Seats[1].Status = "active"
	table.mu.Unlock()

	// Call StartHand()
	err = table.StartHand()
	if err != nil {
		t.Fatalf("StartHand failed: %v", err)
	}

	// Give server time to process and send broadcasts
	time.Sleep(100 * time.Millisecond)

	// Player 1 should receive: hand_started, blind_posted (SB), blind_posted (BB), cards_dealt
	msg1 := readMessage(t, ws1)
	if msg1.Type != "hand_started" {
		t.Errorf("expected first message to be 'hand_started', got %q", msg1.Type)
	}

	msg2 := readMessage(t, ws1)
	if msg2.Type != "blind_posted" {
		t.Errorf("expected second message to be 'blind_posted', got %q", msg2.Type)
	}

	msg3 := readMessage(t, ws1)
	if msg3.Type != "blind_posted" {
		t.Errorf("expected third message to be 'blind_posted', got %q", msg3.Type)
	}

	msg4 := readMessage(t, ws1)
	if msg4.Type != "cards_dealt" {
		t.Errorf("expected fourth message to be 'cards_dealt', got %q", msg4.Type)
	}

	// Player 2 should receive same messages
	msg1p2 := readMessage(t, ws2)
	if msg1p2.Type != "hand_started" {
		t.Errorf("expected first message for player2 to be 'hand_started', got %q", msg1p2.Type)
	}

	msg2p2 := readMessage(t, ws2)
	if msg2p2.Type != "blind_posted" {
		t.Errorf("expected second message for player2 to be 'blind_posted', got %q", msg2p2.Type)
	}

	msg3p2 := readMessage(t, ws2)
	if msg3p2.Type != "blind_posted" {
		t.Errorf("expected third message for player2 to be 'blind_posted', got %q", msg3p2.Type)
	}

	msg4p2 := readMessage(t, ws2)
	if msg4p2.Type != "cards_dealt" {
		t.Errorf("expected fourth message for player2 to be 'cards_dealt', got %q", msg4p2.Type)
	}
}

// TestStartHandBroadcastsCardPrivacy verifies each player only receives their own hole cards
func TestStartHandBroadcastsCardPrivacy(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create two sessions
	session1, err := server.sessionManager.CreateSession("Player1")
	if err != nil {
		t.Fatalf("failed to create session1: %v", err)
	}

	session2, err := server.sessionManager.CreateSession("Player2")
	if err != nil {
		t.Fatalf("failed to create session2: %v", err)
	}

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Connect first client
	wsURL1 := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session1.Token
	dialer := websocket.Dialer{}
	ws1, _, err := dialer.Dial(wsURL1, nil)
	if err != nil {
		t.Fatalf("failed to connect ws1: %v", err)
	}
	defer ws1.Close()

	// Receive session_restored and lobby_state
	_ = readMessage(t, ws1)
	_ = readMessage(t, ws1)

	// Connect second client
	wsURL2 := "ws" + strings.TrimPrefix(testServer.URL, "http") + "?token=" + session2.Token
	ws2, _, err := dialer.Dial(wsURL2, nil)
	if err != nil {
		t.Fatalf("failed to connect ws2: %v", err)
	}
	defer ws2.Close()

	// Receive session_restored and lobby_state for second client
	_ = readMessage(t, ws2)
	_ = readMessage(t, ws2)

	// Both players join table-1
	sendMessage(t, ws1, "join_table", JoinTablePayload{TableId: "table-1"})
	_ = readMessage(t, ws1) // seat_assigned
	_ = readMessage(t, ws1) // table_state

	_ = readMessage(t, ws2) // lobby_state broadcast

	sendMessage(t, ws2, "join_table", JoinTablePayload{TableId: "table-1"})
	_ = readMessage(t, ws2) // seat_assigned
	_ = readMessage(t, ws2) // table_state

	_ = readMessage(t, ws1) // table_state broadcast

	// Mark both players as active to allow hand start
	table := server.tables[0]
	table.mu.Lock()
	table.Seats[0].Status = "active"
	table.Seats[1].Status = "active"
	table.mu.Unlock()

	// Call StartHand()
	err = table.StartHand()
	if err != nil {
		t.Fatalf("StartHand failed: %v", err)
	}

	// Give server time to process and send broadcasts
	time.Sleep(100 * time.Millisecond)

	// Player 1 receives: hand_started, blind_posted (SB), blind_posted (BB), cards_dealt
	_ = readMessage(t, ws1)          // hand_started
	_ = readMessage(t, ws1)          // blind_posted
	_ = readMessage(t, ws1)          // blind_posted
	cardsMsg1 := readMessage(t, ws1) // cards_dealt

	// Player 2 receives same messages
	_ = readMessage(t, ws2)          // hand_started
	_ = readMessage(t, ws2)          // blind_posted
	_ = readMessage(t, ws2)          // blind_posted
	cardsMsg2 := readMessage(t, ws2) // cards_dealt

	// Parse cards for player 1
	var cardsPayload1 CardsDealtPayload
	err = json.Unmarshal(cardsMsg1.Payload, &cardsPayload1)
	if err != nil {
		t.Fatalf("failed to parse cards_dealt for player1: %v", err)
	}

	// Parse cards for player 2
	var cardsPayload2 CardsDealtPayload
	err = json.Unmarshal(cardsMsg2.Payload, &cardsPayload2)
	if err != nil {
		t.Fatalf("failed to parse cards_dealt for player2: %v", err)
	}

	// Player 1 should only have hole cards for seat 0 (their seat)
	if len(cardsPayload1.HoleCards) != 1 {
		t.Errorf("expected player1 to receive hole cards for 1 seat only, got %d seats", len(cardsPayload1.HoleCards))
	}
	if _, ok := cardsPayload1.HoleCards[0]; !ok {
		t.Errorf("expected player1 to receive hole cards for seat 0 (their seat)")
	}
	if _, ok := cardsPayload1.HoleCards[1]; ok {
		t.Errorf("player1 should not receive hole cards for seat 1 (opponent's seat)")
	}

	// Player 2 should only have hole cards for seat 1 (their seat)
	if len(cardsPayload2.HoleCards) != 1 {
		t.Errorf("expected player2 to receive hole cards for 1 seat only, got %d seats", len(cardsPayload2.HoleCards))
	}
	if _, ok := cardsPayload2.HoleCards[1]; !ok {
		t.Errorf("expected player2 to receive hole cards for seat 1 (their seat)")
	}
	if _, ok := cardsPayload2.HoleCards[0]; ok {
		t.Errorf("player2 should not receive hole cards for seat 0 (opponent's seat)")
	}
}
