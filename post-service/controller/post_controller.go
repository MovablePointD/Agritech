package controller

import (
	"net/http"
	"strconv"

	"rxtcloud/post-service/model"
	"rxtcloud/post-service/service"

	"github.com/gin-gonic/gin"
)

// 创建动态
func CreatePost(c *gin.Context) {
	var post model.Post

	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	post.UserID = uint(userID.(uint))
	post.Likes = 0
	post.Views = 0

	// 检查敏感词
	hasImages := post.Images != ""
	hasSensitive := service.HasSensitiveWord(post.Title + " " + post.Content)
	autoPass, reason := service.ShouldAutoPass(hasImages, hasSensitive)

	if autoPass {
		post.Status = 1 // 自动通过
	} else {
		post.Status = -1 // 待审核	}

		if err := service.CreatePost(&post); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "发布失败"})
			return
		}

		if post.Status == 1 {
			c.JSON(http.StatusOK, gin.H{"msg": "发布成功", "data": post})
		} else {
			c.JSON(http.StatusOK, gin.H{"msg": "提交成功，内容需等待审核", "reason": reason, "data": post})
		}
	}
}

// 获取动态列表
func GetPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	postType := c.DefaultQuery("type", "")
	keyword := c.DefaultQuery("keyword", "")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	// 获取当前用户ID（可选，未登录也可查看列表）
	var userID uint
	if uid, exists := c.Get("user_id"); exists {
		userID = uint(uid.(uint))
	}

	list, total, err := service.GetPosts(page, pageSize, postType, keyword, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// 获取动态详情
func GetPost(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	post, err := service.GetPostByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "动态不存在"})
		return
	}

	// 获取当前用户ID
	var userID uint
	if uid, exists := c.Get("user_id"); exists {
		userID = uint(uid.(uint))
	}

	// 只有已发布(1)或作者本人才能查看
	if post.Status != 1 && post.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "动态不存在"})
		return
	}

	// 浏览+1
	service.IncPostViews(uint(id))

	// 判断当前用户是否已点赞
	isLiked := userID > 0 && service.IsLiked(userID, "post", uint(id))

	c.JSON(http.StatusOK, gin.H{
		"post":     post,
		"is_liked": isLiked,
	})
}

// 获取我的动态
func GetMyPosts(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	list, err := service.GetPostsByUserID(uint(userID.(uint)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"list": list})
}

// 更新动态
func UpdatePost(c *gin.Context) {
	var post model.Post

	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	userID, _ := c.Get("user_id")
	existing, _ := service.GetPostByID(post.ID)

	if existing == nil || existing.UserID != uint(userID.(uint)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作"})
		return
	}

	if err := service.UpdatePost(&post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "更新成功"})
}

// 删除动态
func DeletePost(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	post, _ := service.GetPostByID(uint(id))
	if post == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "动态不存在"})
		return
	}

	userID, _ := c.Get("user_id")
	if post.UserID != uint(userID.(uint)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}

	if err := service.DeletePost(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "删除成功"})
}

// 点赞/取消点赞动态
func LikePost(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	result, err := service.ToggleLike(userID.(uint), "post", uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "点赞失败"})
		return
	}

	msg := "点赞成功"
	if !result.Liked {
		msg = "已取消点赞"
	}
	c.JSON(http.StatusOK, gin.H{"msg": msg, "liked": result.Liked, "likes": result.Likes})
}

// ========== 动态评论 ==========

// 创建评论
func CreateCommentPost(c *gin.Context) {
	var comment model.CommentPost

	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	comment.UserID = uint(userID.(uint))

	if err := service.CreateCommentPost(&comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "评论失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "评论成功"})
}

// 获取动态评论
func GetCommentPosts(c *gin.Context) {
	println("获取动态评论")
	postID, _ := strconv.Atoi(c.Param("id"))

	list, err := service.GetCommentPosts(uint(postID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	tree := BuildCommentPostTree(list)
	c.JSON(http.StatusOK, tree)
}

// 更新评论
func UpdateCommentPost(c *gin.Context) {
	var comment model.CommentPost

	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	userID, _ := c.Get("user_id")
	comment.UserID = uint(userID.(uint))

	if err := service.UpdateCommentPost(&comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "更新成功"})
}

// 删除评论
func DeleteCommentPost(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	if err := service.DeleteCommentPost(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "删除成功"})
}

// 点赞/取消点赞动态评论
func LikeCommentPost(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	result, err := service.ToggleLike(userID.(uint), "post_comment", uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "点赞失败"})
		return
	}

	msg := "点赞成功"
	if !result.Liked {
		msg = "已取消点赞"
	}
	c.JSON(http.StatusOK, gin.H{"msg": msg, "liked": result.Liked, "likes": result.Likes})
}

// 获取所有动态（管理员，包含已删除）
func GetAllPostsAdmin(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "100"))
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-999"))

	list, total, err := service.GetAllPostsForAdmin(page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// 获取已退回动态列表（管理员）
func GetRejectedPostsAdmin(c *gin.Context) {
	list, err := service.GetRejectedPostsForAdmin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"list": list})
}

// 获取已删除动态列表（管理员）
func GetDeletedPostsAdmin(c *gin.Context) {
	list, err := service.GetDeletedPostsForAdmin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"list": list})
}

// 获取待审核动态列表（管理员）
func GetPendingPostsAdmin(c *gin.Context) {
	list, err := service.GetPendingPostsForAdmin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"list": list})
}

// 审核动态（管理员）
func AuditPost(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Approved bool `json:"approved"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if err := service.AuditPost(uint(id), req.Approved); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "审核失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "审核完成"})
}

// 恢复动态（管理员）
func RestorePost(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := service.RestorePost(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "恢复失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "已恢复"})
}

// ========== 草稿相关 ==========

// 创建草稿
func CreateDraftPost(c *gin.Context) {
	var post model.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	post.UserID = uint(userID.(uint))

	if err := service.CreateDraftPost(&post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "草稿已保存", "data": post})
}

// 更新草稿
func UpdateDraftPost(c *gin.Context) {
	var post model.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	post.UserID = uint(userID.(uint))

	if err := service.UpdateDraftPost(&post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "草稿已更新"})
}

// 获取草稿列表
func GetDraftPosts(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	list, err := service.GetDraftPosts(uint(userID.(uint)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"list": list})
}

// 删除草稿
func DeleteDraftPost(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	if err := service.DeleteDraftPost(uint(id), uint(userID.(uint))); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "草稿已删除"})
}

// 发布草稿（草稿转正式发布，执行敏感词检测）
func PublishDraftPost(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	uid := uint(userID.(uint))

	// 获取草稿
	draft, err := service.GetDraftPostByID(uint(id), uid)
	if err != nil || draft.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "草稿不存在"})
		return
	}

	// 敏感词检测
	hasImages := draft.Images != ""
	hasSensitive := service.HasSensitiveWord(draft.Title + " " + draft.Content)
	autoPass, reason := service.ShouldAutoPass(hasImages, hasSensitive)

	var newStatus int
	if autoPass {
		newStatus = 1
	} else {
		newStatus = -1
	}

	// 更新状态
	if err := service.PublishPostFromDraft(uint(id), uid, newStatus); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发布失败"})
		return
	}

	if newStatus == 1 {
		c.JSON(http.StatusOK, gin.H{"msg": "发布成功"})
	} else {
		c.JSON(http.StatusOK, gin.H{"msg": "提交成功，内容需等待审核", "reason": reason})
	}
}

// 获取热门动态
func GetHotPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	list, total, err := service.GetHotPosts(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
