package application

import (
	"errors"
	"testing"

	"github.com/CatMacales/OrderFlow/internal/order/application/mock"
	"github.com/CatMacales/OrderFlow/internal/order/domain"
	"github.com/CatMacales/OrderFlow/internal/order/storage"
	"github.com/google/uuid"
)

func TestConfirm(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{
		{"success", nil},
		{"missing", storage.ErrOrderNotFound},
		{"invalid state", domain.ErrInvalidOrderState},
		{"repository error", errors.New("unavailable")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			id := uuid.New()
			repository := mock.NewRepositoryMock(t)
			repository.ConfirmMock.Expect(ctx, id).Return(tc.err)
			service := NewService(repository, nil)
			if err := service.Confirm(ctx, id); !errors.Is(err, tc.err) {
				t.Fatalf("error=%v, want %v", err, tc.err)
			}
			if repository.ConfirmAfterCounter() != 1 {
				t.Fatal("expected one repository call")
			}
		})
	}
}

func TestConfirmInvalidID(t *testing.T) {
	wrongVariant := uuid.New()
	wrongVariant[8] = 0
	for _, tc := range []struct {
		name string
		id   uuid.UUID
	}{
		{"nil", uuid.Nil},
		{"wrong version", uuid.MustParse("550e8400-e29b-11d4-a716-446655440000")},
		{"wrong variant", wrongVariant},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repository := mock.NewRepositoryMock(t)
			service := NewService(repository, nil)
			if err := service.Confirm(t.Context(), tc.id); !errors.Is(err, domain.ErrInvalidArgument) {
				t.Fatalf("error=%v, want %v", err, domain.ErrInvalidArgument)
			}
			if repository.ConfirmAfterCounter() != 0 {
				t.Fatal("invalid ID was sent to repository")
			}
		})
	}
}
