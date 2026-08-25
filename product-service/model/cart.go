package model

import "time"

type CartItem struct {
	ID         uint      `json:"id" gorm:"primaryKey"` // 购物车ID
	UserID     uint      `json:"user_id"`              // 用户ID
	ProductID  uint      `json:"product_id"`           // 商品ID
	Quantity   int       `json:"quantity"`             // 数量
	Price      float64   `json:"price"`                // 单价
	TotalPrice float64   `json:"total_price"`          // 总价
	CreatedAt  time.Time `json:"created_at"`           // 创建时间
	UpdatedAt  time.Time `json:"updated_at"`           // 修改时间

	// 关联商品信息
	Product *Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
}

// func (CartItem) TableName() string {
// 	return "cart_items"
// }
