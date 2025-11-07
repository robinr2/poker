package main

import (
	"fmt"
	"log/slog"
	"net/http/httptest"
	
	"github.com/robinr2/poker/internal/server"
)

func main() {
	logger := slog.New(slog.NewTextHandler(&writerAdapter{}, &slog.HandlerOptions{Level: slog.LevelDebug}))
	srv := server.NewServer(logger)
	
	// Test root path
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	
	fmt.Println("=== TEST ROOT PATH ===")
	srv.Router().ServeHTTP(w, req)
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Body: %s\n", w.Body.String())
	
	// Test /health
	req = httptest.NewRequest("GET", "/health", nil)
	w = httptest.NewRecorder()
	fmt.Println("\n=== TEST /health PATH ===")
	srv.Router().ServeHTTP(w, req)
	fmt.Printf("Status: %d\n", w.Code)
}

type writerAdapter struct{}

func (w *writerAdapter) Write(p []byte) (n int, err error) {
	fmt.Print(string(p))
	return len(p), nil
}
