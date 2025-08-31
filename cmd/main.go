package main

import (
	"log"
	"strings"
	"fmt"
	"net/http"
	"rothira/api/health"
	"math/rand"
	"rothira/api/interest"
	"os"
)

type CalculationRequest struct {
	Income float64 `json:"income"`
}

type CalculationResponse struct {
	Outcome float64 `json:"outcome"`
	Message string  `json:"message"`
}

func main() {
	fmt.Print("Starting up the Golang Roth IRA Backend...\n")

	mux := http.NewServeMux()

	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"message": "Hello, Docker! <3 ahhh"}`)
	})

	mux.HandleFunc("/health", health.HealthHandler)

	mux.HandleFunc("/random-number", func(w http.ResponseWriter, r *http.Request) {
		randomValue := rand.Intn(100)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"randomValue": %d}`, randomValue)
	})

	mux.HandleFunc("/calculate-interest", interest.InterestHandler)

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = ":8080"
	} else if !strings.HasPrefix(httpPort, ":") {
		httpPort = ":" + httpPort
	}

	log.Printf("Listening on %s\n", httpPort)
	if err := http.ListenAndServe(httpPort, mux); err != nil {
		log.Fatal(err)
}
}
