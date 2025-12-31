package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Hiroki111/go-backend-example/internal/handler"
	"github.com/Hiroki111/go-backend-example/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestApp(t *testing.T) (http.Handler, *repository.Repository) {
	t.Helper()
	t.Setenv("SECRET_KEY", "test-secret")

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
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
	return routes(handler), repo
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

func TestRegisterUser(t *testing.T) {
	tests := []struct {
		name              string
		body              handler.RegisterUserRequest
		expectedCode      int
		shouldHaveNewUser bool
	}{
		{
			name:              "success",
			body:              handler.RegisterUserRequest{UserName: "new user", Password: "password"},
			expectedCode:      http.StatusCreated,
			shouldHaveNewUser: true,
		},
		{
			name:              "invalid user name",
			body:              handler.RegisterUserRequest{UserName: "", Password: "password"},
			expectedCode:      http.StatusBadRequest,
			shouldHaveNewUser: false,
		},
		{
			name:              "invalid password",
			body:              handler.RegisterUserRequest{UserName: "new user", Password: ""},
			expectedCode:      http.StatusBadRequest,
			shouldHaveNewUser: false,
		},
		{
			name:              "user already exists",
			body:              handler.RegisterUserRequest{UserName: "admin", Password: "password"},
			expectedCode:      http.StatusConflict,
			shouldHaveNewUser: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app, repo := setupTestApp(t)
			rec := executeRequest(t, app, http.MethodPost, "/register-user", test.body)

			if rec.Code != test.expectedCode {
				t.Fatalf("expected %d, got %d", test.expectedCode, rec.Code)
			}

			if test.shouldHaveNewUser {
				user, err := repo.GetUserByCredentials("new user", "password")
				if err != nil {
					t.Fatalf("expected user to be created, but got error: %v", err)
				}

				if user.UserName != test.body.UserName {
					t.Fatalf("expected user name %s, got %s", test.body.UserName, user.UserName)
				}
			}
		})
	}
}

func TestLoginUser(t *testing.T) {
	tests := []struct {
		name               string
		body               handler.LoginUserRequest
		expectedCode       int
		shouldReceiveToken bool
	}{
		{
			name: "success",
			body: handler.LoginUserRequest{
				UserName: "admin",
				Password: "password",
			},
			expectedCode:       http.StatusOK,
			shouldReceiveToken: true,
		},
		{
			name: "invalid credentials",
			body: handler.LoginUserRequest{
				UserName: "admin",
				Password: "wrong",
			},
			expectedCode:       http.StatusUnauthorized,
			shouldReceiveToken: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app, _ := setupTestApp(t)
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
