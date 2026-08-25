package model

import "time"

// AffairAppeal 事务申诉记录
// 用户可以对已完成(status=5)的事务提出申诉
// 事务处理人员可以对自己已接取且已提交首次处理记录(status=4,已有process_content)的事务提出申诉
// 申诉由审查人员(admin)或系统管理员(sysadmin)处理
// 申诉结果对所有参与者公开
type AffairAppeal struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	AffairID      uint       `json:"affair_id" gorm:"index;not null"`  // 关联事务ID
	// 申诉人类型: "user"=用户, "handler"=事务处理人员
	ApplicantType string     `json:"applicant_type" gorm:"size:20;not null"`
	ApplicantID   uint       `json:"applicant_id" gorm:"not null"`      // 申诉人ID
	ApplicantName string     `json:"applicant_name" gorm:"size:50"`     // 申诉人名称
	Reason        string     `json:"reason" gorm:"type:text;not null"`  // 申诉理由
	Images        string     `json:"images" gorm:"type:text"`           // 申诉附图
	// 状态: "pending"=待处理, "processing"=处理中, "resolved"=已处理
	Status       string     `json:"status" gorm:"size:20;default:'pending'"`
	Result       string     `json:"result" gorm:"type:text"`     // 申诉处理结果
	HandlerID    uint       `json:"handler_id"`                   // 申诉处理人ID（admin/sysadmin）
	HandlerName  string     `json:"handler_name" gorm:"size:50"`  // 申诉处理人姓名
	ProcessedAt  *time.Time `json:"processed_at"`                 // 处理完成时间
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (AffairAppeal) TableName() string {
	return "affair_appeals"
}
