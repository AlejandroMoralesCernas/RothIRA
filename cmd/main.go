package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"rothira/api/health"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CalculationRequest struct {
	Income float64 `json:"income"`
}

type CalculationResponse struct {
	Outcome float64 `json:"outcome"`
	Message string  `json:"message"`
}

func main() {
	_ = godotenv.Load()
	
	fmt.Print("Starting up the Goland Roth IRA Backend...\n")
	fmt.Print("Starting up the Golang Roth IRA Backend...\n")

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.GET("/hello", func(c echo.Context) error {
		return c.HTML(http.StatusOK, "Hello, Docker! <3")
	})

	e.GET("/health", func(c echo.Context) error {
		return health.HealthHandler(c)
	})

	e.GET("/random-number", func(c echo.Context) error {
		randomValue := rand.Intn(100)
		return c.String(http.StatusOK, fmt.Sprintf("Your random value is: %d", randomValue))
	})

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
		return c.JSON(http.StatusOK, InterestResponse{
			Total: req.Income * (1.00 + req.Interest),
		})
	})

	// ---- Serve the React SPA at root (/) ----
	// 1) Serve static assets from the built folder
	e.Static("/", "frontend-build")

	// 2) SPA fallback ONLY when a GET request 404s and the client accepts HTML
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if he, ok := err.(*echo.HTTPError); ok &&
			he.Code == http.StatusNotFound &&
			c.Request().Method == http.MethodGet &&
			strings.Contains(c.Request().Header.Get("Accept"), "text/html") {
			_ = c.File("frontend-build/index.html")
			return
		}
		e.DefaultHTTPErrorHandler(err, c)
	}
	// -----------------------------------------

	// Port
	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}
	e.Logger.Fatal(e.Start(":" + httpPort))
}
