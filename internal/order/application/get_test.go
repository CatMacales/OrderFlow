package application

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/CatMacales/OrderFlow/internal/order/application/mock"
	"github.com/CatMacales/OrderFlow/internal/order/domain"
	"github.com/CatMacales/OrderFlow/internal/order/storage"
	"github.com/google/uuid"
)

func TestGet(t *testing.T) {
	ctx := t.Context()
	id := uuid.New()
	item, err := domain.NewItem("p", 1)
	if err != nil {
		t.Fatal(err)
	}
	order, err := domain.NewOrder(id, "c", []domain.Item{item})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		order *domain.Order
		err   error
	}{
		{"success", order, nil}, {"missing", nil, storage.ErrOrderNotFound}, {"repository error", nil, errors.New("unavailable")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repository := mock.NewRepositoryMock(t)
			repository.GetMock.Expect(ctx, id).Return(tc.order, tc.err)
			service := NewService(repository, nil)
			got, err := service.Get(ctx, id)
			if !errors.Is(err, tc.err) || !reflect.DeepEqual(got, tc.order) {
				t.Fatalf("order=%+v err=%v", got, err)
			}
			if repository.GetAfterCounter() != 1 {
				t.Fatal("expected one repository call")
			}
		})
	}
}

func TestGetInvalidID(t *testing.T) {
	wrongVersion := uuid.MustParse("550e8400-e29b-11d4-a716-446655440000")
	wrongVariant := uuid.New()
	wrongVariant[8] = 0
	for _, id := range []uuid.UUID{uuid.Nil, wrongVersion, wrongVariant} {
		t.Run(id.String(), func(t *testing.T) {
			repository := mock.NewRepositoryMock(t)
			service := NewService(repository, nil)
			order, err := service.Get(context.Background(), id)
			if order != nil || !errors.Is(err, domain.ErrInvalidArgument) {
				t.Fatalf("order=%v err=%v", order, err)
			}
			if repository.GetAfterCounter() != 0 {
				t.Fatal("invalid ID was sent to repository")
			}
		})
	}
}
