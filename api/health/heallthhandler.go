package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"rothira/api/health/response"
	"rothira/internal/database"
)

// capture start time once when package loads
var startedAt = time.Now()

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	// Ping Mongo if we have a client
	dbOK := false
	if database.Client != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := database.Client.Ping(ctx, nil); err == nil {
			dbOK = true
		}
	}

	status := "OK"
	code := http.StatusOK
	if !dbOK {
		status = "DB_DOWN"           // keep your shape; encode DB state here
		code = http.StatusServiceUnavailable
	}

	health := response.HealthResponse{
		Status:  status,
		Uptime:  time.Since(startedAt).Truncate(time.Second).String(),
		Version: "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(health)
}
