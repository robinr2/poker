package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()
	
	// Register routes like in serveStaticFiles
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Handler called for path: %s\n", r.URL.Path)
	}
	
	r.Get("/", handler)
	r.Get("/*", handler)
	
	// Test root path
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	
	fmt.Println("=== TEST ROOT PATH ===")
	r.ServeHTTP(w, req)
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Body: %s\n", w.Body.String())
	
	// Test /test
	req = httptest.NewRequest("GET", "/test", nil)
	w = httptest.NewRecorder()
	fmt.Println("\n=== TEST /test PATH ===")
	r.ServeHTTP(w, req)
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Body: %s\n", w.Body.String())
}
