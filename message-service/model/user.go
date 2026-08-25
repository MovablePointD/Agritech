package model

// User 用户模型（消息服务 stub，只包含通知/消息所需字段）
type User struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}
