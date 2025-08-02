package main

import (
	"net/http"
	"os"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"math/rand"
)

type CalculationRequest struct {
	Income float64 `json:"income"`
}

type CalculationResponse struct {
	Outcome float64 `json:"outcome"`
	Message string  `json:"message"`
}

func main() {

	fmt.Print("starting up.....")
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		fmt.Print("hello there!")
		return c.HTML(http.StatusOK, "Hello, Docker! <3")
	})

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, struct{ Status string }{Status: "OK"})
	})

	// write a random print here
	e.GET("/random", func(c echo.Context) error {
		fmt.Println("This is a random print statement!")
		fmt.Println("Generating a random value...")
		// make a random variable and set it something
		randomValue := 42
		fmt.Println("Random value is:", randomValue)
		return c.JSON(http.StatusOK, struct{ Message string }{Message: "Random print executed!"})
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

	e.POST("/calculate", CalculationHandler)

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	e.Logger.Fatal(e.Start(":" + httpPort))
}

// Simple implementation of an integer minimum
// Adapted from: https://gobyexample.com/testing-and-benchmarking
func IntMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func CalculationHandler(c echo.Context) error {
	req := new(CalculationRequest)

	err:= c.Bind(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid Input"})
	}
	result := req.Income

	return c.JSON(http.StatusOK, CalculationResponse{
		Outcome: result,
		Message: "Calculation successful",
	})
}