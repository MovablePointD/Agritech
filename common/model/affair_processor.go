package model

import "time"

// AffairProcessor 事务处理人员申请表
type AffairProcessor struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id" gorm:"not null"`       // 申请人ID
	RealName     string    `json:"real_name" gorm:"size:50"`       // 真实姓名
	Phone        string    `json:"phone" gorm:"size:20"`           // 联系电话
	Address      string    `json:"address" gorm:"size:255"`       // 负责区域/地址
	Intro        string    `json:"intro" gorm:"type:text"`        // 个人简介/申请理由
	CertImages   string    `json:"cert_images" gorm:"type:text"`  // 证明材料图片
	Status       int       `json:"status" gorm:"default:0"`       // 状态: 0-待审核 1-已通过 2-已拒绝
	RejectReason string    `json:"reject_reason" gorm:"type:text"` // 拒绝原因
	AuditTime    *time.Time `json:"audit_time"`                     // 审核时间
	AuditName    string    `json:"audit_name" gorm:"size:50"`      // 审核人姓名
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	User         *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// ProcessorStatusText 状态文本映射
var ProcessorStatusText = map[int]string{
	0: "待审核",
	1: "已通过",
	2: "已拒绝",
}

// ProcessorStatusClass 状态样式映射
var ProcessorStatusClass = map[int]string{
	0: "warning", // 待审核 - 黄色
	1: "success", // 已通过 - 绿色
	2: "danger",  // 已拒绝 - 红色
}
