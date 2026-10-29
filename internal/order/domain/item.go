package domain

import "fmt"

// Item is an immutable product line within an order.
type Item struct {
	productID string
	quantity  int64
}

func NewItem(productID string, quantity int64) (Item, error) {
	item := Item{productID: productID, quantity: quantity}
	if err := item.validate(); err != nil {
		return Item{}, err
	}
	return item, nil
}

func (i Item) ProductID() string {
	return i.productID
}

func (i Item) Quantity() int64 {
	return i.quantity
}

func (i Item) validate() error {
	if i.productID == "" {
		return fmt.Errorf("product_id must not be empty: %w", ErrInvalidArgument)
	}
	if i.quantity <= 0 {
		return fmt.Errorf("quantity must be positive: %w", ErrInvalidArgument)
	}
	return nil
}
