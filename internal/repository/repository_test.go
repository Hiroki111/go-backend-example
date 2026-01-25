package repository

import (
	"testing"

	"github.com/Hiroki111/go-backend-example/internal/domain"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestUpdateProduct_OptimisticLockConflict(t *testing.T) {
	db := setupTestDb(t)

	product := domain.Product{Name: "p1", PriceCents: 100}
	db.Create(&product)

	// Load product twice (simulate two clients)
	var p1, p2 domain.Product
	db.First(&p1, product.ID)
	db.First(&p2, product.ID)

	// First update
	p1.Name = "update-1"
	err := db.
		Model(&domain.Product{}).
		Where("id = ? AND version = ?", p1.ID, p1.Version).
		Updates(map[string]interface{}{
			"name":    p1.Name,
			"version": p1.Version + 1,
		}).Error
	require.NoError(t, err)

	// Second update with stale version
	result := db.
		Model(&domain.Product{}).
		Where("id = ? AND version = ?", p2.ID, p2.Version).
		Updates(map[string]interface{}{
			"name":    "update-2",
			"version": p2.Version + 1,
		})

	require.Equal(t, int64(0), result.RowsAffected)
}

func setupTestDb(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		TranslateError: true,
		Logger:         logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	repo := NewRepository(db)

	if err := repo.Migrate(); err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	if err := repo.Init(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	return db
}
