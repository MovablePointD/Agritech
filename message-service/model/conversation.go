package model

import "time"

// Conversation 私信会话
type Conversation struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	User1ID          uint      `json:"user1_id" gorm:"index:idx_users;not null"`
	User2ID          uint      `json:"user2_id" gorm:"index:idx_users;not null"`
	ConversationType string    `json:"conversation_type" gorm:"type:varchar(20);default:'normal'"` // normal-普通私信, consultation-专家咨询
	ExpertName       string    `json:"expert_name" gorm:"type:varchar(100)"`                       // 咨询时的专家姓名（用于展示）
	LastMessage      string    `json:"last_message" gorm:"type:text"`
	LastMessageAt    time.Time `json:"last_message_at"`
	User1Unread      int       `json:"user1_unread" gorm:"default:0"` // user1 未读数
	User2Unread      int       `json:"user2_unread" gorm:"default:0"` // user2 未读数
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	// 非数据库字段，用于API响应
	OtherUser *UserInfo `json:"other_user" gorm:"-"`
}

// UserInfo 简化的用户信息（用于会话列表展示）
type UserInfo struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}
