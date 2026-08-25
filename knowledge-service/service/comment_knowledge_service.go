package service

import (
	"rxtcloud/common/client"
	"rxtcloud/knowledge-service/config"
	"rxtcloud/knowledge-service/model"

	"gorm.io/gorm/clause"
)

// 创建知识评论（自动计算 level 和 ReplyToUserID）
func CreateCommentKnowledge(comment *model.CommentKnowledge) error {
	// 一级评论
	if comment.ParentID == 0 {
		comment.Level = 1
		comment.ReplyToUserID = 0
	} else {
		// 查询父评论
		var parent model.CommentKnowledge
		if err := config.DB.First(&parent, comment.ParentID).Error; err != nil {
			return err
		}
		// 根据父评论的 level 计算当前评论的 level
		if parent.Level >= 3 {
			// 三级评论的回复依然是三级
			comment.Level = 3
		} else {
			comment.Level = parent.Level + 1
		}
		// 回复目标用户ID
		comment.ReplyToUserID = parent.UserID
	}

	// 创建评论
	if err := config.DB.Create(comment).Error; err != nil {
		return err
	}

	// 发送通知
	go sendKnowledgeCommentNotifications(comment)

	return nil
}

// sendKnowledgeCommentNotifications 发送知识评论相关通知（通过 HTTP 调用 message-service）
func sendKnowledgeCommentNotifications(comment *model.CommentKnowledge) {
	// 获取评论者昵称
	commenter, err := client.GetUserInfo(comment.UserID, "knowledge-service")
	if err != nil {
		return
	}

	// 获取知识信息
	knowledge, err := GetKnowledgeByIDForComment(comment.KnowledgeID)
	if err != nil {
		return
	}

	notifClient := client.NewServiceClient("knowledge-service")

	if comment.ParentID == 0 {
		// 一级评论：通知知识作者
		notifClient.CallAsync("message-service", "POST", "/api/internal/notify/comment", map[string]interface{}{
			"owner_user_id":   knowledge.UserID,
			"commenter_name":  commenter.Nickname,
			"target_title":    knowledge.Title,
			"comment_content": comment.Content,
			"related_id":      knowledge.ID,
			"related_type":    "knowledge",
			"unique_key":      "knowledge_comment_" + itoaUint(knowledge.ID) + "_" + itoaUint(comment.UserID),
		})
	} else {
		// 回复评论：通知被回复者
		notifClient.CallAsync("message-service", "POST", "/api/internal/notify/reply", map[string]interface{}{
			"reply_to_user_id": comment.ReplyToUserID,
			"replier_name":     commenter.Nickname,
			"comment_content":  comment.Content,
			"related_id":       knowledge.ID,
			"related_type":     "knowledge",
		})
	}

	// 检查@通知（暂未实现 HTTP 转发）
}

func itoaUint(n uint) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for tmp := n; tmp > 0; tmp /= 10 {
		digits = append([]byte{byte('0' + tmp%10)}, digits...)
	}
	return string(digits)
}

// GetKnowledgeByIDForComment 根据ID获取知识（评论服务专用）
func GetKnowledgeByIDForComment(id uint) (*model.Knowledge, error) {
	var knowledge model.Knowledge
	err := config.DB.Preload("User").First(&knowledge, id).Error
	return &knowledge, err
}

// 获取某知识下所有评论
func GetCommentKnowledgeByKnowledgeID(knowledgeID uint) ([]model.CommentKnowledge, error) {
	var comments []model.CommentKnowledge
	err := config.DB.Preload("User").
		Where("knowledge_id = ?", knowledgeID).
		Order("created_at asc").
		Find(&comments).Error
	return comments, err
}

// 更新知识评论
func UpdateCommentKnowledge(comment *model.CommentKnowledge) error {
	// 先查询原记录，保留不可修改的字段
	var existing model.CommentKnowledge
	if err := config.DB.First(&existing, comment.ID).Error; err != nil {
		return err
	}

	// 只更新允许修改的字段
	updates := map[string]interface{}{
		"content": comment.Content,
	}

	// 保留原始字段
	updates["user_id"] = existing.UserID
	updates["knowledge_id"] = existing.KnowledgeID
	updates["parent_id"] = existing.ParentID
	updates["level"] = existing.Level
	updates["reply_to_user_id"] = existing.ReplyToUserID
	updates["created_at"] = existing.CreatedAt
	updates["likes"] = existing.Likes

	return config.DB.Model(&model.CommentKnowledge{}).Where("id = ?", comment.ID).Updates(updates).Error
}

// 删除知识评论
func DeleteCommentKnowledge(id uint) error {
	return config.DB.Delete(&model.CommentKnowledge{}, id).Error
}

// 点赞评论
func LikeCommentKnowledge(id uint) error {
	return config.DB.Model(&model.CommentKnowledge{}).Where("id = ?", id).
		Update("likes", clause.Expr{SQL: "likes + 1"}).Error
}
