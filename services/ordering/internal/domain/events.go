package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrderEvent interface {
	EventType() string
	OccurredAt() time.Time
}

type OrderCreated struct {
	OrderID    uuid.UUID
	CustomerID string
	Items      []OrderItem
	CreatedAt  time.Time
}

func (e OrderCreated) EventType() string {
	return "OrderCreated"
}

func (e OrderCreated) OccurredAt() time.Time {
	return e.CreatedAt
}

type OrderReserved struct {
	OrderID    uuid.UUID
	ReservedAt time.Time
}

func (e OrderReserved) EventType() string {
	return "OrderReserved"
}

func (e OrderReserved) OccurredAt() time.Time {
	return e.ReservedAt
}

type StockReservationFailed struct {
	OrderID  uuid.UUID
	Reason   string
	FailedAt time.Time
}

func (e StockReservationFailed) EventType() string {
	return "StockReservationFailed"
}

func (e StockReservationFailed) OccurredAt() time.Time {
	return e.FailedAt
}

type OrderConfirmed struct {
	OrderID     uuid.UUID
	ConfirmedAt time.Time
}

func (e OrderConfirmed) EventType() string {
	return "OrderConfirmed"
}

func (e OrderConfirmed) OccurredAt() time.Time {
	return e.ConfirmedAt
}

type OrderCancelled struct {
	OrderID     uuid.UUID
	Reason      string
	CancelledAt time.Time
}

func (e OrderCancelled) EventType() string {
	return "OrderCancelled"
}

func (e OrderCancelled) OccurredAt() time.Time {
	return e.CancelledAt
}
