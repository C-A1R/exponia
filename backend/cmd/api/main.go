package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/C-A1R/exponia/backend/internal/database"
	"github.com/C-A1R/exponia/backend/internal/logger"

	"github.com/C-A1R/exponia/backend/internal/camera"
)

type healthResponse struct {
	Status string `json:"status"`
}

func main() {
	appLogger := slog.New(
		logger.NewHandler(os.Stdout, slog.LevelInfo),
	)
	slog.SetDefault(appLogger)

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		appLogger.Error("DATABASE_URL is not set")
		return
	}

	db, err := database.Open(context.Background(), databaseURL)
	if err != nil {
		appLogger.Error(
			"failed to open database",
			slog.Any("error", err),
		)
		return
	}
	defer db.Close()

	appLogger.Info("connected to DB")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler(appLogger))

	cameraRepository := camera.NewRepository(db)
	cameraHandler := camera.NewHandler(cameraRepository, appLogger)
	mux.HandleFunc("POST /api/v1/cameras", cameraHandler.Create)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	shutdownSignal, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	serverErrors := make(chan error, 1)

	go func() {
		appLogger.Info("starting HTTP server", slog.String("addr", server.Addr))
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error(
				"HTTP server failed",
				slog.Any("error", err),
			)
			return
		}

	case <-shutdownSignal.Done():
		appLogger.Info("shutdown signal received")
	}

	shutdownContext, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownContext); err != nil {
		appLogger.Error("graceful shutdown failed", slog.Any("error", err))

		if closeErr := server.Close(); closeErr != nil {
			appLogger.Error("force close HTTP server", slog.Any("error", closeErr))
		}
	}

	appLogger.Info("HTTP server stopped")
}

func healthHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		response := healthResponse{
			Status: "ok",
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error(
				"failed to encode health response",
				slog.Any("error", err),
			)
		}
	}
}
