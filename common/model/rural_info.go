package model

import "time"

// RuralInfo 农村基础信息介绍
type RuralInfo struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`                              // 发布者
	Title     string    `json:"title" gorm:"size:200;not null"`       // 信息标题
	Content   string    `json:"content" gorm:"type:text"`             // 正文内容（支持Markdown）
	Type      string    `json:"type" gorm:"size:50;default:overview"` // 类型
	Address   string    `json:"address" gorm:"size:255"`              // 所属地区
	Images    string    `json:"images" gorm:"type:text"`              // 图片JSON数组
	Views     int       `json:"views" gorm:"default:0"`               // 浏览数
	Status    int       `json:"status" gorm:"default:1"`              // -1待审核 0隐藏 1正常 2已删除
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// RuralInfoTypeText 类型文本映射
var RuralInfoTypeText = map[string]string{
	"overview":       "综合概况",
	"village_intro":  "村情介绍",
	"resource":       "农业资源",
	"culture":        "乡村文化",
	"transportation": "交通设施",
	"education":      "教育医疗",
	"other":          "其他信息",
}
