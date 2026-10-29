package memory

import (
	"context"
	"errors"
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
	if order == nil {
		return errors.New("save order to memory: order must not be nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exist := r.orders[order.ID()]; exist {
		return fmt.Errorf("save order to memory: %w", storage.ErrOrderAlreadyExists)
	}
	r.orders[order.ID()] = *order
	return nil
}

func (r *Repository) Get(_ context.Context, id uuid.UUID) (*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	order, exist := r.orders[id]
	if !exist {
		return nil, fmt.Errorf("get order from memory: %w", storage.ErrOrderNotFound)
	}

	return &order, nil
}

func (r *Repository) Confirm(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	order, exist := r.orders[id]
	if !exist {
		return fmt.Errorf("get order from memory to confirm: %w", storage.ErrOrderNotFound)
	}

	err := order.Confirm()
	if err != nil {
		return fmt.Errorf("confirm order in memory: %w", err)
	}

	r.orders[id] = order

	return nil
}
