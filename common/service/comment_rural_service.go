package service

import (
	"rxtcloud/common/config"
	"rxtcloud/common/model"
)

// ==================== CommentRuralInfo 服务 ====================

func CreateCommentRuralInfo(comment *model.CommentRuralInfo) error {
	return config.DB.Create(comment).Error
}

func GetCommentRuralInfos(targetType string, targetID uint) ([]model.CommentRuralInfo, error) {
	var comments []model.CommentRuralInfo
	err := config.DB.Preload("User").
		Where("target_type = ? AND target_id = ?", targetType, targetID).
		Order("created_at asc").
		Find(&comments).Error
	return comments, err
}

func UpdateCommentRuralInfo(id uint, userID uint, content string) error {
	result := config.DB.Model(&model.CommentRuralInfo{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("content", content)
	if result.RowsAffected == 0 {
		return config.DB.Error
	}
	return result.Error
}

func DeleteCommentRuralInfo(id uint, userID uint) error {
	result := config.DB.Where("id = ? AND user_id = ?", id, userID).
		Delete(&model.CommentRuralInfo{})
	if result.RowsAffected == 0 {
		return config.DB.Error
	}
	return result.Error
}
