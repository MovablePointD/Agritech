package model

import "time"

// Message 私信消息
type Message struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	ConversationID uint      `json:"conversation_id" gorm:"index;not null"`
	SenderID       uint      `json:"sender_id" gorm:"not null"`
	Content        string    `json:"content" gorm:"type:text;not null"`
	IsRead         int       `json:"is_read" gorm:"default:0"` // 0-未读 1-已读
	CreatedAt      time.Time `json:"created_at"`
}
