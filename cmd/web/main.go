package main

import (
	"fmt"
	"log"
	"net/http"
)

const portNumber = ":8080"

func main() {
	fmt.Printf("Starting application on port %s", portNumber)

	server := &http.Server{
		Addr:    portNumber,
		Handler: routes(),
	}

	err := server.ListenAndServe()
	log.Fatal(err)
}
