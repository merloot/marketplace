package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	StatusPending   OrderStatus = "PENDING"
	StatusReserved  OrderStatus = "RESERVED"
	StatusConfirmed OrderStatus = "CONFIRMED"
	StatusCancelled OrderStatus = "CANCELLED"
)

var (
	ErrOrderAlreadyExists = errors.New("order already exists")
	ErrOrderNotFound      = errors.New("order not found")
	ErrInvalidState       = errors.New("invalid state for this operation")
)

type Order struct {
	id                uuid.UUID
	status            OrderStatus
	items             []OrderItem
	version           int
	uncommittedEvents []OrderEvent
}

func NewOrder(customerID string, items []OrderItem) (*Order, error) {
	if len(items) == 0 {
		return nil, errors.New("order must have at least one item")
	}
	o := &Order{
		id: uuid.New(),
	}
	event := OrderCreated{
		OrderID:    o.ID(),
		CustomerID: customerID,
		Items:      items,
		CreatedAt:  time.Now(),
	}
	o.recordEvent(event)
	return o, nil
}

func (o *Order) ID() uuid.UUID                   { return o.id }
func (o *Order) UncommittedEvents() []OrderEvent { return o.uncommittedEvents }
func (o *Order) Version() int                    { return o.version }
func (o *Order) Status() OrderStatus             { return o.status }

func LoadFromEvents(events []OrderEvent) *Order {
	o := &Order{}
	for _, event := range events {
		o.apply(event)
	}
	return o
}

func (o *Order) Confirm() error {
	if o.status != StatusPending && o.status != StatusReserved {
		return ErrInvalidState
	}
	event := OrderConfirmed{
		OrderID:     o.ID(),
		ConfirmedAt: time.Now(),
	}
	o.recordEvent(event)
	return nil
}

func (o *Order) Cancel() error {
	if o.status == StatusConfirmed {
		return ErrInvalidState
	}
	event := OrderCancelled{
		OrderID:     o.ID(),
		CancelledAt: time.Now(),
	}
	o.recordEvent(event)
	return nil
}

func (o *Order) MarkAsReserved() error {
	if o.status != StatusPending {
		return ErrInvalidState
	}
	event := OrderReserved{
		OrderID:    o.ID(),
		ReservedAt: time.Now(),
	}
	o.recordEvent(event)
	return nil
}

func (o *Order) ClearUncommittedEvents() {
	o.uncommittedEvents = nil
}

func (o *Order) apply(event OrderEvent) {
	switch e := event.(type) {
	case OrderCreated:
		o.id = e.OrderID
		o.status = StatusPending
		o.items = e.Items
	case OrderReserved:
		o.status = StatusReserved
	case OrderConfirmed:
		o.status = StatusConfirmed
	case OrderCancelled:
		o.status = StatusCancelled
	}
	o.version++
}

func (o *Order) recordEvent(event OrderEvent) {
	o.apply(event)
	o.uncommittedEvents = append(o.uncommittedEvents, event)
}
