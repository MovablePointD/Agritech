// Package service 提供通用点赞/取消点赞服务
//
// ToggleLike: 原子切换（已赞→取消，未赞→点赞），返回最新状态
// 评论点赞使用 GREATEST(likes-1, 0) 防止减到负数
// GetUserLikes: 批量查询一批目标的点赞状态，用于列表渲染
package service

import (
	"strconv"

	"rxtcloud/common/config"
	"rxtcloud/common/model"

	"gorm.io/gorm/clause"
)

// ToggleLikeResult 点赞切换结果
type ToggleLikeResult struct {
	Liked bool `json:"liked"`
	Likes int  `json:"likes"`
}

// ToggleLike 点赞/取消点赞原子切换：已存在则删除记录并减数，不存在则创建记录并加数
func ToggleLike(userID uint, key string, targetID uint) (*ToggleLikeResult, error) {
	var existingLike model.Like
	err := config.DB.Where("user_id = ? AND target_type = ? AND target_id = ?",
		userID, "comment_rural_info", targetID).First(&existingLike).Error

	if err == nil {
		// 已点赞 → 取消点赞
		config.DB.Delete(&existingLike)
		_ = key
		newLikes, decErr := decrementCommentLikes(targetID)
		if decErr != nil {
			return nil, decErr
		}
		return &ToggleLikeResult{Liked: false, Likes: newLikes}, nil
	}

	// 未点赞 → 点赞
	like := model.Like{
		UserID:     userID,
		TargetType: "comment_rural_info",
		TargetID:   targetID,
	}
	if err := config.DB.Create(&like).Error; err != nil {
		return nil, err
	}
	newLikes, incErr := incrementCommentLikes(targetID)
	if incErr != nil {
		return nil, incErr
	}
	return &ToggleLikeResult{Liked: true, Likes: newLikes}, nil
}

// ToggleLikeCommentRuralInfo 农村评论点赞切换
func ToggleLikeCommentRuralInfo(id uint, userID uint) (*ToggleLikeResult, error) {
	key := "comment_rural_info:" + strconv.FormatUint(uint64(id), 10)
	return ToggleLike(userID, key, id)
}

func incrementCommentLikes(targetID uint) (int, error) {
	if err := config.DB.Model(&model.CommentRuralInfo{}).Where("id = ?", targetID).
		Update("likes", clause.Expr{SQL: "likes + 1"}).Error; err != nil {
		return 0, err
	}
	var c model.CommentRuralInfo
	config.DB.First(&c, targetID)
	return c.Likes, nil
}

func decrementCommentLikes(targetID uint) (int, error) {
	if err := config.DB.Model(&model.CommentRuralInfo{}).Where("id = ?", targetID).
		Update("likes", clause.Expr{SQL: "GREATEST(likes - 1, 0)"}).Error; err != nil {
		return 0, err
	}
	var c model.CommentRuralInfo
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
