package service

import (
	"rxtcloud/common/config"
	"rxtcloud/common/model"
)

// CreateCommentNotification 创建评论相关通知（防重复）
func CreateCommentNotification(userID uint, notifType, title, content string, relatedID uint, relatedType string, sourceUserID uint, uniqueKey string) error {
	// 检查是否已存在相同的通知（防重复）
	if uniqueKey != "" {
		var existing model.Notification
		err := config.DB.Where("unique_key = ?", uniqueKey).First(&existing).Error
		if err == nil {
			return nil // 已存在，不重复创建
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
