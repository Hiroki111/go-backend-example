package repository

import (
	"errors"

	"github.com/Hiroki111/go-backend-example/internal/domain"
	"gorm.io/gorm"
)

func (r *Repository) CreateUser(data domain.User, hashedPassword []byte) error {
	result := r.db.Create(
		&domain.User{
			UserName: data.UserName,
			Password: string(hashedPassword),
			Role:     data.Role,
		},
	)

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

	return &user, nil
}
