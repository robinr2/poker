package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Hub manages active WebSocket clients.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	logger     *slog.Logger
	mu         sync.RWMutex
}

// Client represents a WebSocket connection.
type Client struct {
	hub   *Hub
	conn  *websocket.Conn
	send  chan []byte
	Token string
}

// NewHub creates and returns a new Hub instance.
func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
	}
}

// Run starts the Hub's main event loop to process register, unregister, and broadcast messages.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.logger.Info("client registered", "total_clients", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.logger.Info("client unregistered", "total_clients", len(h.clients))
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client's send channel is full, skip sending to this client
					h.logger.Warn("client send channel full, skipping message")
				}
			}
			h.mu.RUnlock()
		}
	}
}

// HandleWebSocket returns an HTTP handler for WebSocket upgrade and connection handling.
func (s *Server) HandleWebSocket(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			s.logger.Error("websocket upgrade failed", "error", err)
			return
		}

		client := &Client{
			hub:  hub,
			conn: conn,
			send: make(chan []byte, 256),
		}

		hub.register <- client

		s.logger.Info("new websocket connection", "remote_addr", conn.RemoteAddr())

		// Extract token from query parameter
		token := r.URL.Query().Get("token")
		shouldCloseOnInvalidToken := false
		if token != "" {
			// Try to validate the token
			session, err := s.sessionManager.GetSession(token)
			if err != nil {
				s.logger.Warn("invalid token provided", "token", token, "error", err)
				// Send error message and mark for immediate closure after sending
				client.SendError("Invalid or expired token", s.logger)
				shouldCloseOnInvalidToken = true
			} else {
				// Token is valid, set client token and prepare to send session_restored
				client.Token = token
				s.logger.Info("valid token provided", "token", token)

				// Send session_restored message after registration
				go func() {
					client.SendSessionRestored(session, s.logger)
					// Send lobby_state after session_restored
					client.SendLobbyState(s, s.logger)
				}()
			}
		}

		// Start client goroutines
		go client.readPump(s.sessionManager, s, s.logger)
		go client.writePump()

		// If token was invalid, close the connection after a short delay to let message be sent
		if shouldCloseOnInvalidToken {
			go func() {
				time.Sleep(10 * time.Millisecond)
				client.conn.Close()
				hub.unregister <- client
			}()
		}
	}
}

// readPump reads messages from the WebSocket connection.
func (c *Client) readPump(sm *SessionManager, server *Server, logger *slog.Logger) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.logger.Error("websocket error", "error", err)
			}
			return
		}

		// Parse the message as JSON
		var wsMsg WebSocketMessage
		err = json.Unmarshal(message, &wsMsg)
		if err != nil {
			c.SendError("Invalid JSON message", logger)
			continue
		}

		// Route message by type
		switch wsMsg.Type {
		case "set_name":
			err := c.HandleSetName(sm, server, logger, wsMsg.Payload)
			if err != nil {
				c.SendError(err.Error(), logger)
				logger.Warn("failed to handle set_name", "error", err)
			}
		case "join_table":
			err := c.HandleJoinTable(sm, server, logger, wsMsg.Payload)
			if err != nil {
				c.SendError(err.Error(), logger)
				logger.Warn("failed to handle join_table", "error", err)
			}
		default:
			c.SendError("Unknown message type: "+wsMsg.Type, logger)
			logger.Warn("unknown message type", "type", wsMsg.Type)
		}
	}
}

// writePump writes messages to the WebSocket connection.
func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for message := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			return
		}
	}
}
