package model

import "time"

type Product struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Title        string    `json:"title"`      // 标题
	Price        float64   `json:"price"`      // 价格
	Content      string    `json:"content"`    // 内容
	Status       int       `json:"status"`     // 状态 -1-待审核 0-下架 1-上架
	Type         string    `json:"type"`       // 类型
	ImageURL     string    `json:"image_url"`  // 图片URL
	Publisher    uint      `json:"publisher"`   // 发布者ID
	PublisherUser *User    `json:"publisher_user,omitempty" gorm:"foreignKey:Publisher"`
	Partner      uint      `json:"partner"`     // 合伙人
	Address      string    `json:"address"`    // 地址
	Stock        int       `json:"stock"`      // 库存数量
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
