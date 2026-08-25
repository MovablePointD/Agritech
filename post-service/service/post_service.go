// Package service 提供动态(Post)核心业务逻辑
//
// 状态说明：1-已发布  -1-待审核  0-已退回  2-已删除  3-草稿
//
// 热门推荐算法：
//
//	HotScore = (likes * 3 + views * 1) / (TIMESTAMPDIFF(DAY, created_at, NOW()) + 2)
//	综合点赞数(权重3)、浏览量(权重1)和发布时间衰减(+2防止零除)
//
// 评论树：CreateCommentPost 自动计算 level(最多3级) 和 ReplyToUserID
package service

import (
	"rxtcloud/common/client"
	"rxtcloud/post-service/config"
	"rxtcloud/post-service/model"

	"gorm.io/gorm/clause"
)

// 创建动态
func CreatePost(post *model.Post) error {
	// 默认待审核状态
	if post.Status == 0 {
		post.Status = -1
	}
	return config.DB.Create(post).Error
}

// 获取动态列表（分页，返回已发布的+当前用户自己的所有状态）
func GetPosts(page, pageSize int, postType, keyword string, userID uint) ([]model.Post, int64, error) {
	var posts []model.Post
	var total int64

	// status=1（已发布）+ 作者本人的待审核(-1)/退回(0)
	query := config.DB.Model(&model.Post{}).Preload("User")
	if userID > 0 {
		query = query.Where("status = 1 OR (status IN (-1, 0) AND user_id = ?)", userID)
	} else {
		query = query.Where("status = 1")
	}
	if postType != "" {
		query = query.Where("type = ?", postType)
	}
	if keyword != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&posts).Error

	return posts, total, err
}

// 获取用户动态
func GetPostsByUserID(userID uint) ([]model.Post, error) {
	var posts []model.Post
	err := config.DB.Preload("User").
		Where("user_id = ? AND status = 1", userID).
		Order("created_at desc").
		Find(&posts).Error
	return posts, err
}

// 删除动态（软删除）
func DeletePost(id uint) error {
	return config.DB.Model(&model.Post{}).Where("id = ?", id).Update("status", 2).Error
}

// 获取全部已发布动态（游客可浏览）
func GetPublishedPosts(page, pageSize int) ([]model.Post, int64, error) {
	var posts []model.Post
	var total int64

	query := config.DB.Model(&model.Post{}).Where("status = 1").Preload("User")
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&posts).Error

	return posts, total, err
}

// 获取热门动态（综合浏览量、点赞数与发布时间的热度算法）
// 热度 = (likes * 3 + views * 1) / (TIMESTAMPDIFF(DAY, created_at, NOW()) + 2)
func GetHotPosts(page, pageSize int) ([]model.Post, int64, error) {
	var posts []model.Post
	var total int64

	hotScore := "(likes * 3 + views * 1) / (TIMESTAMPDIFF(DAY, created_at, NOW()) + 2)"

	query := config.DB.Model(&model.Post{}).Where("status = 1")
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("User").
		Order(hotScore + " desc, created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&posts).Error

	return posts, total, err
}

// 获取待审核动态列表（管理员）
func GetPendingPostsForAdmin() ([]model.Post, error) {
	var posts []model.Post
	err := config.DB.Preload("User").
		Where("status = -1").
		Order("created_at desc").
		Find(&posts).Error
	return posts, err
}

// 审核动态
func AuditPost(id uint, approved bool) error {
	status := 1 // 通过
	if !approved {
		status = 0 // 隐藏
	}
	return config.DB.Model(&model.Post{}).Where("id = ?", id).Update("status", status).Error
}

// 恢复动态（重新提交审核）
func RestorePost(id uint) error {
	return config.DB.Model(&model.Post{}).Where("id = ?", id).Update("status", -1).Error
}

// 根据ID获取动态
func GetPostByID(id uint) (*model.Post, error) {
	var post model.Post
	err := config.DB.Preload("User").First(&post, id).Error
	return &post, err
}

// 更新动态
func UpdatePost(post *model.Post) error {
	// 先查询原记录，保留不可修改的字段
	var existing model.Post
	if err := config.DB.First(&existing, post.ID).Error; err != nil {
		return err
	}

	// 只更新允许修改的字段
	updates := map[string]interface{}{
		"title":   post.Title,
		"content": post.Content,
		"images":  post.Images,
	}

	// 保留原始字段
	updates["user_id"] = existing.UserID
	updates["type"] = existing.Type
	updates["created_at"] = existing.CreatedAt
	updates["likes"] = existing.Likes
	updates["views"] = existing.Views
	updates["status"] = existing.Status

	return config.DB.Model(&model.Post{}).Where("id = ?", post.ID).Updates(updates).Error
}

// 点赞动态
func LikePost(id uint) error {
	return config.DB.Model(&model.Post{}).Where("id = ?", id).
		Update("likes", clause.Expr{SQL: "likes + 1"}).Error
}

// 浏览+1
func IncPostViews(id uint) error {
	return config.DB.Model(&model.Post{}).Where("id = ?", id).
		Update("views", clause.Expr{SQL: "views + 1"}).Error
}

// 创建动态评论（自动计算 level 和 ReplyToUserID）
func CreateCommentPost(comment *model.CommentPost) error {
	// 一级评论
	if comment.ParentID == 0 {
		comment.Level = 1
		comment.ReplyToUserID = 0
	} else {
		// 查询父评论
		var parent model.CommentPost
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
	go sendPostCommentNotifications(comment)

	return nil
}

// sendPostCommentNotifications 发送动态评论相关通知（通过 HTTP 调用 message-service）
func sendPostCommentNotifications(comment *model.CommentPost) {
	// 获取动态信息
	post, err := GetPostByID(comment.PostID)
	if err != nil {
		return
	}

	notifClient := client.NewServiceClient("post-service")

	if comment.ParentID == 0 {
		// 一级评论：通知动态作者
		notifClient.CallAsync("message-service", "POST", "/api/internal/notify/comment", map[string]interface{}{
			"owner_user_id":   post.UserID,
			"commenter_id":    comment.UserID,
			"target_title":    post.Title,
			"comment_content": comment.Content,
			"related_id":      post.ID,
			"related_type":    "post",
			"unique_key":      "post_comment_" + itoaUint(post.ID) + "_" + itoaUint(comment.UserID),
		})
	} else {
		// 回复评论：通知被回复者
		notifClient.CallAsync("message-service", "POST", "/api/internal/notify/reply", map[string]interface{}{
			"reply_to_user_id": comment.ReplyToUserID,
			"replier_id":       comment.UserID,
			"comment_content":  comment.Content,
			"related_id":       post.ID,
			"related_type":     "post",
		})
	}
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

// 获取动态评论
func GetCommentPosts(postID uint) ([]model.CommentPost, error) {
	var comments []model.CommentPost
	err := config.DB.Preload("User").
		Where("post_id = ?", postID).
		Order("created_at asc").
		Find(&comments).Error
	return comments, err
}

// 更新动态评论
func UpdateCommentPost(comment *model.CommentPost) error {
	// 先查询原记录，保留不可修改的字段
	var existing model.CommentPost
	if err := config.DB.First(&existing, comment.ID).Error; err != nil {
		return err
	}

	// 只更新允许修改的字段
	updates := map[string]interface{}{
		"content": comment.Content,
	}

	// 保留原始字段
	updates["user_id"] = existing.UserID
	updates["post_id"] = existing.PostID
	updates["parent_id"] = existing.ParentID
	updates["level"] = existing.Level
	updates["reply_to_user_id"] = existing.ReplyToUserID
	updates["created_at"] = existing.CreatedAt
	updates["likes"] = existing.Likes

	return config.DB.Model(&model.CommentPost{}).Where("id = ?", comment.ID).Updates(updates).Error
}

// 删除动态评论
func DeleteCommentPost(id uint) error {
	return config.DB.Delete(&model.CommentPost{}, id).Error
}

// 点赞动态评论
func LikeCommentPost(id uint) error {
	return config.DB.Model(&model.CommentPost{}).Where("id = ?", id).
		Update("likes", clause.Expr{SQL: "likes + 1"}).Error
}

// 获取所有动态（管理员，包含已删除）
func GetAllPostsForAdmin(page, pageSize int, status int) ([]model.Post, int64, error) {
	var posts []model.Post
	var total int64

	query := config.DB.Model(&model.Post{})
	if status != -999 { // -999 表示获取所有状态
		query = query.Where("status = ?", status)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := config.DB.Preload("User").
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&posts).Error

	return posts, total, err
}

// 获取已退回动态（审核不通过）
func GetRejectedPostsForAdmin() ([]model.Post, error) {
	var posts []model.Post
	err := config.DB.Preload("User").
		Where("status = 0"). // 0 = 隐藏/退回
		Order("created_at desc").
		Find(&posts).Error
	return posts, err
}

// 获取已删除动态
func GetDeletedPostsForAdmin() ([]model.Post, error) {
	var posts []model.Post
	err := config.DB.Preload("User").
		Where("status = 2"). // 2 = 删除
		Order("created_at desc").
		Find(&posts).Error
	return posts, err
}

// ========== 草稿相关 ==========

// 创建草稿
func CreateDraftPost(post *model.Post) error {
	post.Status = 3
	post.Likes = 0
	post.Views = 0
	return config.DB.Create(post).Error
}

// 更新草稿（仅允许本人更新 status=3 的记录）
func UpdateDraftPost(post *model.Post) error {
	updates := map[string]interface{}{
		"title":   post.Title,
		"content": post.Content,
		"images":  post.Images,
	}
	if post.Type != "" {
		updates["type"] = post.Type
	}
	return config.DB.Model(&model.Post{}).
		Where("id = ? AND status = 3 AND user_id = ?", post.ID, post.UserID).
		Updates(updates).Error
}

// 获取用户的草稿列表
func GetDraftPosts(userID uint) ([]model.Post, error) {
	var posts []model.Post
	err := config.DB.Preload("User").
		Where("user_id = ? AND status = 3", userID).
		Order("updated_at desc").
		Find(&posts).Error
	return posts, err
}

// 删除草稿（物理删除，因为是草稿没必要软删除）
func DeleteDraftPost(id uint, userID uint) error {
	return config.DB.Where("id = ? AND status = 3 AND user_id = ?", id, userID).
		Delete(&model.Post{}).Error
}

// 获取草稿详情（仅本人可查看）
func GetDraftPostByID(id uint, userID uint) (*model.Post, error) {
	var post model.Post
	err := config.DB.Preload("User").
		Where("id = ? AND status = 3 AND user_id = ?", id, userID).
		First(&post).Error
	return &post, err
}

// 发布草稿（status: 3 → 1 或 -1）
func PublishPostFromDraft(id uint, userID uint, newStatus int) error {
	return config.DB.Model(&model.Post{}).
		Where("id = ? AND status = 3 AND user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"status":     newStatus,
			"updated_at": nil, // 触发 GORM 自动更新
		}).Error
}
