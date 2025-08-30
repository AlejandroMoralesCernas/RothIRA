package health

import (
	"net/http"
	"time"
	"rothira/api/health/response"
	"encoding/json"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	health := response.HealthResponse{
		Status:  "OK",
		Uptime:  time.Since(startTime).Truncate(time.Second).String(),
		Version: "1.0.0",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(health)
}