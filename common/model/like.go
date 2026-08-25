package model

import "time"

// Like 用户点赞记录表（用于去重）
type Like struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"index:idx_user_target;not null"`          // 点赞用户
	TargetType string    `json:"target_type" gorm:"index:idx_user_target;size:30;not null"` // post / post_comment / knowledge / knowledge_comment / product_comment
	TargetID   uint      `json:"target_id" gorm:"index:idx_user_target;not null"`          // 目标ID
	CreatedAt  time.Time `json:"created_at"`
}
