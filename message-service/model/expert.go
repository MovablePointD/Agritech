package model

// Expert 专家模型（消息服务 stub，只包含咨询所需字段）
type Expert struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	UserID   uint   `json:"user_id"`
	Status   int    `json:"status"` // 0-待审核 1-已认证
	RealName string `json:"real_name"`
}
