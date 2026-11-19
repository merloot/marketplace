package domain

import "github.com/google/uuid"

type OrderItem struct {
	ProductID   uuid.UUID
	ProductName string
	Quantity    int
	PriceAmount int64
}

func NewOrderItem(productID uuid.UUID, productName string, quantity int, priceAmount int64) OrderItem {
	return OrderItem{
		ProductID:   productID,
		ProductName: productName,
		Quantity:    quantity,
		PriceAmount: priceAmount,
	}
}
