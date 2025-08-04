package main

import (
	"net/http"
	"os"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"math/rand"
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

	fmt.Print("Starting up the Goland Roth IRA Backend...\n")
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, "Hello, Docker! <3")
	})

	e.GET("/health", func(c echo.Context) error {
		return health.HealthHandler(c)
	})

	// random number generator
	e.GET("/random-number", func(c echo.Context) error {
		randomValue := rand.Intn(100) // Generate a random number between 0 and 99
		return c.String(http.StatusOK, fmt.Sprintf("Your random value is: %d", randomValue))
	})

	// income * interest calculator
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

		// interest calculation
		total := req.Income * (1.00 + req.Interest)

		return c.JSON(http.StatusOK, InterestResponse{
			Total: total,
		})
	})

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	e.Logger.Fatal(e.Start(":" + httpPort))
}