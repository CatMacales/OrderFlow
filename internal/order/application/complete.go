package application

import (
	"context"
	"fmt"

	"github.com/CatMacales/OrderFlow/internal/order/domain"
	"github.com/google/uuid"
)

func (s *Service) Complete(ctx context.Context, id uuid.UUID) error {
	const op = "complete order"

	if err := domain.ValidateOrderID(id); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err := s.repository.Complete(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
