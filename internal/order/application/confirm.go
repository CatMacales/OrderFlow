package application

import (
	"context"
	"fmt"

	"github.com/CatMacales/OrderFlow/internal/order/domain"
	"github.com/google/uuid"
)

func (s *Service) Confirm(ctx context.Context, id uuid.UUID) error {
	const op = "confirm order"

	if err := domain.ValidateOrderID(id); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err := s.repository.Confirm(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
