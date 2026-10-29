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
		wg.Add(1)
		go func() {
			defer wg.Done()
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
		}()
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
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			err := repository.Create(context.Background(), order)
			if err == nil {
				successes.Add(1)
			} else if !errors.Is(err, storage.ErrOrderAlreadyExists) {
				t.Errorf("unexpected error: %v", err)
			}
		}()
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
