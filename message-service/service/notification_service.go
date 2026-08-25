// Package service 提供系统通知服务
//
// 通知类型：
//   - comment_post / comment_knowledge - 评论通知
//   - reply_comment - 回复通知
//   - at_comment - @提及通知
//   - stock_low - 库存不足通知
//   - new_message - 新私信通知
//
// 去重机制：通过 uniqueKey（如 "comment_目标ID_用户ID"）防止重复通知
package service

import (
	"rxtcloud/message-service/config"
	"rxtcloud/message-service/model"
	"strings"
)

// CreateNotification 创建通知（防重复）
func CreateNotification(userID uint, notifType, title, content string, relatedID uint, relatedType string) error {
	notification := model.Notification{
		UserID:      userID,
		Type:        notifType,
		Title:       title,
		Content:     content,
		RelatedID:   relatedID,
		RelatedType: relatedType,
		IsRead:      0,
	}
	return config.DB.Create(&notification).Error
}

// CreateCommentNotification 创建带去重功能的通知：通过 uniqueKey 防止同一事件重复通知
// 例如同一用户在同一个动态下的评论只通知一次
func CreateCommentNotification(userID uint, notifType, title, content string, relatedID uint, relatedType string, sourceUserID uint, uniqueKey string) error {
	if uniqueKey != "" {
		var existing model.Notification
		err := config.DB.Where("unique_key = ?", uniqueKey).First(&existing).Error
		if err == nil {
			// 已经存在相同的通知，不重复创建
			return nil
		}
	}

	notification := model.Notification{
		UserID:       userID,
		Type:         notifType,
		Title:        title,
		Content:      content,
		RelatedID:    relatedID,
		RelatedType:  relatedType,
		SourceUserID: sourceUserID,
		UniqueKey:    uniqueKey,
		IsRead:       0,
	}
	return config.DB.Create(&notification).Error
}

func GetUserNotifications(userID uint, page, pageSize int) ([]model.Notification, int64, error) {
	var notifications []model.Notification
	var total int64

	query := config.DB.Model(&model.Notification{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&notifications).Error

	return notifications, total, err
}

func GetUnreadNotificationCount(userID uint) (int64, error) {
	var count int64
	err := config.DB.Model(&model.Notification{}).Where("user_id = ? AND is_read = ?", userID, 0).Count(&count).Error
	return count, err
}

func MarkNotificationAsRead(id uint, userID uint) error {
	return config.DB.Model(&model.Notification{}).Where("id = ? AND user_id = ?", id, userID).Update("is_read", 1).Error
}

func MarkAllNotificationsAsRead(userID uint) error {
	return config.DB.Model(&model.Notification{}).Where("user_id = ? AND is_read = ?", userID, 0).Update("is_read", 1).Error
}

// CheckAndNotifyLowStock 库存预警：库存<20 发提醒，库存=0 发告警
func CheckAndNotifyLowStock(product *model.Product) {
	if product.Stock > 0 && product.Stock < 20 {
		// 库存低于20但仍有库存，发送提醒
		title := "商品库存不足提醒"
		content := "您的商品【" + product.Title + "】库存不足，当前库存：" +
			itoa(uint(product.Stock)) + "，建议及时补充库存。"
		CreateNotification(product.Publisher, model.NotificationTypeStockLow, title, content, product.ID, "product")
	} else if product.Stock == 0 {
		// 库存为0
		title := "商品库存耗尽提醒"
		content := "您的商品【" + product.Title + "】库存已耗尽，建议及时补充库存或下架商品。"
		CreateNotification(product.Publisher, model.NotificationTypeStockLow, title, content, product.ID, "product")
	}
}

// NotifyPostComment 通知：有人在动态下发表评论
func NotifyPostComment(post *model.Post, commenter *model.User, commentContent string) {
	if post.UserID == commenter.ID {
		// 自己评论自己的动态，不通知
		return
	}
	title := "收到新评论"
	content := "用户【" + commenter.Nickname + "】在您的动态【" + post.Title + "】下发表评论：" + truncateString(commentContent, 50)
	uniqueKey := "post_comment_" + itoa(post.ID) + "_" + itoa(commenter.ID)
	CreateCommentNotification(post.UserID, model.NotificationTypeCommentPost, title, content, post.ID, "post", commenter.ID, uniqueKey)
}

// NotifyKnowledgeComment 通知：有人在知识下发表评论
func NotifyKnowledgeComment(knowledge *model.Knowledge, commenter *model.User, commentContent string) {
	if knowledge.UserID == commenter.ID {
		// 自己评论自己的知识，不通知
		return
	}
	title := "收到新评论"
	content := "用户【" + commenter.Nickname + "】在您的知识【" + knowledge.Title + "】下发表评论：" + truncateString(commentContent, 50)
	uniqueKey := "knowledge_comment_" + itoa(knowledge.ID) + "_" + itoa(commenter.ID)
	CreateCommentNotification(knowledge.UserID, model.NotificationTypeCommentKnowledge, title, content, knowledge.ID, "knowledge", commenter.ID, uniqueKey)
}

// NotifyReplyComment 通知：有人回复了您的评论
func NotifyReplyComment(replyToUser *model.User, replier *model.User, originalCommentContent string, relatedID uint, relatedType string) {
	if replyToUser.ID == replier.ID {
		// 自己回复自己，不通知
		return
	}
	title := "收到回复"
	content := "用户【" + replier.Nickname + "】回复了您的评论：" + truncateString(originalCommentContent, 50)
	uniqueKey := "reply_" + itoa(replier.ID) + "_" + itoa(replyToUser.ID) + "_" + itoa(relatedID)
	CreateCommentNotification(replyToUser.ID, model.NotificationTypeReplyComment, title, content, relatedID, relatedType, replier.ID, uniqueKey)
}

// NotifyAtUser 通知：有人在评论中@了您
func NotifyAtUser(atUser *model.User, mentioner *model.User, commentContent string, relatedID uint, relatedType string) {
	if atUser.ID == mentioner.ID {
		// 自己@自己，不通知
		return
	}
	title := "有人@了您"
	content := "用户【" + mentioner.Nickname + "】在评论中提到了您：" + truncateString(commentContent, 50)
	uniqueKey := "at_" + itoa(mentioner.ID) + "_" + itoa(atUser.ID) + "_" + itoa(relatedID)
	CreateCommentNotification(atUser.ID, model.NotificationTypeAtComment, title, content, relatedID, relatedType, mentioner.ID, uniqueKey)
}

// CheckAndNotifyAt 解析评论文本中的 @用户名，向被@用户发送通知
func CheckAndNotifyAt(commentContent string, commenter *model.User, relatedID uint, relatedType string) {
	// 解析@用户名（格式：@用户名）
	atPattern := "@"
	if strings.Contains(commentContent, atPattern) {
		// 查找被@的用户名
		parts := strings.Split(commentContent, "@")
		for i := 1; i < len(parts); i++ {
			username := strings.Fields(parts[i])[0] // 获取@后面的第一个词作为用户名
			if username != "" {
				// 查找用户
				var atUser model.User
				err := config.DB.Where("username = ? OR nickname = ?", username, username).First(&atUser).Error
				if err == nil && atUser.ID != commenter.ID {
					NotifyAtUser(&atUser, commenter, commentContent, relatedID, relatedType)
				}
			}
		}
	}
}

// GetUserByIDForNotify 根据ID获取用户（通知服务专用）
func GetUserByIDForNotify(userID uint) (*model.User, error) {
	var user model.User
	err := config.DB.First(&user, userID).Error
	return &user, err
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// itoa 简单的整数转字符串
func itoa(n uint) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+uint(n%10))) + result
		n /= 10
	}
	return result
}
