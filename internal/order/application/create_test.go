package application

import (
	"context"
	"errors"
	"testing"

	"github.com/CatMacales/OrderFlow/internal/order/application/mock"
	"github.com/CatMacales/OrderFlow/internal/order/domain"
	"github.com/google/uuid"
)

func TestCreate(t *testing.T) {
	ctx := t.Context()
	id := uuid.New()
	input := CreateInput{CustomerID: " Customer-42 ", Items: []CreateItemInput{{ProductID: " SKU-Mug ", Quantity: 2}, {ProductID: "sku-mug", Quantity: 1}}}
	storageErr := errors.New("storage unavailable")
	for _, tc := range []struct {
		name string
		err  error
	}{
		{"success", nil},
		{"repository error", storageErr}} {
		t.Run(tc.name, func(t *testing.T) {
			repository := mock.NewRepositoryMock(t)
			repository.CreateMock.Set(func(gotCtx context.Context, order *domain.Order) error {
				if gotCtx != ctx || order.ID() != id || order.CustomerID() != input.CustomerID || order.Status() != domain.StatusCreated {
					t.Fatalf("unexpected context or order: %+v", order)
				}
				items := order.Items()
				if len(items) != len(input.Items) {
					t.Fatalf("items: %v", items)
				}
				for i, item := range items {
					if item.ProductID() != input.Items[i].ProductID || item.Quantity() != input.Items[i].Quantity {
						t.Fatalf("item %d: %+v", i, item)
					}
				}
				return tc.err
			})
			service := NewService(repository, func() uuid.UUID { return id })
			gotID, err := service.Create(ctx, input)
			if !errors.Is(err, tc.err) {
				t.Fatalf("error: %v, want %v", err, tc.err)
			}
			wantID := id
			if tc.err != nil {
				wantID = uuid.Nil
			}
			if gotID != wantID || repository.CreateAfterCounter() != 1 {
				t.Fatalf("id=%v calls=%d", gotID, repository.CreateAfterCounter())
			}
		})
	}
}

func TestCreateInvalidInput(t *testing.T) {
	validItems := []CreateItemInput{{ProductID: "p", Quantity: 1}}
	for _, tc := range []struct {
		name  string
		input CreateInput
	}{
		{"empty customer", CreateInput{Items: validItems}},
		{"nil items", CreateInput{CustomerID: "c"}},
		{"empty items", CreateInput{CustomerID: "c", Items: []CreateItemInput{}}},
		{"empty product", CreateInput{CustomerID: "c", Items: []CreateItemInput{{Quantity: 1}}}},
		{"zero quantity", CreateInput{CustomerID: "c", Items: []CreateItemInput{{ProductID: "p"}}}},
		{"negative quantity", CreateInput{CustomerID: "c", Items: []CreateItemInput{{ProductID: "p", Quantity: -1}}}},
		{"duplicate product", CreateInput{CustomerID: "c", Items: []CreateItemInput{{ProductID: "p", Quantity: 1}, {ProductID: "p", Quantity: 2}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repository := mock.NewRepositoryMock(t)
			service := NewService(repository, uuid.New)
			id, err := service.Create(context.Background(), tc.input)
			if id != uuid.Nil || !errors.Is(err, domain.ErrInvalidArgument) {
				t.Fatalf("id=%v err=%v", id, err)
			}
			if repository.CreateAfterCounter() != 0 {
				t.Fatal("invalid order was sent to repository")
			}
		})
	}
}
