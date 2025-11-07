package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robinr2/poker/internal/server"
)

func main() {
	// Set up structured logging with JSON format
	handler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	logger.Info("starting poker application")

	// Create and start the server
	srv := server.NewServer(logger)

	// Start server in a goroutine
	go func() {
		if err := srv.Start("127.0.0.1:8080"); err != nil {
			logger.Error("server error", "error", err)
		}
	}()

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	sig := <-sigChan
	logger.Info("shutdown signal received", "signal", sig.String())

	// Gracefully shutdown the server
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("error during server shutdown", "error", err)
	}

	logger.Info("server shutdown complete")
	os.Exit(0)
}

