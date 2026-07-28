package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/C-A1R/exponia/backend/internal/database"
)

type healthResponse struct {
	Status string `json:"status"`
}

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := database.Open(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	log.Println("connected to PostgreSQL")


	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("starting HTTP server on %s", server.Addr)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	response := healthResponse{
		Status: "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}