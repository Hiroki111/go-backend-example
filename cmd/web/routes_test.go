package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Hiroki111/go-backend-example/internal/handlers"
	"github.com/Hiroki111/go-backend-example/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestApp(t *testing.T) http.Handler {
	t.Helper()
	t.Setenv("SECRET_KEY", "test-secret")

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
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

	handler := handlers.NewHandler(repo)
	return routes(handler)
}

func executeRequest(
	t *testing.T,
	app http.Handler,
	method, path string,
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

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	return rec
}

func TestLoginAPI_Success(t *testing.T) {
	app := setupTestApp(t)
	body := map[string]string{
		"user_name": "admin",
		"password":  "password",
	}
	rec := executeRequest(t, app, http.MethodPost, "/login-user", body)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid json response")
	}

	if resp["access_token"] == "" {
		t.Fatalf("expected access_token in response")
	}
}

func TestLoginAPI_InvalidPassword(t *testing.T) {
	app := setupTestApp(t)
	body := map[string]string{
		"user_name": "admin",
		"password":  "wrong",
	}
	rec := executeRequest(t, app, http.MethodPost, "/login-user", body)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
