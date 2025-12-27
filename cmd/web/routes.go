package main

import (
	"net/http"

	"github.com/Hiroki111/go-backend-example/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)

	mux.HandleFunc("/ping", handlers.Ping)

	return mux
}
