// model/comment_product.go
package model

import "time"

type CommentProduct struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	UserID        uint      `json:"user_id"`
	ProductID     uint      `json:"product_id"`
	Content       string    `json:"content"`
	ParentID      uint      `json:"parent_id"`      // 0 = 一级评论（楼中楼）
	Level         int       `json:"level"`          // 层级：1=一级，2=二级，3=三级
	ReplyToUserID uint      `json:"reply_to_user_id"` // 被回复的用户ID（仅二级及以上），无外键约束
	Likes         int       `json:"likes"`          // 点赞数
	CreatedAt     time.Time `json:"created_at"`
	User          *User     `json:"user,omitempty" gorm:"foreignKey:UserID;-:constraint"` // 评论者信息
}
