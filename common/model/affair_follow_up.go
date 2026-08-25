package model

import "time"

// AffairFollowUp 事务追问/追答记录
// 用户在事务处理人员发布回应后、确认完成前可以提出追问
// 处理人员对追问进行追答
type AffairFollowUp struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	AffairID  uint      `json:"affair_id" gorm:"index;not null"`  // 关联事务ID
	// 类型: "question"=用户追问, "answer"=处理人员追答
	Type       string    `json:"type" gorm:"size:20;not null"`
	UserID     uint      `json:"user_id" gorm:"not null"`          // 提问/回答者ID
	UserName   string    `json:"user_name" gorm:"size:50"`         // 提问/回答者名称
	Content    string    `json:"content" gorm:"type:text;not null"` // 追问/追答内容
	Images     string    `json:"images" gorm:"type:text"`          // 追问/追答附图
	CreatedAt  time.Time `json:"created_at"`
}

func (AffairFollowUp) TableName() string {
	return "affair_follow_ups"
}
