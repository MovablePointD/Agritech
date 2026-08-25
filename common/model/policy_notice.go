package model

import "time"

// PolicyNotice 农村政策公告
type PolicyNotice struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id"`                              // 发布者
	Title       string    `json:"title" gorm:"size:200;not null"`       // 公告标题
	Content     string    `json:"content" gorm:"type:text"`             // 正文内容（支持Markdown）
	Category    string    `json:"category" gorm:"size:50"`              // 政策分类
	Address     string    `json:"address" gorm:"size:255"`              // 关联农村地址（用于与农村信息、事务关联匹配，忽略大小写）
	PublishDept string    `json:"publish_dept" gorm:"size:100"`         // 发布部门
	PublishDate string    `json:"publish_date" gorm:"size:20"`          // 政策生效日期
	Images      string    `json:"images" gorm:"type:text"`              // 图片JSON数组
	Attachment  string    `json:"attachment" gorm:"size:500"`           // 附件URL
	Views       int       `json:"views" gorm:"default:0"`               // 浏览数
	IsTop       bool      `json:"is_top" gorm:"default:false"`          // 是否置顶
	Status      int       `json:"status" gorm:"default:1"`              // -1待审核 0隐藏 1正常 2已删除
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	User        *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// PolicyCategoryText 政策分类文本映射
var PolicyCategoryText = map[string]string{
	"subsidy":       "补贴政策",
	"land":          "土地政策",
	"environmental": "环保政策",
	"technology":    "科技政策",
	"comprehensive": "综合政策",
	"healthcare":    "医疗养老",
	"education":     "教育政策",
	"other":         "其他政策",
}
