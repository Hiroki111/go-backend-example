package main

import (
	"net/http"

	"github.com/Hiroki111/go-backend-example/internal/domain"
	"github.com/Hiroki111/go-backend-example/internal/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	_ "github.com/Hiroki111/go-backend-example/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

func routes(handler *handler.Handler) http.Handler {
	mux := chi.NewRouter()

	mux.Route("/", func(r chi.Router) {
		r.Use(middleware.Recoverer)

		// NOTE: "POST /register-admin" should be protected by token in real world
		r.Post("/register-admin", handler.RegisterAdmin)
		r.Post("/register-customer", handler.RegisterCustomer)
		r.Post("/login-user", handler.LoginUser)

		r.Get("/products", handler.GetProducts)
		r.Get("/products/{id}", handler.GetProductById)
		r.Post("/products", handler.RequireRole(domain.AdminRole, handler.CreateProduct))
		r.Patch("/products/{id}", handler.RequireRole(domain.AdminRole, handler.UpdateProduct))
		r.Delete("/products/{id}", handler.RequireRole(domain.AdminRole, handler.DeleteProduct))

		r.Post("/orders", handler.RequireRole(domain.CustomerRole, handler.CreateOrder))
		r.Get("/orders", handler.RequireRole(domain.AdminRole, handler.GetOrders))
		r.Get("/orders/{id}", handler.RequireToken(handler.GetOrderById))
		r.Patch("/orders/{id}", handler.RequireRole(domain.AdminRole, handler.UpdateOrder))
		r.Delete("/orders/{id}", handler.RequireRole(domain.AdminRole, handler.DeleteOrder))
	})

	// infra / public routes
	mux.Get("/ping", handler.Ping)
	mux.Handle("/metrics", promhttp.Handler())

	// Swagger route
	mux.Get("/swagger/*", httpSwagger.WrapHandler)

	return mux
}
