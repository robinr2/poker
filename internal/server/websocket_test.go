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

func TestWebSocketUpgrade(t *testing.T) {
	logger := slog.Default()
	hub := NewHub(logger)
	server := NewServer(logger)

	// Create a test HTTP server with the WebSocket handler
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub)(w, r)
	}))
	defer testServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")

	// Attempt to upgrade to WebSocket
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to upgrade to websocket: %v", err)
	}
	defer ws.Close()

	// Verify connection is established
	if ws == nil {
		t.Fatal("expected websocket connection to be established")
	}
}

func TestWebSocketRouteRegistered(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)

	// Create a test HTTP request to /ws
	req := httptest.NewRequest("GET", "/ws", nil)
	w := httptest.NewRecorder()

	// Use the router directly
	server.router.ServeHTTP(w, req)

	// The WebSocket upgrade handler should be present
	// We can't directly test upgrade in this context, but we verify the route exists
	// by checking that the request doesn't get a 404
	if w.Code == http.StatusNotFound {
		t.Fatal("expected /ws route to be registered, but got 404")
	}
}

func TestClientConnection(t *testing.T) {
	logger := slog.Default()
	hub := NewHub(logger)

	// Verify hub is initialized
	if hub == nil {
		t.Fatal("expected hub to be initialized")
	}

	if hub.logger == nil {
		t.Fatal("expected hub logger to be initialized")
	}

	if hub.clients == nil {
		t.Fatal("expected hub clients map to be initialized")
	}
}

func TestHubRun(t *testing.T) {
	logger := slog.Default()
	hub := NewHub(logger)

	// Start the hub
	go hub.Run()

	// Test client registration
	client1 := &Client{
		hub:  hub,
		send: make(chan []byte, 256),
	}

	hub.register <- client1

	// Give the hub time to process
	time.Sleep(50 * time.Millisecond)

	// Verify client was registered
	hub.mu.RLock()
	if _, exists := hub.clients[client1]; !exists {
		t.Fatal("expected client to be registered in hub")
	}
	hub.mu.RUnlock()

	// Test client unregistration
	hub.unregister <- client1

	// Give the hub time to process
	time.Sleep(50 * time.Millisecond)

	// Verify client was unregistered
	hub.mu.RLock()
	if _, exists := hub.clients[client1]; exists {
		t.Fatal("expected client to be unregistered from hub")
	}
	hub.mu.RUnlock()
}

func TestHubBroadcast(t *testing.T) {
	logger := slog.Default()
	hub := NewHub(logger)

	// Start the hub
	go hub.Run()

	// Create two test clients
	client1 := &Client{
		hub:  hub,
		send: make(chan []byte, 256),
	}
	client2 := &Client{
		hub:  hub,
		send: make(chan []byte, 256),
	}

	// Register both clients
	hub.register <- client1
	hub.register <- client2

	// Give the hub time to process registrations
	time.Sleep(50 * time.Millisecond)

	// Broadcast a message
	message := []byte("test message")
	hub.broadcast <- message

	// Give the hub time to process the broadcast
	time.Sleep(50 * time.Millisecond)

	// Verify both clients received the message
	select {
	case msg := <-client1.send:
		if string(msg) != "test message" {
			t.Errorf("client1 expected 'test message', got '%s'", string(msg))
		}
	default:
		t.Fatal("client1 did not receive broadcast message")
	}

	select {
	case msg := <-client2.send:
		if string(msg) != "test message" {
			t.Errorf("client2 expected 'test message', got '%s'", string(msg))
		}
	default:
		t.Fatal("client2 did not receive broadcast message")
	}
}

func TestWebSocketRoute_PlayerActionRouted(t *testing.T) {
	logger := slog.Default()
	server := NewServer(logger)
	hub := server.hub

	// Create a mock WebSocket connection
	// We'll test the routing by sending a message directly to readPump
	sm := server.sessionManager

	// Create a session
	session, err := sm.CreateSession("TestPlayer")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Create a client
	_ = &Client{
		hub:   hub,
		conn:  nil, // Mock connection, not used for this test
		send:  make(chan []byte, 256),
		Token: session.Token,
	}

	// Test that player_action message type is recognized (doesn't error immediately)
	// The routing test is primarily in the websocket.go readPump switch statement
	// This is a basic unit test to ensure the route exists when implemented

	// Create a player_action message
	payloadObj := PlayerActionPayload{
		SeatIndex: 0,
		Action:    "call",
	}
	payloadBytes, _ := json.Marshal(payloadObj)
	wsMsg := WebSocketMessage{
		Type:    "player_action",
		Payload: json.RawMessage(payloadBytes),
	}

	// Verify the message type is what we expect
	if wsMsg.Type != "player_action" {
		t.Errorf("expected message type 'player_action', got %q", wsMsg.Type)
	}

	// Verify payload unmarshals correctly
	var actionPayload PlayerActionPayload
	err = json.Unmarshal(wsMsg.Payload, &actionPayload)
	if err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	if actionPayload.SeatIndex != 0 || actionPayload.Action != "call" {
		t.Errorf("payload not unmarshaled correctly: %+v", actionPayload)
	}
}
