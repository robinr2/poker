package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

// Server represents the HTTP server with router and WebSocket support.
type Server struct {
	router      chi.Router
	logger      *slog.Logger
	upgrader    *websocket.Upgrader
	httpServer  *http.Server
	hub         *Hub
	mu          sync.RWMutex
}

// NewServer creates and returns a new Server instance.
func NewServer(logger *slog.Logger) *Server {
	hub := NewHub(logger)
	s := &Server{
		router: chi.NewRouter(),
		logger: logger,
		upgrader: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for development
				return true
			},
		},
		hub: hub,
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

