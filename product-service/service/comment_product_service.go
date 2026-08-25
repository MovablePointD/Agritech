// service/comment_product_service.go
package service

import (
	"rxtcloud/product-service/config"
	"rxtcloud/product-service/model"

	"gorm.io/gorm/clause"
)

// 创建商品评论（自动计算 level 和 ReplyToUserID）
func CreateCommentProduct(comment *model.CommentProduct) error {
	// 一级评论
	if comment.ParentID == 0 {
		comment.Level = 1
		comment.ReplyToUserID = 0
	} else {
		// 查询父评论
		var parent model.CommentProduct
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
	return config.DB.Create(comment).Error
}

// 获取某商品下所有评论
func GetCommentProductByProductID(productID uint) ([]model.CommentProduct, error) {
	var comments []model.CommentProduct
	err := config.DB.Preload("User").
		Where("product_id = ?", productID).
		Order("created_at asc").
		Find(&comments).Error
	return comments, err
}

// 更新商品评论
func UpdateCommentProduct(comment *model.CommentProduct) error {
	// 先查询原记录，保留不可修改的字段
	var existing model.CommentProduct
	if err := config.DB.First(&existing, comment.ID).Error; err != nil {
		return err
	}

	// 只更新允许修改的字段
	updates := map[string]interface{}{
		"content": comment.Content,
	}

	// 保留原始字段
	updates["user_id"] = existing.UserID
	updates["product_id"] = existing.ProductID
	updates["parent_id"] = existing.ParentID
	updates["level"] = existing.Level
	updates["reply_to_user_id"] = existing.ReplyToUserID
	updates["created_at"] = existing.CreatedAt
	updates["likes"] = existing.Likes

	return config.DB.Model(&model.CommentProduct{}).Where("id = ?", comment.ID).Updates(updates).Error
}

// 删除商品评论
func DeleteCommentProduct(id uint) error {
	return config.DB.Delete(&model.CommentProduct{}, id).Error
}

// 点赞商品评论
func LikeCommentProduct(id uint) error {
	return config.DB.Model(&model.CommentProduct{}).Where("id = ?", id).
		Update("likes", clause.Expr{SQL: "likes + 1"}).Error
}
