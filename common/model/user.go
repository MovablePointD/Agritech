package model

import "time"

type User struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	Username    string     `json:"username" gorm:"unique;not null"`
	Password    string     `json:"-"`        // 不返回
	Nickname    string     `json:"nickname"` // 昵称，默认等于username
	Phone       string     `json:"phone"`    // 手机号（可为空，可重复）
	Email       string     `json:"email"`    // 邮箱（可为空，可重复）
	AvatarURL   string     `json:"avatar_url"`
	Signature   string     `json:"signature"` // 个人签名
	Role        string     `json:"role"`      // normal / farmer / expert / processor / admin / sysadmin
	Status      int        `json:"status"`    // 1=正常 0=禁用
	Banned      bool       `json:"banned" gorm:"default:false"` // 是否被封禁
	BannedUntil *time.Time `json:"banned_until"`                // 封禁截止时间
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
