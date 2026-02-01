package service

import (
	"github.com/Hiroki111/go-backend-example/internal/auth"
	"github.com/Hiroki111/go-backend-example/internal/domain"
)

func (s *Service) CreateUser(input domain.User) error {
	return s.repo.CreateUser(input)
}

func (s *Service) GetUser(userName string, password string) (domain.User, error) {
	user, err := s.repo.GetUserByCredentials(userName, password)
	if err != nil {
		return domain.User{}, err
	}
	return *user, nil
}

func (s *Service) GenerateJWTToken(userId uint, role domain.UserRole) (string, error) {
	return auth.GenerateJWTToken(userId, role)
}
