package repository

import (
	"errors"
	"strings"

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
	return r.db.AutoMigrate(&domain.User{}, &domain.Product{})
}

func (r *Repository) Init() error {
	hashed, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	adminUser := domain.User{
		UserName: "admin",
		Password: string(hashed),
	}
	result := r.db.Where(domain.User{UserName: "admin"}).FirstOrCreate(&adminUser)
	return result.Error
}

func (r *Repository) CreateUser(data domain.User) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	result := r.db.Create(&domain.User{UserName: data.UserName, Password: string(hashed)})

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return ErrUserAlreadyExists
		}
		return result.Error
	}

	return nil
}

func (r *Repository) GetUserByCredentials(userName, password string) (*domain.User, error) {
	var user domain.User

	result := r.db.Where(domain.User{UserName: userName}).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	return &user, nil
}

type GetProductsInput struct {
	OrderBy  string
	SortIn   string
	Name     string
	MinPrice int64
	MaxPrice int64
}

func (r *Repository) GetProducts(inputs GetProductsInput) ([]domain.Product, error) {
	var result []domain.Product

	query := r.db.Model(&domain.Product{})
	query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(inputs.Name)+"%").
		Where("price_cents >= ?", inputs.MinPrice).
		Where("price_cents <= ?", inputs.MaxPrice)

	sortIn := "asc"
	if inputs.SortIn == "desc" {
		sortIn = "desc"
	}

	switch inputs.OrderBy {
	case "name":
		query = query.Order("name " + sortIn)
	case "price_cents":
		query = query.Order("price_cents " + sortIn)
	default:
		query = query.Order("created_at " + sortIn)
	}

	if err := query.Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}
