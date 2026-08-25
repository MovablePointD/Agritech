package model

// Address 收货地址（与 common/model/address.go 表结构一致，共享 MySQL）
type Address struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	UserID    uint   `json:"user_id"`
	Receiver  string `json:"receiver"`
	Phone     string `json:"phone"`
	Province  string `json:"province"`
	City      string `json:"city"`
	District  string `json:"district"`
	Address   string `json:"address"`
	IsDefault bool   `json:"is_default"`
}
