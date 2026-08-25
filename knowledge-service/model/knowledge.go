package model

import "time"

type Knowledge struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UserID    uint      `json:"user_id"`
	ImageURL  string    `json:"image_url"`
	VideoURL  string    `json:"video_url"`                      // 视频URL
	Likes     int       `json:"likes"`                          // 点赞数
	Status    int       `json:"status" gorm:"default:-1"`      // 状态 -1-待审核 0-隐藏 1-已发布 2-已删除 3-草稿
	User      *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
