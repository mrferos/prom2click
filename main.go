package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"prom2click/internal/handler"
	"prom2click/internal/remotewrite"
	"time"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	mux := http.NewServeMux()

	server := &http.Server{
		Addr: ":8080",
		// I'll do something better for this later
		Handler: mux,
	}

	rwBatcher, err := remotewrite.NewBatcher()
	if err != nil {
		logger.Error("Could not create remote write batcher: ", err)
		os.Exit(1)
	}

	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/api/v1/write", handler.NewRemoteWriteHandler(rwBatcher))
	mux.Handle("/api/v1/read", handler.NewRemoteReadHandler())
	mux.Handle("/api/v1/query", handler.NewInstantQueryHandler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	// Channel to capture OS signals for graceful shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt) // Capture interrupt signals (Ctrl+C)

	// Run server in a separate goroutine
	go func() {
		logger.Info("Starting server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Could not listen on :8080: ", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal in main Goroutine
	<-stopChan
	logger.Info("Shutdown signal received")

	// Gracefully shutdown the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("Server forced to shutdown:", err)
		os.Exit(1)
	}

	logger.Info("Server exiting gracefully")
}
