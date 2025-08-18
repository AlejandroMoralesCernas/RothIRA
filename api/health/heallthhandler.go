package health

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

var startTime = time.Now()

func HealthHandler(c echo.Context) error {
	health := struct {
		Status  string `json:"status"`
		Uptime  string `json:"uptime"`
		Version string `json:"version"`
	}{
		Status:  "OK",
		Uptime:  time.Since(startTime).Truncate(time.Second).String(),
		Version: "1.0.0",
	}
	return c.JSON(http.StatusOK, health)
}
