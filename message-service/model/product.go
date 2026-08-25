package model

// Product 商品模型（消息服务 stub，只包含库存通知所需字段）
type Product struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Title     string `json:"title"`
	Stock     int    `json:"stock"`
	Publisher uint   `json:"publisher"`
}
