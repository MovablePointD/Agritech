package model

import "time"

type Address struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`                        // 用户ID
	Receiver  string    `json:"receiver"`                       // 收货人
	Phone     string    `json:"phone"`                         // 联系电话
	Province  string    `json:"province"`                      // 省
	City      string    `json:"city"`                          // 市
	District  string    `json:"district"`                      // 区/县
	Detail    string    `json:"detail"`                        // 详细地址
	IsDefault int       `json:"is_default"`                   // 是否默认: 0-否 1-是
	Type      string    `json:"type"`                          // 地址类型: home-家庭 company-公司 school-学校
	Label     string    `json:"label"`                         // 地址标签（如：家、公司）
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
