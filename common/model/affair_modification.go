package model

import "time"

// AffairModification 事务修改记录
// 用户可在事务状态=1(待审核)时多次修改事务详情
// 前端只展示最后一次修改时间
type AffairModification struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	AffairID  uint      `json:"affair_id" gorm:"index;not null"`  // 关联事务ID
	UserID    uint      `json:"user_id" gorm:"not null"`          // 修改用户ID
	UserName  string    `json:"user_name" gorm:"size:50"`         // 修改用户名称
	// 修改前后的字段快照，留作记录
	OldTitle   string `json:"old_title" gorm:"size:200"`
	NewTitle   string `json:"new_title" gorm:"size:200"`
	OldContent string `json:"old_content" gorm:"type:text"`
	NewContent string `json:"new_content" gorm:"type:text"`
	OldAddress string `json:"old_address" gorm:"size:255"`
	NewAddress string `json:"new_address" gorm:"size:255"`
	OldImages  string `json:"old_images" gorm:"type:text"`
	NewImages  string `json:"new_images" gorm:"type:text"`
	OldVideoURL string `json:"old_video_url" gorm:"size:500"`
	NewVideoURL string `json:"new_video_url" gorm:"size:500"`
	CreatedAt  time.Time `json:"created_at"`
}

func (AffairModification) TableName() string {
	return "affair_modifications"
}
