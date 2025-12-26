package main

import (
	"fmt"
	"net/http"
)

const portNumber = ":8080"

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World")
	})
	fmt.Printf("Listening on port %s", portNumber)
	http.ListenAndServe(portNumber, nil)
}
