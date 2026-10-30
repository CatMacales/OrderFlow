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
	const op = "create order"

	domainItems := make([]domain.Item, len(input.Items))
	for i, item := range input.Items {
		newItem, err := domain.NewItem(item.ProductID, item.Quantity)
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("%s: build order item at index %d: %w", op, i, err)
		}
		domainItems[i] = newItem
	}

	id := s.createID()
	order, err := domain.NewOrder(id, input.CustomerID, domainItems)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: build order: %w", op, err)
	}

	err = s.repository.Create(ctx, order)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}
