package model

// Knowledge 知识模型（消息服务 stub，只包含评论通知所需字段）
type Knowledge struct {
	ID     uint   `json:"id" gorm:"primaryKey"`
	UserID uint   `json:"user_id"`
	Title  string `json:"title"`
}
