package main

import (
	"fmt"
	"log/slog"
	"net/http/httptest"
	
	"github.com/robinr2/poker/internal/server"
)

func main() {
	logger := slog.Default()
	srv := server.NewServer(logger)
	
	// Test root path
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	
	srv.Router().ServeHTTP(w, req)
	
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Body length: %d\n", len(w.Body.Bytes()))
	if w.Code != 200 {
		fmt.Printf("Body: %s\n", w.Body.String())
	}
}
