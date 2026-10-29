package domain

import (
	"fmt"
	"slices"

	"github.com/google/uuid"
)

type Order struct {
	id         uuid.UUID
	customerID string
	status     Status
	items      []Item
}

// NewOrder creates a placed order awaiting store confirmation.
// UUID generation belongs to the caller; only valid v4 IDs are accepted.
func NewOrder(id uuid.UUID, customerID string, items []Item) (*Order, error) {
	if err := ValidateOrderID(id); err != nil {
		return nil, err
	}
	if customerID == "" {
		return nil, fmt.Errorf("customer_id must not be empty: %w", ErrInvalidArgument)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("items must contain at least one line: %w", ErrInvalidArgument)
	}
	seen := make(map[string]struct{}, len(items))
	for i, item := range items {
		if err := item.validate(); err != nil {
			return nil, fmt.Errorf("items[%d]: %w", i, err)
		}
		if _, exists := seen[item.productID]; exists {
			return nil, fmt.Errorf("items[%d].product_id is duplicated: %w", i, ErrInvalidArgument)
		}
		seen[item.productID] = struct{}{}
	}

	return &Order{
		id:         id,
		customerID: customerID,
		status:     StatusCreated,
		items:      slices.Clone(items),
	}, nil
}

func (o *Order) ID() uuid.UUID {
	return o.id
}

func (o *Order) CustomerID() string {
	return o.customerID
}

func (o *Order) Status() Status {
	return o.status
}

func (o *Order) Items() []Item {
	return slices.Clone(o.items)
}

// Confirm accepts an order for processing. Confirmed and completed orders
// already satisfy confirmation, so retries leave their status unchanged.
func (o *Order) Confirm() error {
	switch o.status {
	case StatusCreated:
		o.status = StatusConfirmed
		return nil
	case StatusConfirmed, StatusCompleted:
		return nil
	default:
		return fmt.Errorf("%w: cannot confirm order in %q state", ErrInvalidOrderState, o.status)
	}
}

func (o *Order) Complete() error {
	switch o.status {
	case StatusConfirmed:
		o.status = StatusCompleted
		return nil
	case StatusCompleted:
		return nil
	default:
		return fmt.Errorf("cannot complete order in %q state: %w", o.status, ErrInvalidOrderState)
	}
}

func (o *Order) Cancel() error {
	switch o.status {
	case StatusCreated, StatusConfirmed:
		o.status = StatusCancelled
		return nil
	case StatusCancelled:
		return nil
	default:
		return fmt.Errorf("cannot cancel order in %q state: %w", o.status, ErrInvalidOrderState)
	}
}

func ValidateOrderID(id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("order_id must not be nil: %w", ErrInvalidArgument)
	}
	if id.Version() != 4 || id.Variant() != uuid.RFC4122 {
		return fmt.Errorf(
			"order_id must be a UUID v4 with RFC 4122 variant: %w",
			ErrInvalidArgument,
		)
	}
	return nil
}
