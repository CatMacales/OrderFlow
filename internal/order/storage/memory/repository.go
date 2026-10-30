package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/CatMacales/OrderFlow/internal/order/domain"
	"github.com/CatMacales/OrderFlow/internal/order/storage"
	"github.com/google/uuid"
)

type Repository struct {
	mu     sync.RWMutex
	orders map[uuid.UUID]domain.Order
}

func NewRepository() *Repository {
	return &Repository{
		orders: make(map[uuid.UUID]domain.Order),
	}
}

func (r *Repository) Create(_ context.Context, order *domain.Order) error {
	const op = "create order in memory"

	if order == nil {
		return fmt.Errorf("%s: order must not be nil", op)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exist := r.orders[order.ID()]; exist {
		return fmt.Errorf("%s: %w", op, storage.ErrOrderAlreadyExists)
	}
	r.orders[order.ID()] = *order
	return nil
}

func (r *Repository) Get(_ context.Context, id uuid.UUID) (*domain.Order, error) {
	const op = "get order from memory"

	r.mu.RLock()
	defer r.mu.RUnlock()
	order, exist := r.orders[id]
	if !exist {
		return nil, fmt.Errorf("%s: %w", op, storage.ErrOrderNotFound)
	}

	return &order, nil
}

func (r *Repository) Confirm(_ context.Context, id uuid.UUID) error {
	const op = "confirm order in memory"

	r.mu.Lock()
	defer r.mu.Unlock()
	order, exist := r.orders[id]
	if !exist {
		return fmt.Errorf("%s: %w", op, storage.ErrOrderNotFound)
	}

	err := order.Confirm()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	r.orders[id] = order

	return nil
}
