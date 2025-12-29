package repository

import (
	"github.com/Hiroki111/go-backend-example/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&domain.User{})
	if err != nil {
		return nil, err
	}

	return &Repository{db: db}, nil
}

func (r *Repository) Init() error {
	adminUser := domain.User{
		UserName: "admin",
		Password: "password",
	}
	result := r.db.Where(domain.User{UserName: "admin"}).FirstOrCreate(&adminUser)
	return result.Error
}

func (r *Repository) CreateUser(data domain.User) error {
	result := r.db.Create(&domain.User{UserName: data.UserName, Password: data.Password})

	return result.Error
}
