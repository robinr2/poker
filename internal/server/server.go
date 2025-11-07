package server

import (
	"context"
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
				// CheckOrigin: true allows all origins for development.
				// SECURITY: In production, this must be restricted to specific origins
				// to prevent Cross-Site WebSocket Hijacking (CSWSH) attacks.
				// Use a whitelist of allowed origins in production.
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

	// Serve static files from web/static directory
	s.logger.Debug("registering static file routes")
	s.serveStaticFiles()
}

// serveStaticFiles configures static file serving with SPA fallback to index.html
func (s *Server) serveStaticFiles() {
	// Create a handler for serving static files
	staticHandler := s.serveStaticHandler()
	
	// Mount the handler for both root and all subpaths
	s.router.Get("/", staticHandler)
	s.router.Get("/*", staticHandler)
}

// serveStaticHandler returns an http.HandlerFunc for serving static files with SPA fallback
func (s *Server) serveStaticHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the requested file path
		path := r.URL.Path
		s.logger.Debug("static handler", "path", path)
		
		// Handle root path
		if path == "/" {
			path = "/index.html"
		}
		
		// Try to open the file from web/static
		filePath := "web/static" + path
		s.logger.Debug("checking file", "filePath", filePath)
		fileInfo, err := os.Stat(filePath)
		
		if err == nil && !fileInfo.IsDir() {
			// File exists and is not a directory, serve it
			s.logger.Debug("serving file", "filePath", filePath)
			http.ServeFile(w, r, filePath)
			return
		}
		
		s.logger.Debug("file not found, trying SPA fallback", "err", err)
		
		// File doesn't exist, try to serve index.html (SPA fallback)
		indexPath := "web/static/index.html"
		if _, err := os.Stat(indexPath); err == nil {
			s.logger.Debug("serving SPA fallback", "indexPath", indexPath)
			http.ServeFile(w, r, indexPath)
			return
		}
		
		s.logger.Debug("no SPA fallback available")
		// No index.html fallback available
		http.Error(w, "404 page not found", http.StatusNotFound)
	}
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

