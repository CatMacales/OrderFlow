package application

import (
	"context"

	"github.com/CatMacales/OrderFlow/internal/order/domain"
	"github.com/google/uuid"
)

//go:generate go tool minimock -i Repository -o mock/repository_mock.go -g
type Repository interface {
	Create(ctx context.Context, order *domain.Order) error
	Get(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	Confirm(ctx context.Context, id uuid.UUID) error
	Complete(ctx context.Context, id uuid.UUID) error
	Cancel(ctx context.Context, id uuid.UUID) error
}

type Service struct {
	repository Repository
	createID   func() uuid.UUID
}

func NewService(
	repository Repository,
	createID func() uuid.UUID,
) *Service {
	return &Service{
		repository: repository,
		createID:   createID,
	}
}
