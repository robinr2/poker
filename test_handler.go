package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
)

func serveStaticHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the requested file path
		path := r.URL.Path
		fmt.Printf("DEBUG: Requested path: %s\n", path)
		
		// Handle root path
		if path == "/" {
			path = "/index.html"
			fmt.Printf("DEBUG: Converted root to: %s\n", path)
		}
		
		// Try to open the file from web/static
		filePath := "web/static" + path
		fmt.Printf("DEBUG: Checking file: %s\n", filePath)
		fileInfo, err := os.Stat(filePath)
		
		if err == nil && !fileInfo.IsDir() {
			// File exists and is not a directory, serve it
			fmt.Printf("DEBUG: File exists, serving\n")
			http.ServeFile(w, r, filePath)
			return
		}
		
		fmt.Printf("DEBUG: File not found or is dir, trying SPA fallback: %v\n", err)
		
		// File doesn't exist, try to serve index.html (SPA fallback)
		indexPath := "web/static/index.html"
		if _, err := os.Stat(indexPath); err == nil {
			fmt.Printf("DEBUG: Serving SPA fallback\n")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			http.ServeFile(w, r, indexPath)
			return
		}
		
		fmt.Printf("DEBUG: No SPA fallback available\n")
		// No index.html fallback available
		http.Error(w, "404 page not found", http.StatusNotFound)
	}
}

func main() {
	handler := serveStaticHandler()
	
	// Test root path
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	
	handler(w, req)
	
	fmt.Printf("\nStatus: %d\n", w.Code)
	fmt.Printf("Content-Type: %s\n", w.Header().Get("Content-Type"))
	fmt.Printf("Body length: %d\n", len(w.Body.Bytes()))
}
