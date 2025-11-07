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
