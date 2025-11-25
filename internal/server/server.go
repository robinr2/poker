package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

// Server represents the HTTP server with router and WebSocket support.
type Server struct {
	router         chi.Router
	logger         *slog.Logger
	upgrader       *websocket.Upgrader
	httpServer     *http.Server
	hub            *Hub
	sessionManager *SessionManager
	tables         [4]*Table
	testMode       bool // TEST MODE: When true, enables test-only API endpoints
	mu             sync.RWMutex
}

// NewServer creates and returns a new Server instance.
// If testMode is true, test-only API endpoints will be enabled.
func NewServer(logger *slog.Logger, testMode bool) *Server {
	hub := NewHub(logger)
	sessionManager := NewSessionManager(logger)
	s := &Server{
		router:   chi.NewRouter(),
		logger:   logger,
		testMode: testMode,
		upgrader: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// CheckOrigin: true allows all origins for development.
				// SECURITY: In production, this must be restricted to specific origins
				// to prevent Cross-Site WebSocket Hijacking (CSWSH) attacks.
				// Use a whitelist of allowed origins in production.
				return true
			},
		},
		hub:            hub,
		sessionManager: sessionManager,
	}

	// Preseed 4 tables
	tableNames := [4]string{"Table 1", "Table 2", "Table 3", "Table 4"}
	for i := 0; i < 4; i++ {
		tableID := fmt.Sprintf("table-%d", i+1)
		s.tables[i] = NewTable(tableID, tableNames[i], s)
	}

	s.RegisterRoutes()

	// Start the Hub's event loop in a goroutine
	go hub.Run()

	return s
}

// RegisterRoutes sets up all HTTP routes for the server.
func (s *Server) RegisterRoutes() {
	s.router.Get("/health", HealthCheckHandler(s.logger))
	s.router.HandleFunc("/ws", s.HandleWebSocket(s.hub))

	// Test-only endpoints (only registered when test mode is enabled)
	if s.testMode {
		s.router.Post("/api/test/set-deck", s.HandleSetDeck())
		s.router.Post("/api/test/reset-table", s.HandleResetTable())
		s.logger.Info("test mode enabled - registered test endpoints")
	}

	// Serve static files from web/static directory
	// NOTE: Static file routes MUST be registered AFTER API routes
	// because they use a catch-all pattern
	s.logger.Debug("registering static file routes")
	s.serveStaticFiles()
}

// HandleResetTable handles the POST /api/test/reset-table endpoint
// This endpoint is only available when POKER_TEST_MODE=true
// It resets a table to its initial state (clears hand, seats, and all state)
func (s *Server) HandleResetTable() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Verify test mode is enabled
		if !s.testMode {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Test mode is not enabled",
			})
			return
		}

		// Only accept POST
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Method not allowed",
			})
			return
		}

		// Parse request body
		var payload struct {
			TableID string `json:"tableId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Invalid JSON: " + err.Error(),
			})
			return
		}

		// Find the table
		table := s.GetTableByID(payload.TableID)
		if table == nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Table not found: " + payload.TableID,
			})
			return
		}

		// Reset the table state completely
		table.mu.Lock()
		table.CurrentHand = nil
		table.DealerSeat = nil
		table.TestDeck = nil
		// Clear all seats and reset to initial state
		for i := 0; i < len(table.Seats); i++ {
			if table.Seats[i].Token != nil {
				// Clear session's table/seat info for players who were seated
				token := *table.Seats[i].Token
				_, err := s.sessionManager.UpdateSession(token, nil, nil)
				if err != nil {
					s.logger.Warn("failed to clear session during table reset", "token", token, "error", err)
				}
			}
			table.Seats[i] = Seat{
				Index:  i,
				Token:  nil,
				Status: "empty",
				Stack:  1000, // Reset to initial stack
			}
		}
		table.mu.Unlock()

		s.logger.Info("table reset via test endpoint", "tableId", payload.TableID)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Table reset successfully",
		})
	}
}

// serveStaticFiles configures static file serving with SPA fallback to index.html
func (s *Server) serveStaticFiles() {
	// Create a file server for the static directory
	fileServer := http.FileServer(http.Dir("web/static"))

	// Create a handler that tries to serve files, then falls back to index.html for SPA
	staticHandler := func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		s.logger.Debug("static handler request", "path", path, "method", r.Method)

		// Build the full file path
		fullPath := "web/static" + path

		// Check if file exists
		fileInfo, err := os.Stat(fullPath)

		// If file exists and is not a directory, serve it directly
		if err == nil && !fileInfo.IsDir() {
			s.logger.Debug("serving static file", "path", fullPath)
			fileServer.ServeHTTP(w, r)
			return
		}

		// If it's a directory, check for index.html in that directory
		if err == nil && fileInfo.IsDir() {
			indexPath := fullPath + "/index.html"
			if _, err := os.Stat(indexPath); err == nil {
				s.logger.Debug("serving directory index", "path", indexPath)
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		// File doesn't exist - serve index.html for SPA routing
		// This allows client-side routing to work
		s.logger.Debug("serving SPA fallback for route", "path", path)
		http.ServeFile(w, r, "web/static/index.html")
	}

	// Mount the handler for all paths
	s.router.Get("/", http.HandlerFunc(staticHandler))
	s.router.Get("/*", http.HandlerFunc(staticHandler))
}

// Start starts the HTTP server on the specified address.
func (s *Server) Start(addr string) error {
	s.mu.Lock()
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}
	s.mu.Unlock()

	s.logger.Info("starting server", "addr", addr)

	err := s.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return err
}

// Shutdown gracefully shuts down the HTTP server with the given context.
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.RLock()
	httpServer := s.httpServer
	s.mu.RUnlock()

	if httpServer == nil {
		return fmt.Errorf("server not running")
	}

	return httpServer.Shutdown(ctx)
}

// Router returns the chi router for testing purposes
func (s *Server) Router() chi.Router {
	return s.router
}

// IsTestMode returns whether the server is running in test mode
func (s *Server) IsTestMode() bool {
	return s.testMode
}

// GetTableByID returns a table by its ID (thread-safe)
// Returns nil if table is not found
func (s *Server) GetTableByID(tableID string) *Table {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.tables {
		if t != nil && t.ID == tableID {
			return t
		}
	}
	return nil
}

// FindPlayerSeat searches across all tables for a player token and returns their seat (thread-safe)
// Returns a pointer to a copy of the seat if found, nil if not seated at any table
func (s *Server) FindPlayerSeat(token *string) *Seat {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Search all 4 tables for the token
	for _, table := range s.tables {
		if table == nil {
			continue
		}

		// GetSeatByToken is thread-safe (uses its own RLock)
		seat, found := table.GetSeatByToken(token)
		if found {
			return &seat
		}
	}

	// Not found in any table
	return nil
}

// HandleDisconnect handles client disconnect by clearing their seat if they were seated
func (s *Server) HandleDisconnect(token string) error {
	// Find player's seat
	playerSeat := s.FindPlayerSeat(&token)
	if playerSeat == nil {
		// Player not seated, nothing to do
		return nil
	}

	// Find the table containing the player
	var table *Table
	s.mu.RLock()
	for _, t := range s.tables {
		if t != nil {
			seat, found := t.GetSeatByToken(&token)
			if found {
				table = t
				playerSeat = &seat
				break
			}
		}
	}
	s.mu.RUnlock()

	if table == nil {
		return nil
	}

	// Clear the seat
	err := table.ClearSeat(&token)
	if err != nil {
		s.logger.Warn("failed to clear seat on disconnect", "token", token, "error", err)
		return nil // Don't error on disconnect, just log
	}

	// Update session to clear TableID and SeatIndex
	_, err = s.sessionManager.UpdateSession(token, nil, nil)
	if err != nil {
		s.logger.Warn("failed to update session on disconnect", "token", token, "error", err)
	}

	// Broadcast table_state to remaining players at the table BEFORE broadcasting lobby_state
	tableID := table.ID
	err = s.broadcastTableState(tableID, nil)
	if err != nil {
		s.logger.Warn("failed to broadcast table_state on disconnect", "error", err)
	}

	// Broadcast lobby_state to remaining clients
	err = s.broadcastLobbyState()
	if err != nil {
		s.logger.Warn("failed to broadcast lobby state on disconnect", "error", err)
	}

	s.logger.Info("player disconnected and seat cleared", "token", token, "tableId", table.ID)

	return nil
}

// GetClientsAtTable returns all clients currently at a specific table (thread-safe)
func (s *Server) GetClientsAtTable(tableID string) []*Client {
	var clients []*Client

	// Find the table
	var table *Table
	s.mu.RLock()
	for _, t := range s.tables {
		if t != nil && t.ID == tableID {
			table = t
			break
		}
	}
	s.mu.RUnlock()

	if table == nil || s.hub == nil {
		return clients
	}

	// Get all seats at the table
	table.mu.RLock()
	defer table.mu.RUnlock()

	for _, seat := range table.Seats {
		if seat.Token != nil {
			// Find the client with this token in the hub
			s.hub.mu.RLock()
			for client := range s.hub.clients {
				if client.Token == *seat.Token {
					clients = append(clients, client)
					break
				}
			}
			s.hub.mu.RUnlock()
		}
	}

	return clients
}

// BroadcastActionRequest sends an action_request message to all clients at a specific table
// It notifies them that a player needs to act
// It includes calculated minRaise and maxRaise values for raise actions
func (s *Server) BroadcastActionRequest(tableID string, seatIndex int, validActions []string, callAmount, currentBet, pot int) error {
	// Get the table to access hand information
	var table *Table
	s.mu.RLock()
	for _, t := range s.tables {
		if t != nil && t.ID == tableID {
			table = t
			break
		}
	}
	s.mu.RUnlock()

	if table == nil {
		return fmt.Errorf("table not found: %s", tableID)
	}

	// Calculate minRaise, maxRaise, and displayPot
	minRaise := 0
	maxRaise := 0
	displayPot := pot
	table.mu.RLock()
	if table.CurrentHand != nil {
		minRaise = table.CurrentHand.GetMinRaise()
		maxRaise = table.GetMaxRaise(seatIndex, table.CurrentHand)
		displayPot = table.CurrentHand.GetDisplayPot()
	}
	table.mu.RUnlock()

	// Create the action request payload
	payload := ActionRequestPayload{
		SeatIndex:    seatIndex,
		ValidActions: validActions,
		CallAmount:   callAmount,
		CurrentBet:   currentBet,
		PlayerBet:    currentBet,
		Pot:          displayPot,
		MinRaise:     minRaise,
		MaxRaise:     maxRaise,
	}

	// Marshal the payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal action_request payload: %w", err)
	}

	// Create the WebSocket message
	msg := WebSocketMessage{
		Type:    "action_request",
		Payload: payloadBytes,
	}

	// Marshal the message
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal action_request message: %w", err)
	}

	// Get all clients at the table
	clients := s.GetClientsAtTable(tableID)

	// Send to all clients at the table
	for _, client := range clients {
		select {
		case client.send <- msgBytes:
			// Message sent
		default:
			// Client's send channel is full, skip
			s.logger.Warn("client send channel full, skipping action_request", "tableId", tableID, "token", client.Token)
		}
	}

	return nil
}

// BroadcastActionResult sends an action_result message to all clients at a specific table
// It notifies them that a player has acted and provides the result
func (s *Server) BroadcastActionResult(tableID string, seatIndex int, action string, amountActed, newStack, pot int, nextActor *int, roundOver bool, roundWinner *int) error {
	// Create the action result payload
	payload := ActionResultPayload{
		SeatIndex:   seatIndex,
		Action:      action,
		AmountActed: amountActed,
		NewStack:    newStack,
		Pot:         pot,
		NextActor:   nextActor,
		RoundOver:   roundOver,
		RoundWinner: roundWinner,
	}

	// Marshal the payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal action_result payload: %w", err)
	}

	// Create the WebSocket message
	msg := WebSocketMessage{
		Type:    "action_result",
		Payload: payloadBytes,
	}

	// Marshal the message
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal action_result message: %w", err)
	}

	// Get all clients at the table
	clients := s.GetClientsAtTable(tableID)

	// Send to all clients at the table
	for _, client := range clients {
		select {
		case client.send <- msgBytes:
			// Message sent
		default:
			// Client's send channel is full, skip
			s.logger.Warn("client send channel full, skipping action_result", "tableId", tableID, "token", client.Token)
		}
	}

	return nil
}
