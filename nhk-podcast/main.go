// Sample run-helloworld is a minimal Cloud Run service.
package main

import (
	"log"
	"net/http"
	"os"
)


func main() {
	log.Print("starting server...")
	http.HandleFunc("/", Handler)

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	// Start HTTP server.
	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}