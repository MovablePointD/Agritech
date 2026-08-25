package model

import "time"

type Expert struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"unique;not null"` // 关联用户ID
	User       User      `json:"user" gorm:"foreignKey:UserID"` // 关联用户
	RealName   string    `json:"real_name"`                     // 真实姓名
	Phone      string    `json:"phone"`                         // 联系电话
	Profession string    `json:"profession"`                    // 专业领域
	Title      string    `json:"title"`                        // 职位
	Company    string    `json:"company"`                      // 所属单位
	Intro      string    `json:"intro"`                         // 个人简介
	CertNo     string    `json:"cert_no"`                       // 证书编号（可验证）
	CertImage  string    `json:"cert_image"`                    // 证书图片
	Status     int       `json:"status"`                        // 审核状态: 0-待审核 1-已通过 2-未通过
	VerifyAt   *time.Time `json:"verify_at"`                   // 审核时间
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
