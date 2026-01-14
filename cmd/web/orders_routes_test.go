package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Hiroki111/go-backend-example/internal/domain"
	"github.com/Hiroki111/go-backend-example/internal/handler"
)

func TestCreateOrder(t *testing.T) {
	type Payload struct {
		ProductId uint `json:"product_id"`
	}

	products := []domain.Product{
		{Name: "test", PriceCents: 100},
	}

	tests := []struct {
		testName       string
		hasToken       bool
		isProductFound bool
		expectedCode   int
	}{
		{
			testName:       "success",
			hasToken:       true,
			isProductFound: true,
			expectedCode:   http.StatusCreated,
		},
		{
			testName:       "fail - no customer token",
			hasToken:       false,
			isProductFound: true,
			expectedCode:   http.StatusUnauthorized,
		},
		{
			testName:       "fail - product not found",
			hasToken:       true,
			isProductFound: false,
			expectedCode:   http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			app, db := setupTestApp(t)
			seedProducts(t, db, products)

			var token string
			if test.hasToken {
				token = generateJWT(t, db, test.testName, domain.CustomerRole)
			}

			var productId uint
			if test.isProductFound {
				var product domain.Product
				db.Where(domain.Product{Name: products[0].Name}).First(&product)
				productId = product.ID
			}

			rec := executeRequest(t, app, http.MethodPost, "/orders", token, Payload{ProductId: productId})

			if rec.Code != test.expectedCode {
				t.Fatalf("expected code %d, got %d", test.expectedCode, rec.Code)
			}

			if rec.Code == http.StatusCreated {
				var resp handler.CreateOrderResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				var createdOrder domain.Order
				if err := db.Where(domain.Order{ProductID: productId}).First(&createdOrder).Error; err != nil {
					t.Fatalf("order with product ID %d not found in DB", productId)
				}
				if createdOrder.UserID == 0 {
					t.Fatal("expected order to have a user ID")
				}
			}
		})
	}
}
