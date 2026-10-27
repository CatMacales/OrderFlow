package domain_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/CatMacales/OrderFlow/internal/order/domain"
	"github.com/google/uuid"
)

func newItem(t *testing.T, productID string, quantity int64) domain.Item {
	t.Helper()
	item, err := domain.NewItem(productID, quantity)
	if err != nil {
		t.Fatal(err)
	}
	return item
}

func newOrder(t *testing.T) *domain.Order {
	t.Helper()
	order, err := domain.NewOrder(
		uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), "customer-42",
		[]domain.Item{newItem(t, "SKU-MUG", 2), newItem(t, "catalog:notebook:7", 1)},
	)
	if err != nil {
		t.Fatal(err)
	}
	return order
}

func TestNewOrder(t *testing.T) {
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	item := newItem(t, "SKU-MUG", 2)
	tests := []struct {
		name       string
		id         uuid.UUID
		customerID string
		items      []domain.Item
		valid      bool
	}{
		{"valid", id, "customer-42", []domain.Item{item, newItem(t, "catalog:notebook:7", 1)}, true},
		{"preserve customer", id, " Customer-42 \t", []domain.Item{item}, true},
		{"whitespace customer", id, " ", []domain.Item{item}, true},
		{"distinct exact product strings", id, "c", []domain.Item{
			newItem(t, "SKU", 1), newItem(t, "sku", 1), newItem(t, " SKU ", 1),
		}, true},
		{"nil UUID", uuid.Nil, "c", []domain.Item{item}, false},
		{"UUID v1", uuid.MustParse("550e8400-e29b-11d4-a716-446655440000"), "c", []domain.Item{item}, false},
		{"UUID v7", uuid.MustParse("550e8400-e29b-71d4-a716-446655440000"), "c", []domain.Item{item}, false},
		{"wrong variant", uuid.MustParse("550e8400-e29b-41d4-0716-446655440000"), "c", []domain.Item{item}, false},
		{"empty customer", id, "", []domain.Item{item}, false},
		{"nil items", id, "c", nil, false},
		{"empty items", id, "c", []domain.Item{}, false},
		{"zero item", id, "c", []domain.Item{item, {}}, false},
		{"duplicate product", id, "c", []domain.Item{item, newItem(t, "SKU-MUG", 3)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, err := domain.NewOrder(tt.id, tt.customerID, tt.items)
			if !tt.valid {
				if !errors.Is(err, domain.ErrInvalidArgument) {
					t.Fatalf("expected INVALID_ARGUMENT, got %v", err)
				}
				if order != nil {
					t.Fatal("failed construction returned an order")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if order.ID() != tt.id || order.CustomerID() != tt.customerID || order.Status() != domain.StatusCreated || !reflect.DeepEqual(order.Items(), tt.items) {
				t.Fatalf("unexpected order: %+v", order)
			}
		})
	}
}

func TestOrderOwnsItems(t *testing.T) {
	original := newItem(t, "SKU-MUG", 2)
	input := []domain.Item{original}
	order, err := domain.NewOrder(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), "c", input)
	if err != nil {
		t.Fatal(err)
	}
	input[0] = domain.Item{}
	output := order.Items()
	if output[0] != original {
		t.Fatal("input mutation changed the order")
	}
	output[0] = domain.Item{}
	if order.Items()[0] != original {
		t.Fatal("output mutation changed the order")
	}
}

func TestOrderTransitions(t *testing.T) {
	tests := []struct {
		name    string
		initial domain.Status
		action  func(*domain.Order) error
		want    domain.Status
		invalid bool
	}{
		{"created confirm", domain.StatusCreated, (*domain.Order).Confirm, domain.StatusConfirmed, false},
		{"created complete", domain.StatusCreated, (*domain.Order).Complete, domain.StatusCreated, true},
		{"created cancel", domain.StatusCreated, (*domain.Order).Cancel, domain.StatusCancelled, false},
		{"confirmed confirm", domain.StatusConfirmed, (*domain.Order).Confirm, domain.StatusConfirmed, false},
		{"confirmed complete", domain.StatusConfirmed, (*domain.Order).Complete, domain.StatusCompleted, false},
		{"confirmed cancel", domain.StatusConfirmed, (*domain.Order).Cancel, domain.StatusCancelled, false},
		{"completed confirm", domain.StatusCompleted, (*domain.Order).Confirm, domain.StatusCompleted, false},
		{"completed complete", domain.StatusCompleted, (*domain.Order).Complete, domain.StatusCompleted, false},
		{"completed cancel", domain.StatusCompleted, (*domain.Order).Cancel, domain.StatusCompleted, true},
		{"cancelled confirm", domain.StatusCancelled, (*domain.Order).Confirm, domain.StatusCancelled, true},
		{"cancelled complete", domain.StatusCancelled, (*domain.Order).Complete, domain.StatusCancelled, true},
		{"cancelled cancel", domain.StatusCancelled, (*domain.Order).Cancel, domain.StatusCancelled, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := newOrder(t)
			// Reach each initial state through public domain actions.
			switch tt.initial {
			case domain.StatusConfirmed, domain.StatusCompleted:
				if err := order.Confirm(); err != nil {
					t.Fatal(err)
				}
				if tt.initial == domain.StatusCompleted {
					if err := order.Complete(); err != nil {
						t.Fatal(err)
					}
				}
			case domain.StatusCancelled:
				if err := order.Cancel(); err != nil {
					t.Fatal(err)
				}
			}
			id, customerID, items := order.ID(), order.CustomerID(), order.Items()
			err := tt.action(order)
			if tt.invalid {
				if !errors.Is(err, domain.ErrInvalidOrderState) {
					t.Fatalf("expected INVALID_ORDER_STATE, got %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if order.Status() != tt.want {
				t.Fatalf("status = %q, want %q", order.Status(), tt.want)
			}
			if order.ID() != id || order.CustomerID() != customerID || !reflect.DeepEqual(order.Items(), items) {
				t.Fatal("transition changed order identity, customer, or items")
			}
		})
	}
}

func TestZeroOrderRejectsActions(t *testing.T) {
	for _, action := range []struct {
		name string
		run  func(*domain.Order) error
	}{
		{"confirm", (*domain.Order).Confirm},
		{"complete", (*domain.Order).Complete},
		{"cancel", (*domain.Order).Cancel},
	} {
		t.Run(action.name, func(t *testing.T) {
			var order domain.Order
			if err := action.run(&order); !errors.Is(err, domain.ErrInvalidOrderState) {
				t.Fatalf("expected INVALID_ORDER_STATE, got %v", err)
			}
			if order.Status() != "" {
				t.Fatal("action changed zero order")
			}
		})
	}
}
