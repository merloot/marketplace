package events

import "github.com/google/uuid"

var (
	ProductCreatedEvent = "ProductCreated"
)

type ProductCreated struct {
	ProductID   uuid.UUID `json:"product_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       int64     `json:"price"`
	SKU         string    `json:"sku"`
}

type OrderCreated struct {
	OrderID    uuid.UUID `json:"order_id"`
	ProductID  uuid.UUID `json:"product_id"`
	Quantity   int       `json:"quantity"`
	TotalPrice int64     `json:"total_price"`
}

type StockReserved struct {
	OrderID   uuid.UUID `json:"order_id"`
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}
