package model

import "time"

type Post struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`   // 发布者
	Title     string    `json:"title"`     // 动态标题
	Content   string    `json:"content"`   // 动态正文
	Images    string    `json:"images"`    // 图片JSON数组 ["url1","url2"]
	VideoURL  string    `json:"video_url"` // 视频URL
	Likes     int       `json:"likes"`     // 点赞数
	Views     int       `json:"views"`     // 浏览数
	Type      string    `json:"type"`      // 类型: normal-普通 question-提问 share-分享
	Status    int       `json:"status"`    // 状态: -1-待审核 0-隐藏 1-正常 2-删除 3-草稿
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
