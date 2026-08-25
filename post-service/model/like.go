package model

import "time"

// Like 点赞记录（与 common/model/like.go 字段一致，共享同一 MySQL 表）
type Like struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"index:idx_user_target;not null"`
	TargetType string    `json:"target_type" gorm:"index:idx_user_target;size:30;not null"`
	TargetID   uint      `json:"target_id" gorm:"index:idx_user_target;not null"`
	CreatedAt  time.Time `json:"created_at"`
}
