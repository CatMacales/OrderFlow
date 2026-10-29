package application

import (
	"context"
	"fmt"

	"github.com/CatMacales/OrderFlow/internal/order/domain"
	"github.com/google/uuid"
)

type CreateInput struct {
	CustomerID string
	Items      []CreateItemInput
}

type CreateItemInput struct {
	ProductID string
	Quantity  int64
}

func (s *Service) Create(ctx context.Context, input CreateInput) (uuid.UUID, error) {
	domainItems := make([]domain.Item, len(input.Items))
	for i, item := range input.Items {
		newItem, err := domain.NewItem(item.ProductID, item.Quantity)
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("build order item at index %d: %w", i, err)
		}
		domainItems[i] = newItem
	}

	id := s.createID()
	order, err := domain.NewOrder(id, input.CustomerID, domainItems)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("build order: %w", err)
	}

	err = s.repository.Create(ctx, order)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("create order: %w", err)
	}

	return id, nil
}
