package application

import (
	"context"
	"fmt"

	"github.com/CatMacales/OrderFlow/internal/order/domain"
	"github.com/google/uuid"
)

func (s *Service) Confirm(ctx context.Context, id uuid.UUID) error {
	if err := domain.ValidateOrderID(id); err != nil {
		return fmt.Errorf("validate order id: %w", err)
	}

	err := s.repository.Confirm(ctx, id)
	if err != nil {
		return fmt.Errorf("confirm order: %w", err)
	}

	return nil
}
