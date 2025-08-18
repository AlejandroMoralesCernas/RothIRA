package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"rothira/api/health"
)

type CalculationRequest struct {
	Income float64 `json:"income"`
}

type CalculationResponse struct {
	Outcome float64 `json:"outcome"`
	Message string  `json:"message"`
}

func main() {
	fmt.Println("Starting up the Golang Roth IRA Backend...")

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Root endpoint
	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, "Hello, Docker! <3")
	})

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return health.HealthHandler(c)
	})

	// Random number generator endpoint
	e.GET("/random-number", func(c echo.Context) error {
		randomValue := rand.Intn(100)
		return c.String(http.StatusOK, fmt.Sprintf("Your random value is: %d", randomValue))
	})

	// Interest calculation endpoint
	e.POST("/calculate-interest", func(c echo.Context) error {
		type InterestRequest struct {
			Income   float64 `json:"income"`
			Interest float64 `json:"interest"`
		}

		type InterestResponse struct {
			Total float64 `json:"total"`
		}

		req := new(InterestRequest)
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
		}

		total := req.Income * (1.00 + req.Interest)

		return c.JSON(http.StatusOK, InterestResponse{
			Total: total,
		})
	})

	// Serve static frontend build and SPA fallback
	e.Static("/", "frontend-build")
	e.Any("/*", func(c echo.Context) error {
		return c.File("frontend-build/index.html")
	})

	// Determine port
	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}
	fmt.Printf("Server listening on port %s\n", httpPort)

	// Start server
	e.Logger.Fatal(e.Start(":" + httpPort))
}
