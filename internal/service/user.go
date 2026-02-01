package service

import (
	"github.com/Hiroki111/go-backend-example/internal/auth"
	"github.com/Hiroki111/go-backend-example/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

func (s *Service) CreateUser(input domain.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(input.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}
	return s.repo.CreateUser(input, hashedPassword)
}

func (s *Service) GetUser(userName string, password string) (domain.User, error) {
	user, err := s.repo.GetUserByCredentials(userName, password)
	if err != nil {
		return domain.User{}, err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	)
	if err != nil {
		return domain.User{}, ErrInvalidCredentials
	}
	return *user, nil
}

func (s *Service) GenerateJWTToken(userId uint, role domain.UserRole) (string, error) {
	return auth.GenerateJWTToken(userId, role)
}
