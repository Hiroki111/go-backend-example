package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Hiroki111/go-backend-example/internal/auth"
	"github.com/Hiroki111/go-backend-example/internal/domain"
	"github.com/Hiroki111/go-backend-example/internal/handler"
	"github.com/Hiroki111/go-backend-example/internal/repository"
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

	handler := handler.NewHandler(repo)
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

func generateJWT(t *testing.T, db *gorm.DB, testName string, role domain.UserRole) string {
	t.Helper()

	hashed, err := bcrypt.GenerateFromPassword([]byte("test"), bcrypt.DefaultCost)
	require.NoError(t, err)

	user := domain.User{UserName: "user_" + testName, Password: string(hashed), Role: role}
	result := db.Create(&user)
	require.NoError(t, result.Error)

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
		p := product
		if result := db.Create(&p); result.Error != nil {
			t.Fatal(result.Error)
		}
		seededProducts[i] = p
	}

	return seededProducts
}
