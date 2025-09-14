package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"rothira/internal/database"
)

func main() {
	// Explicitly initialize DB (instead of doing work at import time)
	database.Init()
	defer database.Close()

	mux := http.NewServeMux()

	// Simple route to confirm DB connectivity
	mux.HandleFunc("/db-ping", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := database.Client.Ping(ctx, nil); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, `{"ok":false,"error":"db unreachable"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"ok":true,"db":"reachable"}`)
	})

	// Port config
	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = ":8080"
	} else if !strings.HasPrefix(httpPort, ":") {
		httpPort = ":" + httpPort
	}

	log.Printf("Listening on %s\n", httpPort)
	log.Fatal(http.ListenAndServe(httpPort, mux))
}
