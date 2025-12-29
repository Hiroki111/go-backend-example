package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Hiroki111/go-backend-example/internal/handlers"
	"github.com/Hiroki111/go-backend-example/internal/repository"
)

const portNumber = ":8080"

func main() {
	fmt.Println("Connecting to database")
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		"localhost",
		"go_backend_user",
		"password",
		"go_backend_example",
		"5432",
		"disable",
		"UTC",
	)

	repo, err := repository.NewRepository(dsn)
	if err != nil {
		log.Fatal(err)
	}
	repo.Init()

	handler := handlers.NewHandler(repo)
	server := &http.Server{
		Addr:    portNumber,
		Handler: routes(handler),
	}

	fmt.Printf("Starting application on port %s\n", portNumber)
	err = server.ListenAndServe()
	log.Fatal(err)
}
