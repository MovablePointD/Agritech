package model

import "time"

type SensitiveWord struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Word      string    `json:"word" gorm:"size:100;uniqueIndex;not null"` // 敏感词
	Level     int       `json:"level" gorm:"default:1"`                  // 级别：1-轻度 2-中度 3-重度
	Category  string    `json:"category" gorm:"size:50"`                 // 类别：politics-政治 porn-色情 advertising-广告 gamble-赌博 fraud-诈骗 other-其他
	CreatedAt time.Time `json:"created_at"`
}

var SensitiveWordLevelText = map[int]string{
	1: "轻度",
	2: "中度",
	3: "重度",
}

var SensitiveWordCategoryText = map[string]string{
	"politics":    "政治敏感",
	"porn":       "色情低俗",
	"advertising": "广告推广",
	"gamble":     "赌博博彩",
	"fraud":      "欺诈诈骗",
	"other":      "其他违规",
}
