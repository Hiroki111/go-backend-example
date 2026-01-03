package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"testing"

	"github.com/Hiroki111/go-backend-example/internal/domain"
	"github.com/Hiroki111/go-backend-example/internal/handler"
	"gorm.io/gorm"
)

func seedProducts(t *testing.T, db *gorm.DB, products []domain.Product) {
	t.Helper()

	for _, product := range products {
		p := product
		if result := db.Create(&p); result.Error != nil {
			t.Fatal(result.Error)
		}
	}
}

func TestGetProducts_WithSorting(t *testing.T) {
	products := []domain.Product{
		{Name: "apple", PriceCents: 100},
		{Name: "banana", PriceCents: 300},
		{Name: "cherry", PriceCents: 200},
	}

	tests := []struct {
		orderBy, sortIn             string
		expectedProductNamesInOrder []string
	}{
		{orderBy: "name", sortIn: "asc", expectedProductNamesInOrder: []string{"apple", "banana", "cherry"}},
		{orderBy: "name", sortIn: "desc", expectedProductNamesInOrder: []string{"cherry", "banana", "apple"}},
		{orderBy: "name", sortIn: "", expectedProductNamesInOrder: []string{"apple", "banana", "cherry"}},
		{orderBy: "price_cents", sortIn: "asc", expectedProductNamesInOrder: []string{"apple", "cherry", "banana"}},
		{orderBy: "price_cents", sortIn: "desc", expectedProductNamesInOrder: []string{"banana", "cherry", "apple"}},
		{orderBy: "price_cents", sortIn: "", expectedProductNamesInOrder: []string{"apple", "cherry", "banana"}},
		{orderBy: "created_at", sortIn: "asc", expectedProductNamesInOrder: []string{"apple", "banana", "cherry"}},
		{orderBy: "created_at", sortIn: "desc", expectedProductNamesInOrder: []string{"cherry", "banana", "apple"}},
		{orderBy: "created_at", sortIn: "", expectedProductNamesInOrder: []string{"apple", "banana", "cherry"}},
	}

	for _, test := range tests {
		testName := fmt.Sprintf("order by %s", test.orderBy)
		path := fmt.Sprintf("/products?orderBy=%s", test.orderBy)
		if test.sortIn != "" {
			testName += fmt.Sprintf(", sorted in %s order", test.sortIn)
			path += fmt.Sprintf("&sortIn=%s", test.sortIn)
		}

		t.Run(testName, func(t *testing.T) {
			app, db := setupTestApp(t)
			seedProducts(t, db, products)

			rec := executeRequest(t, app, http.MethodGet, path, nil)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
			}

			var resp map[string][]handler.ProductResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("invalid json response")
			}

			items, ok := resp["items"]
			if !ok {
				t.Fatalf("expected items field in response")
			}

			actualNames := make([]string, 0, len(items))
			for _, item := range items {
				actualNames = append(actualNames, item.Name)
			}

			if !reflect.DeepEqual(test.expectedProductNamesInOrder, actualNames) {
				t.Fatalf("expected %v, got %v", test.expectedProductNamesInOrder, actualNames)
			}
		})
	}
}

func TestGetProducts_WithFiltering(t *testing.T) {
	products := []domain.Product{
		{Name: "apple"},
		{Name: "banana"},
		{Name: "cherry"},
	}

	tests := []struct {
		name                 string
		keyword              string
		expectedProductNames []string
	}{
		{name: "Matching one word", keyword: "ap", expectedProductNames: []string{"apple"}},
		{name: "Matching one word - case insensitive", keyword: "Ap", expectedProductNames: []string{"apple"}},
		{name: "Matching multiple words", keyword: "a", expectedProductNames: []string{"apple", "banana"}},
		{name: "Matching nothing", keyword: "aa", expectedProductNames: []string{}},
		{name: "Empty keyword", keyword: "", expectedProductNames: []string{"apple", "banana", "cherry"}},
	}

	for _, test := range tests {
		path := fmt.Sprintf("/products?name=%s", test.keyword)

		t.Run(test.name, func(t *testing.T) {
			app, db := setupTestApp(t)
			seedProducts(t, db, products)

			rec := executeRequest(t, app, http.MethodGet, path, nil)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
			}

			var resp map[string][]handler.ProductResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("invalid json response")
			}

			items, ok := resp["items"]
			if !ok {
				t.Fatalf("expected items field in response")
			}

			if len(test.expectedProductNames) != len(items) {
				t.Fatalf("expected %d items, got %d", len(test.expectedProductNames), len(items))
			}

			actualNames := make([]string, 0, len(items))
			for _, item := range items {
				actualNames = append(actualNames, item.Name)
			}
			sort.Strings(actualNames)
			sort.Strings(test.expectedProductNames)

			if !reflect.DeepEqual(actualNames, test.expectedProductNames) {
				t.Fatalf("expected products %v, got %v", test.expectedProductNames, actualNames)
			}
		})
	}
}
