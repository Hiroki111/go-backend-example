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

func TestLoginAPI(t *testing.T) {
	tests := []struct {
		name               string
		body               handlers.LoginUserRequest
		expectedCode       int
		shouldReceiveToken bool
	}{
		{
			name: "success",
			body: handlers.LoginUserRequest{
				UserName: "admin",
				Password: "password",
			},
			expectedCode:       http.StatusOK,
			shouldReceiveToken: true,
		},
		{
			name: "invalid credentials",
			body: handlers.LoginUserRequest{
				UserName: "admin",
				Password: "wrong",
			},
			expectedCode:       http.StatusUnauthorized,
			shouldReceiveToken: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := setupTestApp(t)
			rec := executeRequest(t, app, http.MethodPost, "/login-user", test.body)

			if rec.Code != test.expectedCode {
				t.Fatalf("expected %d, got %d", test.expectedCode, rec.Code)
			}

			if test.shouldReceiveToken {
				var resp map[string]string
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("invalid json response")
				}

				if token, ok := resp["access_token"]; !ok || token == "" {
					t.Fatalf("expected access_token in response")
				}
			}
		})
	}
}
