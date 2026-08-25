package model

import "time"

// Notification 通知消息
type Notification struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id"`                                    // 接收通知的用户
	Type         string    `json:"type"`                                       // 类型
	Title        string    `json:"title"`                                      // 通知标题
	Content      string    `json:"content"`                                    // 通知内容
	RelatedID    uint      `json:"related_id"`                                 // 关联ID（如商品ID、订单ID）
	RelatedType  string    `json:"related_type"`                               // 关联类型: product/order/post/knowledge/comment
	SourceUserID uint      `json:"source_user_id"`                             // 触发通知的用户ID（用于去重）
	UniqueKey    string    `json:"unique_key" gorm:"type:varchar(255);uniqueIndex"` // 唯一标识，防止重复通知
	IsRead       int       `json:"is_read" gorm:"default:0"`                   // 是否已读: 0-未读 1-已读
	CreatedAt    time.Time `json:"created_at"`
}

// NotificationTypeText 通知类型文本
var NotificationTypeText = map[string]string{
	"stock_low":              "库存不足提醒",
	"order":                  "新订单通知",
	"system":                 "系统通知",
	"comment_post":           "动态评论通知",
	"comment_knowledge":      "知识评论通知",
	"reply_comment":          "回复通知",
	"at_comment":             "@通知",
	"new_message":            "新私信",
	"affair_approved":        "事务审核通知",
	"affair_accepted":        "事务被接取",
	"affair_processed":       "处理报告已提交",
	"affair_completed":       "事务已完成",
	"affair_follow_up":       "事务被追问",
	"affair_appeal":          "收到新申诉",
}

// NotificationType 用于通知类型的常量定义
const (
	NotificationTypeStockLow         = "stock_low"
	NotificationTypeOrder            = "order"
	NotificationTypeSystem           = "system"
	NotificationTypeCommentPost      = "comment_post"
	NotificationTypeCommentKnowledge = "comment_knowledge"
	NotificationTypeReplyComment     = "reply_comment"
	NotificationTypeAtComment        = "at_comment"
	NotificationTypeNewMessage       = "new_message"
	NotificationTypeAffairApproved   = "affair_approved"
	NotificationTypeAffairAccepted   = "affair_accepted"
	NotificationTypeAffairProcessed  = "affair_processed"
	NotificationTypeAffairCompleted  = "affair_completed"
	NotificationTypeAffairFollowUp   = "affair_follow_up"
	NotificationTypeAffairAppeal     = "affair_appeal"
)
