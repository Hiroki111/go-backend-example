# go-backend-example

Goal: Learn how Go is used to build robust, production-ready services.

🧩 Topics

- Build HTTP APIs using:
  - net/http
  - chi
  - middleware design
- Caching by Redis
- Configuration and logging (e.g., viper, zap, logrus)
- Environment variables and config management
- Graceful shutdowns and context cancellation
- Dependency injection (without frameworks!)
- Testing with testing, testify, httptest
- Makefiles and CI basics

🧱 Project Idea

GoShop: A RESTful e-commerce API
- CRUD endpoints for products, users, and orders
- Postgres + GORM
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
- Use PostgreSQL (via GORM).
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
POST /register-admin — create a user account (admin)

POST /register-customer — create a user account (customer)

POST /login-user — issue JWT for a user

GET /products — list products

GET /products/{id} — get product

POST /products — create product (only for admins)

PATCH /products/{id} — update product (only for admins)

DELETE /products/{id} — delete product (only for admins)

POST /orders — create order (only for customers)

GET /orders — list orders for user (only for admins)

GET /orders/{id} — get order (only for admins and the customer who owns the order)

PATCH /orders/{id} — update order (only for admins)

DELETE /orders/{id} — delete order (only for admins)
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

## Note:

### How to develop and run the app

Make sure the following ones are installed:

- Go (V1.25 or higher recommended)
- Docker

If you use Windows, install MinGW-w64 (Follow [this guide](https://code.visualstudio.com/docs/cpp/config-mingw#_installing-the-mingww64-toolchain)). Then, open Git Bash, and run `export CGO_ENABLED=1`

Create `.env` by following `.env.example`.

To run the app locally for development purposes, run:
```bash
docker compose up postgres redis prometheus grafana -d
go run ./cmd/web/
```
This will run the app and infra services separately.

To build the app's image and run it with the infra services by Docker Compose, run:
```bash
docker compose up --build -d
```

(Use `docker compose stop` for stopping containers, `docker compose start` for starting them again, and `docker compose down` for removing them)

Then, run `go run ./cmd/web`. This will run the app and migrate DB tables.

Open a new teminal and run the following to seed the DB:

```bash
docker exec -i postgres psql -U go_backend_user -d go_backend_example < db/seeds/0001_seed_products.sql
```

Note that the `db/seeds` directory contains SQL files used to populate the database
with mock data for local development.

These files:
- Are **not** used in production
- Should be run manually
- May be safely deleted or modified for local testing


### Docker command cheat sheet

```bash
docker exec -it redis redis-cli

# After running the command above, try following ones:
# KEYS products:*
# KEYS product:*
# GET "<key>"
# MONITOR
# CONFIG GET maxmemory-policy
# CONFIG GET maxmemory

```

### Prometheus
- Go to http://localhost:9090/targets, make sure "goshop-backend" is up.
- Go to http://localhost:9090, enter an item (e.g. app_cache_products_hits_total) to the search box, hit Execute and click Graph

### Grafana
- Go to http://localhost:3000
- Use admin for both username and password.
- In the left side bar, click Data sources -> Add data source -> Prometheus. Use http://host.docker.internal:9090 as URL -> Save & test.
- Then, create a dashboard. In the left sidebar, click Dashboards -> New -> New dashboard -> Add visualization. Select Prometheus as the data source.

### Future enhancements

- Add proper DB migration (it can be the final step)