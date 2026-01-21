package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Hiroki111/go-backend-example/internal/cache"
	"github.com/Hiroki111/go-backend-example/internal/handler"
	"github.com/Hiroki111/go-backend-example/internal/repository"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const portNumber = ":8080"

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	db, err := newPostgresDB()
	if err != nil {
		log.Fatal(err)
	}

	redisClient, err := newRedisClient()
	if err != nil {
		log.Fatal(err)
	}

	productsCache := cache.NewRedisProductsCache(
		redisClient,
	)

	repo := repository.NewRepository(db)
	if err := repo.Migrate(); err != nil {
		log.Fatal(err)
	}
	if err := repo.Init(); err != nil {
		log.Fatal(err)
	}

	h := handler.NewHandler(repo, productsCache)

	server := &http.Server{
		Addr:    portNumber,
		Handler: routes(h),
	}

	// Channel that listens for OS signals
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		fmt.Printf("Starting application on port %s\n", portNumber)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Block until signal received
	<-shutdownCh
	fmt.Println("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("failed to close redis client: %v", err)
		}
	}()

	// Attempt graceful shutdown (i.e., Stop accepting new requests, finish the ones in progress, then exit cleanly)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	fmt.Println("Server exited properly")
}

func newPostgresDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		getEnv("DB_HOST"),
		getEnv("DB_USER"),
		getEnv("DB_PASSWORD"),
		getEnv("DB_NAME"),
		getEnv("DB_PORT"),
		getEnv("DB_SSLMODE"),
		getEnv("DB_TIMEZONE"),
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
}

func newRedisClient() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_ADDR"),
		Password: getEnv("REDIS_PASSWORD"),
	})

	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancle()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}

func getEnv(key string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	panic(fmt.Sprintf("Env variable %s not found", key))
}
