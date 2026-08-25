package model

import "time"

// RuralAffair 农村事务
//
// 完整生命周期状态流转:
//
//	用户提出[1] → (用户修改[1]) → 审核通过[2] → 人员接取[4] → 首次处理回应
//	├─ 用户追问 → [6]追问中 → 人员追答 → [4]处理中（用户可确认完成或继续追问）
//	└─ 用户确认完成 → [5]已完成
//	     └─ 用户申诉 → [7]申诉中 → admin/sysadmin处理 → [5]已完成
//	人员申诉（状态=4且有process_content时） → [7]申诉中 → admin/sysadmin处理 → [4]处理中
//	审核不通过 → [3]
type RuralAffair struct {
	ID              uint       `json:"id" gorm:"primaryKey"`
	Title           string     `json:"title" gorm:"size:200;not null"`       // 事务标题
	Content         string     `json:"content" gorm:"type:text"`             // 事务描述
	Type            string     `json:"type" gorm:"size:50"`                  // 类型: infrastructure-基础设施 hardware-硬件 env-环境 safety-安全 other-其他
	Address         string     `json:"address" gorm:"size:255"`              // 事发地点
	Images          string     `json:"images" gorm:"type:text"`               // 图片JSON数组
	VideoURL        string     `json:"video_url" gorm:"size:500"`            // 视频URL
	UserID          uint       `json:"user_id"`                             // 提交用户
	Status          int        `json:"status" gorm:"default:1"`             // 1-待审核 2-审核通过 3-审核不通过 4-处理中 5-已完成 6-追问中 7-申诉中
	RejectReason    string     `json:"reject_reason" gorm:"type:text"`       // 审核不通过原因
	HandlerID       uint       `json:"handler_id"`                          // 处理人员ID
	HandlerName     string     `json:"handler_name" gorm:"size:50"`          // 处理人员姓名
	ProcessContent  string     `json:"process_content" gorm:"type:text"`    // 处理详情
	ProcessImages   string     `json:"process_images" gorm:"type:text"`      // 处理图片
	ProcessTime     *time.Time `json:"process_time"`                        // 接取时间(开始处理)
	FirstResponseAt *time.Time `json:"first_response_at"`                   // 首次提交处理记录时间(用于判断可申诉条件)
	CompletedTime   *time.Time `json:"completed_time"`                      // 完成时间
	LastModifiedAt  *time.Time `json:"last_modified_at"`                    // 用户最后一次修改时间(仅前端展示用)
	AuditTime       *time.Time `json:"audit_time"`                          // 审核时间
	AuditName       string     `json:"audit_name" gorm:"size:50"`            // 审核人员姓名
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	User            *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// AffairStatusText 状态文本映射
var AffairStatusText = map[int]string{
	1: "待审核",
	2: "审核通过",
	3: "审核不通过",
	4: "处理中",
	5: "已完成",
	6: "追问中",
	7: "申诉中",
}

// AffairTypeText 事务类型文本映射
var AffairTypeText = map[string]string{
	"infrastructure": "基础设施",
	"hardware": "硬件问题",
	"env": "环境问题",
	"safety": "安全隐患",
	"other": "其他",
}

// AffairStatusClass 状态样式映射 (对应 Element Plus tag type)
var AffairStatusClass = map[int]string{
	1: "warning",   // 待审核 - 黄色
	2: "primary",   // 审核通过 - 蓝色
	3: "danger",    // 审核不通过 - 红色
	4: "info",      // 处理中 - 灰色
	5: "success",   // 已完成 - 绿色
	6: "warning",   // 追问中 - 黄色(需关注)
	7: "danger",    // 申诉中 - 红色(需处理)
}

// AffairAppealStatusText 申诉状态文本
var AffairAppealStatusText = map[string]string{
	"pending":    "待处理",
	"processing": "处理中",
	"resolved":   "已处理",
}
