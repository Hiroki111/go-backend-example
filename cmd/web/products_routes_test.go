package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/Hiroki111/go-backend-example/internal/domain"
	"github.com/Hiroki111/go-backend-example/internal/handler"
	"gorm.io/gorm"
)

func setupProducts(t *testing.T, db *gorm.DB) []domain.Product {
	t.Helper()

	products := make([]domain.Product, 100)
	names := []string{"apple", "banana", "cherry"}

	for i := range products {
		products[i] = domain.Product{
			Name:       names[i%3] + "-" + strconv.Itoa(i),
			PriceCents: int64(50 * i),
		}
		if result := db.Create(&domain.Product{
			Name:       products[i].Name,
			PriceCents: products[i].PriceCents,
		}); result.Error != nil {
			t.Fatal(result.Error)
		}
	}

	return products
}

func TestGetProducts_WithSorting(t *testing.T) {
	tests := []struct {
		orderBy, sortIn string
	}{
		{
			orderBy: "name",
			sortIn:  "desc",
		},
		{
			orderBy: "price_cents",
			sortIn:  "desc",
		},
		{
			orderBy: "name",
			sortIn:  "asc",
		},
		{
			orderBy: "price_cents",
			sortIn:  "asc",
		},
		{
			orderBy: "name",
			sortIn:  "",
		},
		{
			orderBy: "price_cents",
			sortIn:  "",
		},
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
			setupProducts(t, db)

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

			asc := test.sortIn == "" || test.sortIn == "asc"
			for i := 0; i < len(items)-1; i++ {
				switch test.orderBy {
				case "name":
					if asc && items[i].Name > items[i+1].Name {
						t.Fatalf("expected ascending order by name")
					}
					if !asc && items[i].Name < items[i+1].Name {
						t.Fatalf("expected descending order by name")
					}
				case "price_cents":
					if asc && items[i].PriceCents > items[i+1].PriceCents {
						t.Fatalf("expected ascending order by price")
					}
					if !asc && items[i].PriceCents < items[i+1].PriceCents {
						t.Fatalf("expected descending order by price")
					}
				}
			}
		})
	}
}
