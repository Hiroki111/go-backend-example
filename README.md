# go-backend-example

Goal: Learn how Go is used to build robust, production-ready services.

🧩 Topics

- Build HTTP APIs using:
  - net/http
  - chi or gorilla/mux router
  - middleware design
- Configuration and logging (e.g., viper, zap, logrus)
- Environment variables and config management
- Graceful shutdowns and context cancellation
- Dependency injection (without frameworks!)
- Testing with testing, testify, httptest
- Makefiles and CI basics

🧱 Project Idea

GoShop: A RESTful e-commerce API
- CRUD endpoints for products, users, and orders
- Postgres + GORM or SQLx
- JWT-based authentication
- Dockerized, ready for local deployment


🏗 High-level spec

GoShop is a simple e-commerce backend (or “mini store”) exposing REST APIs to manage products, users, orders, etc. Over time, this can be extended with search, caching, payments (mocked), etc.

Entities & operations
- User: register, login, profile
- Product: CRUD (create, read, update, delete)
- Order: create order (user specifies items, quantities), view order history
- Cart (optional): add items to cart, remove, checkout
- (Optional) Inventory / Stock: decrement stock on orders
- (Optional) Payment: mock or stub payment processing

Security / auth

Use JWT for authentication:
- /login → return JWT token (access token)
- Use middleware to validate token and extract user identity for protected routes
- Password hashing (e.g. bcrypt)
- Role-based access (e.g. admin user to CRUD products) (optional extension)
- Input validation, error handling

Persistence & data
- Use PostgreSQL (via database/sql, sqlx, or GORM — you can choose).
- Migrations (e.g. with golang-migrate)
- Basic query + indexing
- Relationships (user — orders, order — items)

```
/<root>
  /cmd
    /web        ← main.go, bootstrap, config load
  /internal
    /server     ← HTTP handler registration, middleware
    /handler    ← handlers/controllers
    /service    ← business logic
    /repository ← DB interactions
    /model      ← domain structs
    /auth       ← JWT, token logic
    /config     ← config structs and load logic
  /pkg (optional) ← reusable things (logger, utils)
  /migrations
  /scripts
  go.mod
  Dockerfile
  Makefile / scripts
```

HTTP API (routes)
```
POST /api/v1/register — create user account

POST /api/v1/login — issue JWT

GET /api/v1/products — list products

GET /api/v1/products/{id} — get product

POST /api/v1/products — create product (admin)

PUT /api/v1/products/{id} — update product (admin)

DELETE /api/v1/products/{id} — delete product (admin)

POST /api/v1/orders — create order

GET /api/v1/orders — list orders for user

GET /api/v1/orders/{id} — get order
```

Deployment & tooling

- Dockerfile to containerize
- Docker Compose (Postgres + GoShop) for local dev
- Logging (structured logging, e.g. zap or logrus)
- Configuration (via environment variables or config file)
- Graceful shutdown (catch signals, cleanup, context)
- Testing:
  - Unit tests for service / repository
  - Integration tests (spin up a test Postgres)
  - HTTP handler tests using httptest

Iterative roadmap for GoShop

1. Minimal viable API
  - User registration, login
  - Product list and get
  - JWT middleware
  - Basic DB schema & repository

2. CRUD endpoints
  - Full product CRUD (create, update, delete)
  - Validation and error handling

3. Order endpoints
  - Create order, view orders
  - Enforce user owns their orders

4. Enhancements
  - Cart as intermediate state
  - Inventory / stock decrementing
  - Pagination, filtering for list endpoints
  - Search (maybe via text index)
  - Caching (Redis)

5. Polish / production concerns
  - Logging, structured logs
  - Metrics (Prometheus)
  - Health checks, readiness / liveness
  - API versioning & backward compatibility
  - Docker Compose / local setup scripts
  - Documentation (OpenAPI / Swagger)
  - Error codes & error structure consistently

Optional: Once this project is solid, refactor into microservices (splitting product, order, user into separate services).