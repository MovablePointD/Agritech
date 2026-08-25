package service

import (
	"rxtcloud/common/client"
	"rxtcloud/message-service/config"
	"rxtcloud/message-service/model"
)

// ToggleLikeResult 点赞切换结果
type ToggleLikeResult struct {
	Liked bool `json:"liked"`
	Likes int  `json:"likes"`
}

// ToggleLike 点赞/取消点赞（写了）Like 记录到共享表，然后异步转发到目标服务更新计数
// 这是 message-service 作为点赞转发中枢的入口
func ToggleLike(userID uint, targetType string, targetID uint) (*ToggleLikeResult, error) {
	// 1. 写入 Like 记录表（共享 MySQL）
	var existingLike model.Like
	err := config.DB.Where("user_id = ? AND target_type = ? AND target_id = ?",
		userID, targetType, targetID).First(&existingLike).Error
	isLike := true

	if err == nil {
		// 已点赞 → 取消点赞
		config.DB.Delete(&existingLike)
		isLike = false
	} else {
		// 未点赞 → 点赞
		like := model.Like{
			UserID:     userID,
			TargetType: targetType,
			TargetID:   targetID,
		}
		if err := config.DB.Create(&like).Error; err != nil {
			return nil, err
		}
	}

	// 2. 异步转发到目标服务，更新目标的 likes 计数
	client.ForwardLikeAsync(client.LikeCallbackBody{
		UserID:     userID,
		TargetType: targetType,
		TargetID:   targetID,
		IsLike:     isLike,
	}, "message-service")

	// 3. 返回成功（实际 likes 计数由目标服务异步更新）
	return &ToggleLikeResult{Liked: isLike, Likes: -1}, nil
}

// GetUserLikes 获取用户对一批目标的点赞状态（本地 Like 表查询）
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
