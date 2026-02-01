package repository

import (
	"errors"
	"math"
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

func (r *Repository) CreateUser(data domain.User) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	result := r.db.Create(&domain.User{UserName: data.UserName, Password: string(hashed), Role: data.Role})

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
	Offset   int
	Limit    int
}

type UpdateProductsInput struct {
	ID         uint
	Name       *string
	PriceCents *uint
}

type UpdateOrderInput struct {
	ID         uint
	PriceCents *uint
}

func (r *Repository) GetProductsWithTotalCount(inputs GetProductsInput) ([]domain.Product, int64, error) {
	var result []domain.Product
	var total int64

	query := r.db.Model(&domain.Product{})
	if inputs.Name != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(inputs.Name)+"%")
	}

	query = query.
		Where("price_cents >= ?", inputs.MinPrice).
		Where("price_cents <= ?", inputs.MaxPrice)

	sortIn := "asc"
	if inputs.SortIn == "desc" {
		sortIn = "desc"
	}

	orderBy := "created_at"
	if inputs.OrderBy == "name" || inputs.OrderBy == "price_cents" {
		orderBy = inputs.OrderBy
	}

	query = query.Order(orderBy + " " + sortIn)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// NOTE: Don't use query.Offset before query.Count
	query = query.Limit(inputs.Limit).Offset(inputs.Offset)

	if err := query.Find(&result).Error; err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

func (r *Repository) GetProductById(id uint) (domain.Product, error) {
	var product domain.Product

	err := r.db.First(&product, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Product{}, ErrItemNotFound
		}
		return domain.Product{}, err
	}

	return product, nil
}

func (r *Repository) CreateProduct(data domain.Product) (domain.Product, error) {
	product := domain.Product{Name: data.Name, PriceCents: data.PriceCents}
	result := r.db.Create(&product)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return domain.Product{}, ErrProductAlreadyExists
		}
		return domain.Product{}, result.Error
	}

	return product, nil
}

func (r *Repository) UpdateProduct(data UpdateProductsInput) (domain.Product, error) {
	var product domain.Product
	if err := r.db.First(&product, data.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Product{}, ErrItemNotFound
		}
		return domain.Product{}, err
	}

	updates := map[string]interface{}{
		"version": product.Version + 1,
	}

	if data.Name != nil {
		updates["name"] = *data.Name
	}
	if data.PriceCents != nil {
		updates["price_cents"] = *data.PriceCents
	}

	result := r.db.
		Model(&product).
		Where("id = ? AND version = ?", product.ID, product.Version).
		Updates(updates)

	if err := result.Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.Product{}, ErrProductAlreadyExists
		}
		return domain.Product{}, err
	}

	if result.RowsAffected == 0 {
		return domain.Product{}, ErrOptimisticLockFailed
	}

	return product, nil
}

func (r *Repository) DeleteProduct(id uint) error {
	result := r.db.Delete(&domain.Product{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrItemNotFound
	}

	return nil
}

func (r *Repository) CreateOrder(order domain.Order) error {
	return r.db.Create(&order).Error
}

type GetOrdersInput struct {
	OrderBy    string
	SortIn     string
	ProductIDs []uint
	Offset     int
	Limit      int
}

func (r *Repository) GetOrdersWithTotalCount(inputs GetOrdersInput) ([]domain.Order, int64, error) {
	var result []domain.Order
	var total int64

	query := r.db.Model(&domain.Order{}).Preload("Product")

	if len(inputs.ProductIDs) > 0 {
		query = query.Where("product_id IN ?", inputs.ProductIDs)
	}

	sortIn := "asc"
	if inputs.SortIn == "desc" {
		sortIn = "desc"
	}

	orderBy := "created_at"
	if inputs.OrderBy == "product_id" || inputs.OrderBy == "user_id" {
		orderBy = inputs.OrderBy
	}

	query = query.Order(orderBy + " " + sortIn)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// NOTE: Don't use query.Offset before query.Count
	query = query.Limit(inputs.Limit).Offset(inputs.Offset)

	if err := query.Find(&result).Error; err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

func (r *Repository) GetOrderById(id uint) (domain.Order, error) {
	var order domain.Order

	err := r.db.Preload("Product").Preload("User").First(&order, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Order{}, ErrItemNotFound
		}
		return domain.Order{}, err
	}

	return order, nil
}

func (r *Repository) UpdateOrder(data UpdateOrderInput) (domain.Order, error) {
	var order domain.Order
	if err := r.db.Preload("Product").Preload("User").First(&order, data.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Order{}, ErrItemNotFound
		}
		return domain.Order{}, err
	}

	if data.PriceCents != nil {
		order.PriceCents = *data.PriceCents
	}

	if err := r.db.Save(&order).Error; err != nil {
		return domain.Order{}, err
	}

	return order, nil
}

func (r *Repository) DeleteOrder(id uint) error {
	result := r.db.Delete(&domain.Order{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrItemNotFound
	}

	return nil
}

func GetDefaultQueryForProducts() GetProductsInput {
	return GetProductsInput{
		OrderBy:  "",
		SortIn:   "",
		Name:     "",
		MinPrice: 0,
		MaxPrice: math.MaxInt64,
		Offset:   0,
		Limit:    20,
	}
}
