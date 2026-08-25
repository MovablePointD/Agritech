package model

import "time"

// AffairAppealMaterial 申诉辅助材料
// 申诉期间，事务的双方参与者都可以提交辅助材料
// - 用户申诉时：处理人员可以提交辅助材料
// - 处理人员申诉时：用户可以提交辅助材料
type AffairAppealMaterial struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	AppealID  uint      `json:"appeal_id" gorm:"index;not null"`  // 关联申诉ID
	// 提交者类型: "user"=用户, "handler"=处理人员
	SubmitterType string `json:"submitter_type" gorm:"size:20;not null"`
	SubmitterID   uint   `json:"submitter_id" gorm:"not null"`
	SubmitterName string `json:"submitter_name" gorm:"size:50"`
	Content       string `json:"content" gorm:"type:text"`   // 辅助说明
	Images        string `json:"images" gorm:"type:text"`    // 辅助材料附图
	CreatedAt     time.Time `json:"created_at"`
}

func (AffairAppealMaterial) TableName() string {
	return "affair_appeal_materials"
}
