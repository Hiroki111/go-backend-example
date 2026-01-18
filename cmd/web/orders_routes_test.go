package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/Hiroki111/go-backend-example/internal/domain"
	"github.com/Hiroki111/go-backend-example/internal/handler"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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
				token = generateJWTByRole(t, db, domain.CustomerRole)
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

func TestGetOrders_WithSorting(t *testing.T) {
	app, db := setupTestApp(t)

	products := []domain.Product{
		{Name: "apple", PriceCents: 100},
		{Name: "banana", PriceCents: 300},
		{Name: "cherry", PriceCents: 200},
	}
	products = seedProducts(t, db, products)

	users := []domain.User{
		{UserName: "Alice", Role: domain.CustomerRole},
		{UserName: "Bob", Role: domain.CustomerRole},
	}
	users = seedUsers(t, db, users)

	orders := []domain.Order{
		{ProductID: products[0].ID, UserID: users[0].ID, PriceCents: products[0].PriceCents},
		{ProductID: products[1].ID, UserID: users[0].ID, PriceCents: products[1].PriceCents},
		{ProductID: products[2].ID, UserID: users[1].ID, PriceCents: products[2].PriceCents},
	}
	orders = seedOrders(t, db, orders)

	tests := []struct {
		testName                    string
		orderBy, sortIn             string
		expectedProductNamesInOrder []string
	}{
		{testName: "order by product_id asc", orderBy: "product_id", sortIn: "asc", expectedProductNamesInOrder: []string{"apple", "banana", "cherry"}},
		{testName: "order by product_id desc", orderBy: "product_id", sortIn: "desc", expectedProductNamesInOrder: []string{"cherry", "banana", "apple"}},
		{testName: "order by product_id", orderBy: "product_id", sortIn: "", expectedProductNamesInOrder: []string{"apple", "banana", "cherry"}},
		{testName: "order by user_id asc", orderBy: "user_id", sortIn: "asc", expectedProductNamesInOrder: []string{"apple", "banana", "cherry"}},
		{testName: "order by user_id desc", orderBy: "user_id", sortIn: "desc", expectedProductNamesInOrder: []string{"cherry", "apple", "banana"}},
		{testName: "order by user_id", orderBy: "user_id", sortIn: "", expectedProductNamesInOrder: []string{"apple", "banana", "cherry"}},
		{testName: "order by created_at asc", orderBy: "created_at", sortIn: "asc", expectedProductNamesInOrder: []string{"apple", "banana", "cherry"}},
		{testName: "order by created_at desc", orderBy: "created_at", sortIn: "desc", expectedProductNamesInOrder: []string{"cherry", "banana", "apple"}},
		{testName: "order by created_at", orderBy: "created_at", sortIn: "", expectedProductNamesInOrder: []string{"apple", "banana", "cherry"}},
	}

	for _, test := range tests {
		path := fmt.Sprintf("/orders?orderBy=%s", test.orderBy)
		if test.sortIn != "" {
			path += fmt.Sprintf("&sortIn=%s", test.sortIn)
		}

		t.Run(test.testName, func(t *testing.T) {
			token := generateJWTByRole(t, db, domain.AdminRole)
			rec := executeRequest(t, app, http.MethodGet, path, token, nil)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected code %d, got %d", http.StatusOK, rec.Code)
			}

			var resp handler.GetOrdersResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("invalid json response")
			}

			if len(resp.Items) != len(test.expectedProductNamesInOrder) {
				t.Fatalf("expected %d items, got %d", len(test.expectedProductNamesInOrder), len(resp.Items))
			}

			for i, expectedName := range test.expectedProductNamesInOrder {
				if resp.Items[i].ProductName != expectedName {
					t.Fatalf("at index %d, expected %s, got %s", i, expectedName, resp.Items[i].ProductName)
				}
			}
		})
	}
}

func TestGetOrders_WithFilteringByProductIDs(t *testing.T) {
	app, db := setupTestApp(t)

	products := []domain.Product{
		{Name: "apple"},
		{Name: "banana"},
	}
	products = seedProducts(t, db, products)

	users := []domain.User{
		{Role: domain.CustomerRole},
	}
	users = seedUsers(t, db, users)

	orders := []domain.Order{
		{ProductID: products[0].ID, UserID: users[0].ID},
		{ProductID: products[1].ID, UserID: users[0].ID},
		{ProductID: products[1].ID, UserID: users[0].ID},
	}
	orders = seedOrders(t, db, orders)

	tests := []struct {
		testName         string
		productIDs       []uint
		expectedOrderIds []uint
	}{
		{
			testName:         "Has one ID",
			productIDs:       []uint{products[0].ID},
			expectedOrderIds: []uint{orders[0].ID},
		},
		{
			testName:         "Has multiple IDs",
			productIDs:       []uint{products[0].ID, products[1].ID},
			expectedOrderIds: []uint{orders[0].ID, orders[1].ID, orders[2].ID},
		},
		{
			testName:         "Has one ID and one non-existent ID",
			productIDs:       []uint{products[0].ID, uint(len(products) + 1)},
			expectedOrderIds: []uint{orders[0].ID},
		},
		{
			testName:         "Empty product_ids param",
			productIDs:       []uint{},
			expectedOrderIds: []uint{orders[0].ID, orders[1].ID, orders[2].ID},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			productIDStrings := make([]string, len(test.productIDs))
			for i, productId := range test.productIDs {
				productIDStrings[i] = strconv.Itoa(int(productId))
			}

			path := fmt.Sprintf("/orders?product_ids=%s", strings.Join(productIDStrings, ","))
			token := generateJWTByRole(t, db, domain.AdminRole)
			rec := executeRequest(t, app, http.MethodGet, path, token, nil)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
			}

			var resp handler.GetOrdersResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("invalid json response")
			}

			if len(test.expectedOrderIds) != len(resp.Items) {
				t.Fatalf("expected %d items, got %d", len(test.expectedOrderIds), len(resp.Items))
			}

			returnedOrderIds := make([]uint, len(resp.Items))
			for i, order := range resp.Items {
				returnedOrderIds[i] = order.ID
			}
			sort.Slice(returnedOrderIds, func(i, j int) bool { return returnedOrderIds[i] < returnedOrderIds[j] })
			sort.Slice(test.expectedOrderIds, func(i, j int) bool { return test.expectedOrderIds[i] < test.expectedOrderIds[j] })

			if !reflect.DeepEqual(returnedOrderIds, test.expectedOrderIds) {
				t.Fatalf("expected order IDs %v, got %v", test.expectedOrderIds, returnedOrderIds)
			}
		})
	}
}

func TestGetOrders_WithPagination(t *testing.T) {
	app, db := setupTestApp(t)

	users := seedUsers(t, db, []domain.User{{Role: domain.CustomerRole}})

	products := make([]domain.Product, 100)
	for i := range products {
		products[i] = domain.Product{Name: strconv.Itoa(i)}
	}
	products = seedProducts(t, db, products)

	orders := make([]domain.Order, 100)
	for i := range orders {
		orders[i] = domain.Order{
			ProductID: products[i].ID,
			UserID:    users[0].ID,
		}
	}
	orders = seedOrders(t, db, orders)

	tests := []struct {
		name                                  string
		page, limit                           string
		expectedItemCount, expectedTotalCount int
		expectedFirstProductName              string
		expectedHasNext                       bool
		expectedCode                          int
	}{
		{
			name: "Use blank page and blank limit",
			page: "", limit: "",
			expectedItemCount: 20, expectedTotalCount: 100, expectedFirstProductName: "0", expectedHasNext: true, expectedCode: http.StatusOK,
		},
		{
			name: "Use page and limit",
			page: "2", limit: "5",
			expectedItemCount: 5, expectedTotalCount: 100, expectedFirstProductName: "5", expectedHasNext: true, expectedCode: http.StatusOK,
		},
		{
			name: "Use blank page and limit",
			page: "", limit: "5",
			expectedItemCount: 5, expectedTotalCount: 100, expectedFirstProductName: "0", expectedHasNext: true, expectedCode: http.StatusOK,
		},
		{
			name: "Use page and blank limit",
			page: "2", limit: "",
			expectedItemCount: 20, expectedTotalCount: 100, expectedFirstProductName: "20", expectedHasNext: true, expectedCode: http.StatusOK,
		},
		{
			name: "Page offset exceeds total count",
			page: "2", limit: "101",
			expectedItemCount: 0, expectedTotalCount: 100, expectedHasNext: false, expectedCode: http.StatusOK,
		},
		{
			name: "Use non-numeric page",
			page: "abc", limit: "", expectedCode: http.StatusBadRequest,
		},
		{
			name: "Use non-numeric limit",
			page: "", limit: "abc", expectedCode: http.StatusBadRequest,
		},
		{
			name: "Use limit that exceeds the limit",
			page: "", limit: strconv.Itoa(handler.MaxPageLimit + 1), expectedCode: http.StatusBadRequest,
		},
		{
			name: "Use page 0",
			page: "0", limit: "", expectedCode: http.StatusBadRequest,
		},
		{
			name: "Use limit 0",
			page: "", limit: "0", expectedCode: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := fmt.Sprintf("/orders?page=%s&limit=%s&orderBy=created_at&sortIn=asc", test.page, test.limit)
			token := generateJWTByRole(t, db, domain.AdminRole)
			rec := executeRequest(t, app, http.MethodGet, path, token, nil)

			if rec.Code != test.expectedCode {
				t.Fatalf("expected %d, got %d", test.expectedCode, rec.Code)
			}

			if test.expectedCode != http.StatusOK {
				return
			}

			var resp handler.GetOrdersResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("invalid json response")
			}

			if test.expectedItemCount != len(resp.Items) {
				t.Fatalf("expected %d items, got %d", test.expectedItemCount, len(resp.Items))
			}

			if test.expectedTotalCount != resp.Total {
				t.Fatalf("expected %d total items, got %d", test.expectedTotalCount, resp.Total)
			}

			if test.expectedFirstProductName != "" {
				if resp.Items[0].ProductName != test.expectedFirstProductName {
					t.Fatalf("expected first item %s, got %s", test.expectedFirstProductName, resp.Items[0].ProductName)
				}
			}

			if test.expectedHasNext != resp.HasNext {
				t.Fatalf("expected HasNext %v, got %v", test.expectedHasNext, resp.HasNext)
			}

		})
	}
}

func TestGetOrder_ById(t *testing.T) {
	app, db := setupTestApp(t)

	users := make([]domain.User, 2)
	for i := range users {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("test"), bcrypt.DefaultCost)
		users[i] = domain.User{
			UserName: "user_" + strconv.Itoa(i),
			Role:     domain.CustomerRole,
			Password: string(hashed),
		}
	}
	users = seedUsers(t, db, users)

	product := seedProducts(t, db, []domain.Product{{Name: "test", PriceCents: 100}})[0]
	orders := seedOrders(t, db, []domain.Order{{ProductID: product.ID, UserID: users[0].ID}})
	expectedCustomerName := users[0].UserName
	expectedProductName := product.Name

	tests := []struct {
		name         string
		idString     string
		getToken     func(t *testing.T, testName string) string
		expectedCode int
	}{
		{
			name:     "Order found by admin",
			idString: fmt.Sprint(orders[0].ID),
			getToken: func(t *testing.T, testName string) string {
				return generateJWTByRole(t, db, domain.AdminRole)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:     "Order found by customer",
			idString: fmt.Sprint(orders[0].ID),
			getToken: func(t *testing.T, testName string) string {
				return generateJWTByExistingUser(t, users[0])
			},
			expectedCode: http.StatusOK,
		},
		{
			name:     "Order found by wrong customer",
			idString: fmt.Sprint(orders[0].ID),
			getToken: func(t *testing.T, testName string) string {
				return generateJWTByExistingUser(t, users[1])
			},
			expectedCode: http.StatusForbidden,
		},
		{
			name:     "Order not found",
			idString: fmt.Sprint(orders[0].ID + 1),
			getToken: func(t *testing.T, testName string) string {
				return generateJWTByRole(t, db, domain.AdminRole)
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name:     "Non-numeric ID",
			idString: "abc",
			getToken: func(t *testing.T, testName string) string {
				return generateJWTByRole(t, db, domain.AdminRole)
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := fmt.Sprintf("/orders/%s", test.idString)
			token := test.getToken(t, test.name)

			rec := executeRequest(t, app, http.MethodGet, path, token, nil)

			if rec.Code != test.expectedCode {
				t.Fatalf("expected %d, got %d", test.expectedCode, rec.Code)
			}

			if test.expectedCode != http.StatusOK {
				return
			}

			var resp handler.GetOrderResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("invalid json response")
			}

			expectedId, _ := strconv.ParseUint(test.idString, 10, 64)
			if resp.Item.ID != uint(expectedId) {
				t.Fatalf("expected ID %d, got %d", expectedId, resp.Item.ID)
			}

			if resp.Item.ProductName != expectedProductName {
				t.Fatalf("expected product name %s, got %s", expectedProductName, resp.Item.ProductName)
			}

			if resp.Item.CustomerName != expectedCustomerName {
				t.Fatalf("expected customer name %s, got %s", expectedCustomerName, resp.Item.CustomerName)
			}
		})
	}
}

func TestUpdateOrder(t *testing.T) {
	type Payload struct {
		PriceCents *int64 `json:"price_cents"`
	}

	app, db := setupTestApp(t)

	product := seedProducts(t, db, []domain.Product{{Name: "test", PriceCents: 100}})[0]
	user := seedUsers(t, db, []domain.User{{UserName: "customer", Role: domain.CustomerRole}})[0]

	tests := []struct {
		testName     string
		payload      Payload
		getId        func(order domain.Order) uint
		expectedCode int
	}{
		{
			testName: "success",
			payload:  Payload{PriceCents: int64Ptr(10)},
			getId: func(order domain.Order) uint {
				return order.ID
			},
			expectedCode: http.StatusOK,
		},
		{
			testName: "success - empty payload",
			payload:  Payload{},
			getId: func(order domain.Order) uint {
				return order.ID
			},
			expectedCode: http.StatusOK,
		},
		{
			testName: "fail - non-existent ID",
			payload:  Payload{PriceCents: int64Ptr(10)},
			getId: func(order domain.Order) uint {
				return order.ID + 1
			},
			expectedCode: http.StatusNotFound,
		},
		{
			testName: "fail - negative price_cents",
			payload:  Payload{PriceCents: int64Ptr(-10)},
			getId: func(order domain.Order) uint {
				return order.ID
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			order := seedOrders(t, db, []domain.Order{{UserID: user.ID, ProductID: product.ID, PriceCents: product.PriceCents}})[0]
			id := test.getId(order)
			token := generateJWTByRole(t, db, domain.AdminRole)
			path := fmt.Sprintf("/orders/%d", id)
			rec := executeRequest(t, app, http.MethodPatch, path, token, test.payload)

			if rec.Code != test.expectedCode {
				t.Fatalf("expected code %d, got %d", test.expectedCode, rec.Code)
			}

			if rec.Code == http.StatusOK {
				var resp handler.UpdateOrderResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				var updated domain.Order
				db.First(&updated, id)

				if test.payload.PriceCents != nil {
					if updated.PriceCents != *test.payload.PriceCents {
						t.Fatalf("expected price_cents %d, got %v", *test.payload.PriceCents, updated.PriceCents)
					}
				} else {
					if updated.PriceCents != order.PriceCents {
						t.Fatalf("expected price_cents %d, got %v", order.PriceCents, updated.PriceCents)
					}
				}

				if resp.Item.PriceCents != updated.PriceCents {
					t.Fatalf("response mismatch: %d vs %d", resp.Item.PriceCents, updated.PriceCents)
				}
			}
		})
	}
}

func TestDeleteOrder(t *testing.T) {
	app, db := setupTestApp(t)

	product := seedProducts(t, db, []domain.Product{{Name: "test", PriceCents: 100}})[0]
	user := seedUsers(t, db, []domain.User{{UserName: "user", Role: domain.CustomerRole}})[0]

	tests := []struct {
		testName     string
		getId        func(order domain.Order) uint
		getToken     func() string
		expectedCode int
	}{
		{
			testName: "success",
			getId: func(order domain.Order) uint {
				return order.ID
			},
			getToken: func() string {
				return generateJWTByRole(t, db, domain.AdminRole)
			},
			expectedCode: http.StatusOK,
		},
		{
			testName: "fail - non-existent ID",
			getId: func(order domain.Order) uint {
				return order.ID + 1
			},
			getToken: func() string {
				return generateJWTByRole(t, db, domain.AdminRole)
			},
			expectedCode: http.StatusNotFound,
		},
		{
			testName: "fail - empty token",
			getId: func(order domain.Order) uint {
				return order.ID
			},
			getToken: func() string {
				return ""
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			testName: "fail - customer's token",
			getId: func(order domain.Order) uint {
				return order.ID
			},
			getToken: func() string {
				return generateJWTByRole(t, db, domain.CustomerRole)
			},
			expectedCode: http.StatusForbidden,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			order := seedOrders(t, db, []domain.Order{{ProductID: product.ID, UserID: user.ID}})[0]
			id := test.getId(order)
			token := test.getToken()
			path := fmt.Sprintf("/orders/%d", id)

			rec := executeRequest(t, app, http.MethodDelete, path, token, nil)

			if rec.Code != test.expectedCode {
				t.Fatalf("expected code %d, got %d", test.expectedCode, rec.Code)
			}

			if test.expectedCode == http.StatusOK {
				if err := db.First(&domain.Order{}, id).Error; err == nil {
					t.Fatalf("order ID %d was not deleted", id)
				} else if !errors.Is(err, gorm.ErrRecordNotFound) {
					t.Fatalf("unexpected DB error: %v", err)
				}
			}
		})
	}
}
