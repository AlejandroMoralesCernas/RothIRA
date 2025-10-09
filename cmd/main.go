// cmd/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"	// for env vars
	"strings"

	authapi "rothira/api/auth"
	"rothira/internal/database"
	"rothira/internal/routes"
)

func main() {
	// 1) DB init/close
	database.Init()
	defer database.Close() // ensure disconnect on exit

	// 2) Build deps for routes
	usersCol := database.DB.Collection("users")
	authH := &authapi.Handler{Users: usersCol}

	// CORS allowed origin (for dev, allow localhost:3030)
	// setting to ""
	allowedOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:3030"
	}

	// 3) Build the router (inject a real DB ping)
	handler, err := routes.New(routes.Deps{
		Auth:          authH,
		AllowedOrigin: allowedOrigin,
		DBPing: func(ctx context.Context) error {
			return database.Client.Ping(ctx, nil)
		},
	})
	if err != nil {
		log.Fatalf("router build: %v", err)
	}

	// 4) Port
	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = ":8080"
	} else if !strings.HasPrefix(httpPort, ":") {
		httpPort = ":" + httpPort
	}

	log.Printf("Listening on %s (CORS origin: %s)\n", httpPort, allowedOrigin)
	log.Fatal(http.ListenAndServe(httpPort, handler))
}
