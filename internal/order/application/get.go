package application

import (
	"context"
	"fmt"

	"github.com/CatMacales/OrderFlow/internal/order/domain"
	"github.com/google/uuid"
)

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	if err := domain.ValidateOrderID(id); err != nil {
		return nil, fmt.Errorf("validate order id: %w", err)
	}
	order, err := s.repository.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}

	return order, nil
}
