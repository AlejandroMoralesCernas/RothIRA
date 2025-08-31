package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"rothira/api/health"
	"rothira/api/interest"
	"rothira/api/dbhealth"
	"rothira/internal/database"
)

func main() {
	fmt.Println("Starting up the Golang Roth IRA Backend...")

	// Connect Mongo (global database.Client / database.DB)
	database.Init()
	fmt.Println("Mongo connected")

	mux := http.NewServeMux()

	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"message":"Hello, Docker! <3 ahhh"}`)
	})

	mux.HandleFunc("/health", health.HealthHandler)

	mux.HandleFunc("/random-number", func(w http.ResponseWriter, r *http.Request) {
		randomValue := rand.Intn(100)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"randomValue":%d}`, randomValue)
	})

	mux.HandleFunc("/calculate-interest", interest.InterestHandler)

	// DB health (pings Mongo)
	mux.Handle("/db-health", dbhealth.Handler(database.Client))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// Ensure format ":8080"
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	srv := &http.Server{
		Addr:              port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Println("server error:", err)
	}
}
