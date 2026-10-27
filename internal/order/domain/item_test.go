package domain_test

import (
	"errors"
	"math"
	"testing"

	"github.com/CatMacales/OrderFlow/internal/order/domain"
)

func TestNewItem(t *testing.T) {
	tests := []struct {
		name      string
		productID string
		quantity  int64
		valid     bool
	}{
		{"valid", "SKU-MUG", 2, true},
		{"external format", "catalog:notebook:7", 1, true},
		{"preserve case and whitespace", " SKU-Mug \t", 1, true},
		{"whitespace is nonempty", " ", 1, true},
		{"maximum quantity", "p", math.MaxInt64, true},
		{"empty product", "", 1, false},
		{"zero quantity", "p", 0, false},
		{"negative quantity", "p", -1, false},
		{"minimum quantity", "p", math.MinInt64, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := domain.NewItem(tt.productID, tt.quantity)
			if !tt.valid {
				if !errors.Is(err, domain.ErrInvalidArgument) {
					t.Fatalf("expected INVALID_ARGUMENT, got %v", err)
				}
				if item != (domain.Item{}) {
					t.Fatalf("failed construction returned an item: %+v", item)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if item.ProductID() != tt.productID || item.Quantity() != tt.quantity {
				t.Fatalf("item changed input: %+v", item)
			}
		})
	}
}
