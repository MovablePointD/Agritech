package model

// User 用户模型（knowledge-service stub，只包含关联查询所需字段）
// 注意：users 表由 common-service 管理，此 stub 仅用于 GORM 关联查询，
// 字段标签需与 common/model/user.go 保持一致，避免 AutoMigrate 误修改表结构
type User struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Username  string `json:"username" gorm:"type:varchar(191);uniqueIndex;not null"`
	Nickname  string `json:"nickname" gorm:"type:varchar(191)"`
	AvatarURL string `json:"avatar_url" gorm:"type:varchar(500)"`
}
