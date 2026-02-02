package repository

import (
	"github.com/Hiroki111/go-backend-example/internal/domain"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Migrate() error {
	return r.db.AutoMigrate(&domain.User{}, &domain.Product{}, &domain.Order{})
}

func (r *Repository) Init() error {
	hashed, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	adminUser := domain.User{
		UserName: "admin",
		Password: string(hashed),
		Role:     domain.AdminRole,
	}
	result := r.db.Where(domain.User{UserName: "admin"}).FirstOrCreate(&adminUser)
	return result.Error
}

func (r *Repository) DB() *gorm.DB {
	return r.db
}

func (r *Repository) withTx(tx *gorm.DB) *Repository {
	if tx == nil {
		return r
	}
	return &Repository{db: tx}
}
