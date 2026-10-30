package memory_test

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/CatMacales/OrderFlow/internal/order/application"
	"github.com/CatMacales/OrderFlow/internal/order/domain"
	"github.com/CatMacales/OrderFlow/internal/order/storage"
	"github.com/CatMacales/OrderFlow/internal/order/storage/memory"
	"github.com/google/uuid"
)

var _ application.Repository = (*memory.Repository)(nil)

func newOrder(t *testing.T, id uuid.UUID, customer string) *domain.Order {
	t.Helper()
	item, err := domain.NewItem(" SKU-Mug ", 2)
	if err != nil {
		t.Fatal(err)
	}
	order, err := domain.NewOrder(id, customer, []domain.Item{item})
	if err != nil {
		t.Fatal(err)
	}
	return order
}

//
// Get
//

func TestCreateGetAndIsolation(t *testing.T) {
	ctx := context.Background()
	repository := memory.NewRepository()
	original := newOrder(t, uuid.New(), " Customer-42 ")
	if err := repository.Create(ctx, original); err != nil {
		t.Fatal(err)
	}
	got, err := repository.Get(ctx, original.ID())
	if err != nil {
		t.Fatal(err)
	}
	if got == original || !reflect.DeepEqual(got, original) {
		t.Fatalf("unexpected stored order: %+v", got)
	}
	if err := original.Confirm(); err != nil {
		t.Fatal(err)
	}
	if err := got.Cancel(); err != nil {
		t.Fatal(err)
	}
	items := got.Items()
	items[0] = domain.Item{}
	stored, err := repository.Get(ctx, original.ID())
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status() != domain.StatusCreated || stored.CustomerID() != " Customer-42 " || stored.Items()[0].ProductID() != " SKU-Mug " || stored.Items()[0].Quantity() != 2 {
		t.Fatalf("stored order was mutated: %+v", stored)
	}
}

func TestCreateDuplicate(t *testing.T) {
	ctx := context.Background()
	repository := memory.NewRepository()
	id := uuid.New()
	original := newOrder(t, id, "original")
	if err := repository.Create(ctx, original); err != nil {
		t.Fatal(err)
	}
	replacement := newOrder(t, id, "replacement")
	if err := replacement.Confirm(); err != nil {
		t.Fatal(err)
	}
	if err := repository.Create(ctx, replacement); !errors.Is(err, storage.ErrOrderAlreadyExists) {
		t.Fatalf("error: %v", err)
	}
	got, err := repository.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, original) {
		t.Fatalf("duplicate overwrote order: %+v", got)
	}
}

func TestGetMissing(t *testing.T) {
	repository := memory.NewRepository()
	order, err := repository.Get(context.Background(), uuid.New())
	if order != nil || !errors.Is(err, storage.ErrOrderNotFound) {
		t.Fatalf("order=%v err=%v", order, err)
	}
}

//
// Create
//

func TestCreateNil(t *testing.T) {
	repository := memory.NewRepository()
	if err := repository.Create(context.Background(), nil); err == nil {
		t.Fatal("expected an error for nil order")
	}
	order := newOrder(t, uuid.New(), "c")
	if err := repository.Create(context.Background(), order); err != nil {
		t.Fatal(err)
	}
}

func TestConcurrentCreateAndGet(t *testing.T) {
	ctx := context.Background()
	repository := memory.NewRepository()
	const count = 32
	orders := make([]*domain.Order, count)
	for i := range orders {
		orders[i] = newOrder(t, uuid.New(), "c")
	}
	var wg sync.WaitGroup
	start := make(chan struct{})
	for _, order := range orders {
		wg.Go(func() {
			<-start
			if err := repository.Create(ctx, order); err != nil {
				t.Error(err)
				return
			}
			for range 8 {
				got, err := repository.Get(ctx, order.ID())
				if err != nil {
					t.Error(err)
					return
				}
				if !reflect.DeepEqual(got, order) {
					t.Error("retrieved order differs from stored order")
					return
				}
			}
		})
	}
	close(start)
	wg.Wait()
}

func TestConcurrentDuplicateCreate(t *testing.T) {
	repository := memory.NewRepository()
	order := newOrder(t, uuid.New(), "c")
	var successes atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for range 32 {
		wg.Go(func() {
			<-start
			err := repository.Create(context.Background(), order)
			if err == nil {
				successes.Add(1)
			} else if !errors.Is(err, storage.ErrOrderAlreadyExists) {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
	close(start)
	wg.Wait()
	if successes.Load() != 1 {
		t.Fatalf("successful creates: %d, want 1", successes.Load())
	}
}

func TestServiceCreateGet(t *testing.T) {
	repository := memory.NewRepository()
	service := application.NewService(repository, uuid.New)
	input := application.CreateInput{CustomerID: " customer ", Items: []application.CreateItemInput{{ProductID: "p", Quantity: 2}, {ProductID: "P", Quantity: 1}}}
	id, err := service.Create(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if id.Version() != 4 || id.Variant() != uuid.RFC4122 {
		t.Fatalf("invalid generated ID: %v", id)
	}
	order, err := service.Get(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if order.ID() != id || order.CustomerID() != input.CustomerID || order.Status() != domain.StatusCreated {
		t.Fatalf("unexpected order: %+v", order)
	}
	items := order.Items()
	if len(items) != len(input.Items) {
		t.Fatalf("items: %+v", items)
	}
	for i, item := range items {
		if item.ProductID() != input.Items[i].ProductID || item.Quantity() != input.Items[i].Quantity {
			t.Fatalf("item %d: %+v", i, item)
		}
	}
	_, err = service.Get(context.Background(), uuid.New())
	if !errors.Is(err, storage.ErrOrderNotFound) {
		t.Fatalf("missing order error: %v", err)
	}
}

//
// Confirm
//

func TestConfirm(t *testing.T) {
	for _, tc := range []struct {
		name    string
		prepare func(*domain.Order) error
		status  domain.Status
		err     error
	}{
		{"created", nil, domain.StatusConfirmed, nil},
		{"confirmed", (*domain.Order).Confirm, domain.StatusConfirmed, nil},
		{"completed", func(order *domain.Order) error {
			if err := order.Confirm(); err != nil {
				return err
			}
			return order.Complete()
		}, domain.StatusCompleted, nil},
		{"cancelled", (*domain.Order).Cancel, domain.StatusCancelled, domain.ErrInvalidOrderState},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			repository := memory.NewRepository()
			order := newOrder(t, uuid.New(), "customer")
			if tc.prepare != nil {
				if err := tc.prepare(order); err != nil {
					t.Fatal(err)
				}
			}
			if err := repository.Create(ctx, order); err != nil {
				t.Fatal(err)
			}
			before, err := repository.Get(ctx, order.ID())
			if err != nil {
				t.Fatal(err)
			}
			if err := repository.Confirm(ctx, order.ID()); !errors.Is(err, tc.err) {
				t.Fatalf("error=%v, want %v", err, tc.err)
			}
			stored, err := repository.Get(ctx, order.ID())
			if err != nil {
				t.Fatal(err)
			}
			if stored.Status() != tc.status {
				t.Fatalf("status=%v, want %v", stored.Status(), tc.status)
			}
			if stored.ID() != order.ID() || stored.CustomerID() != order.CustomerID() || !reflect.DeepEqual(stored.Items(), order.Items()) {
				t.Fatalf("confirmation changed order data: %+v", stored)
			}
			if !reflect.DeepEqual(before, order) {
				t.Fatal("confirmation mutated a previously returned order")
			}
		})
	}
}

func TestConfirmMissing(t *testing.T) {
	repository := memory.NewRepository()
	id := uuid.New()
	if err := repository.Confirm(t.Context(), id); !errors.Is(err, storage.ErrOrderNotFound) {
		t.Fatalf("error=%v, want %v", err, storage.ErrOrderNotFound)
	}
	if order, err := repository.Get(t.Context(), id); order != nil || !errors.Is(err, storage.ErrOrderNotFound) {
		t.Fatalf("missing order was created: order=%v error=%v", order, err)
	}
}

func TestConcurrentConfirmAndGet(t *testing.T) {
	ctx := t.Context()
	repository := memory.NewRepository()
	order := newOrder(t, uuid.New(), "c")
	if err := repository.Create(ctx, order); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	start := make(chan struct{})
	for range 32 {
		wg.Go(func() {
			<-start
			if err := repository.Confirm(ctx, order.ID()); err != nil {
				t.Error(err)
				return
			}
			stored, err := repository.Get(ctx, order.ID())
			if err != nil {
				t.Error(err)
				return
			}
			if stored.Status() != domain.StatusConfirmed {
				t.Errorf("status=%v, want confirmed", stored.Status())
			}
		})
	}
	close(start)
	wg.Wait()
}

//
// Complete
//

func TestComplete(t *testing.T) {
	ctx := t.Context()
	for _, tc := range []struct {
		name    string
		prepare func(*domain.Order) error
		status  domain.Status
		err     error
	}{
		{"success", (*domain.Order).Confirm, domain.StatusCompleted, nil},
		{"invalid state", nil, domain.StatusCreated, domain.ErrInvalidOrderState},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repository := memory.NewRepository()
			order := newOrder(t, uuid.New(), "customer")
			if tc.prepare != nil {
				if err := tc.prepare(order); err != nil {
					t.Fatal(err)
				}
			}
			if err := repository.Create(ctx, order); err != nil {
				t.Fatal(err)
			}
			before, err := repository.Get(ctx, order.ID())
			if err != nil {
				t.Fatal(err)
			}
			if err := repository.Complete(ctx, order.ID()); !errors.Is(err, tc.err) {
				t.Fatalf("error=%v, want %v", err, tc.err)
			}
			stored, err := repository.Get(ctx, order.ID())
			if err != nil {
				t.Fatal(err)
			}
			if stored.Status() != tc.status {
				t.Fatalf("status=%v, want %v", stored.Status(), tc.status)
			}
			if stored.ID() != order.ID() || stored.CustomerID() != order.CustomerID() || !reflect.DeepEqual(stored.Items(), order.Items()) {
				t.Fatalf("complete changed order data: %+v", stored)
			}
			if !reflect.DeepEqual(before, order) {
				t.Fatal("complete mutated a previously returned order")
			}
		})
	}
}

func TestCompleteMissing(t *testing.T) {
	repository := memory.NewRepository()
	id := uuid.New()
	if err := repository.Complete(t.Context(), id); !errors.Is(err, storage.ErrOrderNotFound) {
		t.Fatalf("error=%v, want %v", err, storage.ErrOrderNotFound)
	}
	if order, err := repository.Get(t.Context(), id); order != nil || !errors.Is(err, storage.ErrOrderNotFound) {
		t.Fatalf("missing order was created: order=%v error=%v", order, err)
	}
}

//
// Cancel
//

func TestCancel(t *testing.T) {
	for _, tc := range []struct {
		name    string
		prepare func(*domain.Order) error
		status  domain.Status
		err     error
	}{
		{"success", nil, domain.StatusCancelled, nil},
		{"invalid state", func(order *domain.Order) error {
			if err := order.Confirm(); err != nil {
				return err
			}
			return order.Complete()
		}, domain.StatusCompleted, domain.ErrInvalidOrderState},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			repository := memory.NewRepository()
			order := newOrder(t, uuid.New(), "customer")
			if tc.prepare != nil {
				if err := tc.prepare(order); err != nil {
					t.Fatal(err)
				}
			}
			if err := repository.Create(ctx, order); err != nil {
				t.Fatal(err)
			}
			before, err := repository.Get(ctx, order.ID())
			if err != nil {
				t.Fatal(err)
			}
			if err := repository.Cancel(ctx, order.ID()); !errors.Is(err, tc.err) {
				t.Fatalf("error=%v, want %v", err, tc.err)
			}
			stored, err := repository.Get(ctx, order.ID())
			if err != nil {
				t.Fatal(err)
			}
			if stored.Status() != tc.status {
				t.Fatalf("status=%v, want %v", stored.Status(), tc.status)
			}
			if stored.ID() != order.ID() || stored.CustomerID() != order.CustomerID() || !reflect.DeepEqual(stored.Items(), order.Items()) {
				t.Fatalf("cancel changed order data: %+v", stored)
			}
			if !reflect.DeepEqual(before, order) {
				t.Fatal("cancel mutated a previously returned order")
			}
		})
	}
}

func TestCancelMissing(t *testing.T) {
	repository := memory.NewRepository()
	id := uuid.New()
	if err := repository.Cancel(t.Context(), id); !errors.Is(err, storage.ErrOrderNotFound) {
		t.Fatalf("error=%v, want %v", err, storage.ErrOrderNotFound)
	}
	if order, err := repository.Get(t.Context(), id); order != nil || !errors.Is(err, storage.ErrOrderNotFound) {
		t.Fatalf("missing order was created: order=%v error=%v", order, err)
	}
}
