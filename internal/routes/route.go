// internal/routes/routes.go
package routes

import (
	"context"
	"fmt"
	"net/http"
	"time"

	authapi "rothira/api/auth"
)

// Deps: things the router needs but doesn't create itself.
type Deps struct {
	Auth          *authapi.Handler             // registers /api/auth/*
	AllowedOrigin string                       // CORS origin, e.g. http://localhost:3030
	DBPing        func(ctx context.Context) error // optional real DB ping for /db-ping
}

// New builds the ServeMux, wires handlers, and wraps with CORS.
func New(d Deps) (http.Handler, error) {
	mux := http.NewServeMux()

	// --- health/DB connectivity route ---
	mux.HandleFunc("/db-ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// If caller provided a ping func, actually check Mongo
		if d.DBPing != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if err := d.DBPing(ctx); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				fmt.Fprint(w, `{"ok":false,"error":"db unreachable"}`)
				return
			}
		}
		fmt.Fprint(w, `{"ok":true,"db":"reachable"}`)
	})

	// --- auth wiring ---
	if d.Auth != nil {
		if err := d.Auth.Register(mux); err != nil {
			return nil, fmt.Errorf("auth register: %w", err)
		}
	}

	// --- wrap with CORS and return final handler ---
	return cors(d.AllowedOrigin, mux), nil
}

// simple CORS middleware (moved from main.go)
func cors(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		// fast-path preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
