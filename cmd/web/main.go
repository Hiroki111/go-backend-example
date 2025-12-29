package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Hiroki111/go-backend-example/internal/handlers"
	"github.com/Hiroki111/go-backend-example/internal/repository"
	"github.com/joho/godotenv"
)

const portNumber = ":8080"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connecting to database")
	dsn := buildDSN()
	repo, err := repository.NewRepository(dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err := repo.Migrate(); err != nil {
		log.Fatal(err)
	}
	if err := repo.Init(); err != nil {
		log.Fatal(err)
	}

	handler := handlers.NewHandler(repo)
	server := &http.Server{
		Addr:    portNumber,
		Handler: routes(handler),
	}

	fmt.Printf("Starting application on port %s\n", portNumber)
	err = server.ListenAndServe()
	log.Fatal(err)
}

func buildDSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		getEnv("DB_HOST"),
		getEnv("DB_USER"),
		getEnv("DB_PASSWORD"),
		getEnv("DB_NAME"),
		getEnv("DB_PORT"),
		getEnv("DB_SSLMODE"),
		getEnv("DB_TIMEZONE"),
	)
}

func getEnv(key string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	panic(fmt.Sprintf("Env variable %s not found", key))
}
