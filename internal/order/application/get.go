package application

import (
	"context"
	"fmt"

	"github.com/CatMacales/OrderFlow/internal/order/domain"
	"github.com/google/uuid"
)

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	const op = "get order"

	if err := domain.ValidateOrderID(id); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	order, err := s.repository.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return order, nil
}
