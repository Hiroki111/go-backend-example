package service

import (
	"context"

	"github.com/Hiroki111/go-backend-example/internal/domain"
	"github.com/Hiroki111/go-backend-example/internal/repository"
)

type GetOrderParameters struct {
	OrderBy    string
	SortIn     string
	ProductIDs []uint
	Page       int
	Limit      int
}

func (s *Service) CreateOrder(ctx context.Context, userId uint, productId uint) error {
	product, err := s.repo.GetProductById(productId)
	if err != nil {
		return err
	}

	err = s.repo.CreateOrder(domain.Order{
		UserID:     userId,
		ProductID:  product.ID,
		PriceCents: product.PriceCents,
	})

	return err
}

func (s *Service) GetOrdersWithTotalCount(ctx context.Context, params GetOrderParameters) ([]domain.Order, uint, error) {
	var offset int
	if params.Page > 0 {
		offset = (params.Page - 1) * params.Limit
	}
	inputs := repository.GetOrdersInput{
		OrderBy:    params.OrderBy,
		SortIn:     params.SortIn,
		ProductIDs: params.ProductIDs,
		Limit:      params.Limit,
		Offset:     offset,
	}
	orders, total, err := s.repo.GetOrdersWithTotalCount(inputs)
	if err != nil {
		return make([]domain.Order, 0), 0, err
	}

	return orders, uint(total), err
}

func (s *Service) GetOrderById(id uint) (domain.Order, error) {
	return s.repo.GetOrderById(id)
}

func (s *Service) UpdateOrder(input repository.UpdateOrderInput) (domain.Order, error) {
	return s.repo.UpdateOrder(input)
}

func (s *Service) DeleteOrder(id uint) error {
	return s.repo.DeleteOrder(id)
}
