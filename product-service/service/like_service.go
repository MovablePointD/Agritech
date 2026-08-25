package service

import (
	"errors"
	"rxtcloud/product-service/config"
	"rxtcloud/product-service/model"

	"gorm.io/gorm/clause"
)

// ToggleLikeResult 点赞切换结果
type ToggleLikeResult struct {
	Liked bool `json:"liked"`
	Likes int  `json:"likes"`
}

// ToggleLike 点赞/取消点赞切换（product-service 仅处理 product_comment）
func ToggleLike(userID uint, targetType string, targetID uint) (*ToggleLikeResult, error) {
	if targetType != "product_comment" {
		return nil, errors.New("product-service only supports product_comment, got: " + targetType)
	}

	var existingLike model.Like
	err := config.DB.Where("user_id = ? AND target_type = ? AND target_id = ?",
		userID, targetType, targetID).First(&existingLike).Error

	if err == nil {
		config.DB.Delete(&existingLike)
		newLikes, decErr := decrementLikes(targetType, targetID)
		if decErr != nil {
			return nil, decErr
		}
		return &ToggleLikeResult{Liked: false, Likes: newLikes}, nil
	}

	like := model.Like{
		UserID:     userID,
		TargetType: targetType,
		TargetID:   targetID,
	}
	if err := config.DB.Create(&like).Error; err != nil {
		return nil, err
	}
	newLikes, incErr := incrementLikes(targetType, targetID)
	if incErr != nil {
		return nil, incErr
	}
	return &ToggleLikeResult{Liked: true, Likes: newLikes}, nil
}

func incrementLikes(targetType string, targetID uint) (int, error) {
	if err := config.DB.Model(&model.CommentProduct{}).Where("id = ?", targetID).
		Update("likes", clause.Expr{SQL: "likes + 1"}).Error; err != nil {
		return 0, err
	}
	var c model.CommentProduct
	config.DB.First(&c, targetID)
	return c.Likes, nil
}

func decrementLikes(targetType string, targetID uint) (int, error) {
	if err := config.DB.Model(&model.CommentProduct{}).Where("id = ?", targetID).
		Update("likes", clause.Expr{SQL: "GREATEST(likes - 1, 0)"}).Error; err != nil {
		return 0, err
	}
	var c model.CommentProduct
	config.DB.First(&c, targetID)
	return c.Likes, nil
}

// GetUserLikes 获取用户对一批目标的点赞状态
func GetUserLikes(userID uint, targetType string, targetIDs []uint) map[uint]bool {
	result := make(map[uint]bool)
	var likes []model.Like
	config.DB.Where("user_id = ? AND target_type = ? AND target_id IN ?",
		userID, targetType, targetIDs).Find(&likes)
	for _, l := range likes {
		result[l.TargetID] = true
	}
	return result
}

// IsLiked 判断用户是否已点赞
func IsLiked(userID uint, targetType string, targetID uint) bool {
	var count int64
	config.DB.Model(&model.Like{}).Where(
		"user_id = ? AND target_type = ? AND target_id = ?",
		userID, targetType, targetID).Count(&count)
	return count > 0
}
