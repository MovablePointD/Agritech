package model

import "time"

type Order struct {
	ID           uint        `json:"id" gorm:"primaryKey"`
	OrderNo      string      `json:"order_no" gorm:"unique;not null"` // 订单编号
	UserID       uint        `json:"user_id"`                        // 用户ID
	OrderType    string      `json:"order_type"`                     // 订单类型: buy-购买 sell-出售
	TotalPrice   float64     `json:"total_price"`                   // 总价格
	Receiver     string      `json:"receiver"`                      // 收货人
	Phone        string      `json:"phone"`                         // 收货电话
	Address      string      `json:"address"`                        // 收货地址
	Status       int         `json:"status"`                        // 订单状态: 0-待付款 1-待发货 2-待收货 3-已完成 4-已取消
	PayStatus    int         `json:"pay_status"`                     // 支付状态: 0-未支付 1-已支付
	Remark       string      `json:"remark"`                        // 备注
	ExpressNo    string      `json:"express_no"`                    // 物流单号
	Items        []OrderItem `json:"items" gorm:"foreignKey:OrderID"` // 订单商品列表
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}
