package model

import "time"

// CommentRuralInfo 农村信息评论（支持农村信息介绍和政策公告两种目标）
type CommentRuralInfo struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	UserID        uint      `json:"user_id"`                         // 评论者
	TargetType    string    `json:"target_type" gorm:"size:20"`      // "rural_info" 或 "policy_notice"
	TargetID      uint      `json:"target_id"`                       // 目标ID
	Content       string    `json:"content"`                         // 评论内容
	ParentID      uint      `json:"parent_id"`                       // 0=一级评论
	Level         int       `json:"level"`                           // 层级：1-3
	ReplyToUserID uint      `json:"reply_to_user_id"`                // 被回复者ID
	Likes         int       `json:"likes" gorm:"default:0"`          // 点赞数
	CreatedAt     time.Time `json:"created_at"`
	User          *User     `json:"user,omitempty" gorm:"foreignKey:UserID;-:constraint"`
}
