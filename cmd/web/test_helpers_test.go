package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/Hiroki111/go-backend-example/internal/auth"
	"github.com/Hiroki111/go-backend-example/internal/cache"
	"github.com/Hiroki111/go-backend-example/internal/domain"
	"github.com/Hiroki111/go-backend-example/internal/handler"
	"github.com/Hiroki111/go-backend-example/internal/repository"
	"github.com/Hiroki111/go-backend-example/internal/service"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestApp(t *testing.T) (http.Handler, *gorm.DB) {
	t.Helper()
	t.Setenv("SECRET_KEY", "test-secret")

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		TranslateError: true,
		Logger:         logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	repo := repository.NewRepository(db)

	if err := repo.Migrate(); err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	if err := repo.Init(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	productsCache := cache.NewNoopProductsCache()
	productsCacheWarmer := cache.NewNoopProductsCacheWarmer()
	service := service.NewService(repo, productsCache, productsCacheWarmer)

	handler := handler.NewHandler(service)
	return routes(handler), db
}

func executeRequest(
	t *testing.T,
	app http.Handler,
	method, path, token string,
	body any,
) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("failed to encode body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	return rec
}

func generateJWTByRole(t *testing.T, db *gorm.DB, role domain.UserRole) string {
	t.Helper()

	hashed, err := bcrypt.GenerateFromPassword([]byte("test"), bcrypt.DefaultCost)
	require.NoError(t, err)

	user := domain.User{UserName: "user_" + strconv.FormatInt(time.Now().UnixNano(), 10), Password: string(hashed), Role: role}
	result := db.Create(&user)
	require.NoError(t, result.Error)

	token, err := auth.GenerateJWTToken(user.ID, user.Role)
	require.NoError(t, err)

	return token
}

func generateJWTByExistingUser(t *testing.T, user domain.User) string {
	t.Helper()

	token, err := auth.GenerateJWTToken(user.ID, user.Role)
	require.NoError(t, err)

	return token
}

func strPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}

func seedProducts(t *testing.T, db *gorm.DB, products []domain.Product) []domain.Product {
	t.Helper()

	seededProducts := make([]domain.Product, len(products))
	for i, product := range products {
		if result := db.Create(&product); result.Error != nil {
			t.Fatal(result.Error)
		}
		seededProducts[i] = product
	}

	return seededProducts
}

func seedUsers(t *testing.T, db *gorm.DB, users []domain.User) []domain.User {
	t.Helper()

	seededUsers := make([]domain.User, len(users))
	for i, user := range users {
		user.CreatedAt = time.Now().Add(time.Duration(i) * time.Second)
		if result := db.Create(&user); result.Error != nil {
			t.Fatal(result.Error)
		}
		seededUsers[i] = user
	}

	return seededUsers
}

func seedOrders(t *testing.T, db *gorm.DB, orders []domain.Order) []domain.Order {
	t.Helper()

	seededOrders := make([]domain.Order, len(orders))
	for i, order := range orders {
		order.CreatedAt = time.Now().Add(time.Duration(i) * time.Second)
		if result := db.Create(&order); result.Error != nil {
			t.Fatal(result.Error)
		}
		seededOrders[i] = order
	}

	return seededOrders
}
