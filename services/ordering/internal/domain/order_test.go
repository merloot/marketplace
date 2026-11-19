package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/merloot/marketplace/services/ordering/internal/domain"
)

// Helper для создания тестового товара
func testItem() domain.OrderItem {
	return domain.OrderItem{
		ProductID:   uuid.New(),
		Quantity:    1,
		PriceAmount: 1000,
	}
}

func TestNewOrder(t *testing.T) {
	t.Run("Given valid items, When create, Then order is PENDING and has OrderCreated event", func(t *testing.T) {
		items := []domain.OrderItem{testItem()}
		customerID := "customer-123"

		order, err := domain.NewOrder(customerID, items)

		assert.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, domain.StatusPending, order.Status())
		assert.Equal(t, 1, order.Version())

		// Проверяем накопленные события
		events := order.UncommittedEvents()
		assert.Len(t, events, 1)
		assert.IsType(t, domain.OrderCreated{}, events[0])
	})

	t.Run("Given empty items, When create, Then error", func(t *testing.T) {
		_, err := domain.NewOrder("customer-123", []domain.OrderItem{})
		assert.Error(t, err)
	})
}

func TestOrder_MarkAsReserved(t *testing.T) {
	t.Run("Given PENDING order, When mark reserved, Then status RESERVED", func(t *testing.T) {
		order, _ := domain.NewOrder("cust", []domain.OrderItem{testItem()})
		order.ClearUncommittedEvents() // Очищаем, чтобы проверить только новые события

		err := order.MarkAsReserved()

		assert.NoError(t, err)
		assert.Equal(t, domain.StatusReserved, order.Status())
		assert.Equal(t, 2, order.Version()) // 1 (created) + 1 (reserved)

		events := order.UncommittedEvents()
		assert.Len(t, events, 1)
		assert.IsType(t, domain.OrderReserved{}, events[0])
	})

	t.Run("Given RESERVED order, When mark reserved again, Then error", func(t *testing.T) {
		order, _ := domain.NewOrder("cust", []domain.OrderItem{testItem()})
		order.MarkAsReserved()

		err := order.MarkAsReserved()
		assert.ErrorIs(t, err, domain.ErrInvalidState)
	})
}

func TestOrder_Confirm(t *testing.T) {
	t.Run("Given PENDING order, When confirm, Then status CONFIRMED", func(t *testing.T) {
		order, _ := domain.NewOrder("cust", []domain.OrderItem{testItem()})
		order.ClearUncommittedEvents()

		err := order.Confirm()

		assert.NoError(t, err)
		assert.Equal(t, domain.StatusConfirmed, order.Status())
		assert.Equal(t, 2, order.Version())

		events := order.UncommittedEvents()
		assert.Len(t, events, 1)
		assert.IsType(t, domain.OrderConfirmed{}, events[0])
	})

	t.Run("Given RESERVED order, When confirm, Then status CONFIRMED", func(t *testing.T) {
		order, _ := domain.NewOrder("cust", []domain.OrderItem{testItem()})
		order.MarkAsReserved()
		order.ClearUncommittedEvents()

		err := order.Confirm()

		assert.NoError(t, err)
		assert.Equal(t, domain.StatusConfirmed, order.Status())
	})

	t.Run("Given CONFIRMED order, When confirm again, Then error", func(t *testing.T) {
		order, _ := domain.NewOrder("cust", []domain.OrderItem{testItem()})
		order.Confirm()

		err := order.Confirm()
		assert.ErrorIs(t, err, domain.ErrInvalidState)
	})
}

func TestOrder_Cancel(t *testing.T) {
	t.Run("Given PENDING order, When cancel, Then status CANCELLED", func(t *testing.T) {
		order, _ := domain.NewOrder("cust", []domain.OrderItem{testItem()})
		order.ClearUncommittedEvents()

		err := order.Cancel()

		assert.NoError(t, err)
		assert.Equal(t, domain.StatusCancelled, order.Status())
		assert.Equal(t, 2, order.Version())

		events := order.UncommittedEvents()
		assert.Len(t, events, 1)
		assert.IsType(t, domain.OrderCancelled{}, events[0])
	})

	t.Run("Given CONFIRMED order, When cancel, Then error", func(t *testing.T) {
		order, _ := domain.NewOrder("cust", []domain.OrderItem{testItem()})
		order.Confirm()

		err := order.Cancel()
		assert.ErrorIs(t, err, domain.ErrInvalidState)
	})
}

func TestLoadFromEvents(t *testing.T) {
	orderID := uuid.New()
	customerID := "cust"
	items := []domain.OrderItem{testItem()}

	events := []domain.OrderEvent{
		domain.OrderCreated{
			OrderID:    orderID,
			CustomerID: customerID,
			Items:      items,
			CreatedAt:  time.Now(),
		},
		domain.OrderReserved{
			OrderID:    orderID,
			ReservedAt: time.Now(),
		},
	}

	order := domain.LoadFromEvents(events)

	assert.Equal(t, orderID, order.ID())
	assert.Equal(t, domain.StatusReserved, order.Status())
	assert.Equal(t, 2, order.Version())
	// uncommittedEvents должен быть пустым, потому что мы загрузили из истории
	assert.Empty(t, order.UncommittedEvents())
}

func TestOrder_ApplyDoesNotDuplicateEvents(t *testing.T) {
	// Этот тест проверяет, что apply не добавляет события в uncommitted.
	// Это критично: если при LoadFromEvents мы случайно добавим события в uncommitted,
	// то при следующем сохранении запишем дубликаты.

	orderID := uuid.New()
	events := []domain.OrderEvent{
		domain.OrderCreated{OrderID: orderID, CreatedAt: time.Now()},
	}

	order := domain.LoadFromEvents(events)

	// uncommitted должен быть пуст, несмотря на то что мы применили событие
	assert.Empty(t, order.UncommittedEvents())
}
