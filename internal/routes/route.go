package routes

import (
	"net/http"
	"os"
	"strings"
	"log"
)

func RouteHandler() {
	mux := http.NewServeMux()

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




