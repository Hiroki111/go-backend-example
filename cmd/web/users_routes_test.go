package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Hiroki111/go-backend-example/internal/domain"
	"github.com/Hiroki111/go-backend-example/internal/handler"
)

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
			app, db := setupTestApp(t)
			rec := executeRequest(t, app, http.MethodPost, "/register-user", test.body)

			if rec.Code != test.expectedCode {
				t.Fatalf("expected %d, got %d", test.expectedCode, rec.Code)
			}

			if test.shouldHaveNewUser {
				var user domain.User
				result := db.Where(domain.User{UserName: test.body.UserName}).First(&user)
				if result.Error != nil {
					t.Fatalf("expected user to be created, but got error: %v", result.Error)
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
