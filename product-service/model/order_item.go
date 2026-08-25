package model

import "time"

type OrderItem struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	OrderID     uint      `json:"order_id"`     // 订单ID
	ProductID   uint      `json:"product_id"`   // 商品ID
	ProductName string    `json:"product_name"` // 商品名称
	Price       float64   `json:"price"`        // 单价
	Quantity    int       `json:"quantity"`     // 数量
	Subtotal    float64   `json:"subtotal"`     // 小计
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// func (OrderItem) TableName() string {
// 	return "order_items"
// }
