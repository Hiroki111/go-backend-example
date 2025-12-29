package main

import (
	"net/http"

	"github.com/Hiroki111/go-backend-example/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func routes(handler *handlers.Handler) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)

	mux.Get("/ping", handler.Ping)

	mux.Post("/register-user", handler.RegisterUser)
	mux.Post("/login-user", handler.LoginUser)

	return mux
}
