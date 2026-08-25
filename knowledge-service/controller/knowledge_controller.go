// controller/knowledge_controller.go
package controller

import (
	"net/http"
	"strconv"

	"rxtcloud/knowledge-service/model"
	"rxtcloud/knowledge-service/service"

	"github.com/gin-gonic/gin"
)

// 创建
func CreateKnowledge(c *gin.Context) {
	var k model.Knowledge
	if err := c.ShouldBindJSON(&k); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 设置当前用户ID
	userIDVal, exists := c.Get("user_id")
	if !exists || userIDVal == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
		return
	}
	k.UserID = uint(userID)

	// 检查敏感词
	hasImages := k.ImageURL != "" || k.VideoURL != ""
	hasSensitive := service.HasSensitiveWord(k.Title + " " + k.Content)
	autoPass, reason := service.ShouldAutoPass(hasImages, hasSensitive)

	if autoPass {
		k.Status = 1 // 自动通过
	} else {
		k.Status = -1 // 待审核
	}

	service.CreateKnowledge(k)

	if k.Status == 1 {
		c.JSON(http.StatusOK, gin.H{"msg": "创建成功", "data": k})
	} else {
		c.JSON(http.StatusOK, gin.H{"msg": "提交成功，内容需等待审核", "reason": reason, "data": k})
	}
}

// 分页查询列表
func GetKnowledgeList(c *gin.Context) {
	// 获取参数
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	// 防止异常参数
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	keyword := c.Query("keyword")

	// 获取当前用户ID（可选，未登录也可查看列表）
	var userID uint
	if uid, exists := c.Get("user_id"); exists {
		userID = uint(uid.(uint))
	}

	list, total, err := service.GetKnowledgeList(page, pageSize, keyword, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     list,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// 查询单个
func GetKnowledge(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	k, err := service.GetKnowledgeByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "知识不存在"})
		return
	}

	// 获取当前用户ID
	var userID uint
	if uid, exists := c.Get("user_id"); exists {
		userID = uint(uid.(uint))
	}

	// 只有已发布(1)或作者本人才能查看
	if k.Status != 1 && k.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "知识不存在"})
		return
	}

	// 判断当前用户是否已点赞
	isLiked := userID > 0 && service.IsLiked(userID, "knowledge", uint(id))

	c.JSON(http.StatusOK, gin.H{
		"knowledge": k,
		"is_liked":  isLiked,
	})
}

// 更新
func UpdateKnowledge(c *gin.Context) {
	var k model.Knowledge
	if err := c.ShouldBindJSON(&k); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 获取当前用户ID
	userIDVal, exists := c.Get("user_id")
	if !exists || userIDVal == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
		return
	}
	// 防止修改 user_id，确保只能修改自己的知识
	if existing, _ := service.GetKnowledgeByID(k.ID); existing.ID != 0 {
		// 检查是否为本人或管理员
		if existing.UserID != uint(userID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权修改他人的知识"})
			return
		}
		k.UserID = existing.UserID
	}
	service.UpdateKnowledge(k)
	c.JSON(http.StatusOK, gin.H{"msg": "更新成功"})
}

// 删除
func DeleteKnowledge(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	// 获取当前用户ID
	userIDVal, exists := c.Get("user_id")
	if !exists || userIDVal == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
		return
	}
	// 检查是否为本人或管理员
	existing, _ := service.GetKnowledgeByID(uint(id))
	if existing.ID != 0 && existing.UserID != uint(userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权删除他人的知识"})
		return
	}
	service.DeleteKnowledge(uint(id))
	c.JSON(http.StatusOK, gin.H{"msg": "删除成功"})
}

// 获取当前用户发布的知识列表
func GetMyKnowledge(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists || userIDVal == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
		return
	}
	list, _ := service.GetMyKnowledge(uint(userID))
	c.JSON(http.StatusOK, list)
}

// 获取所有知识（管理员，包含已删除）
func GetAllKnowledgeAdmin(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-999"))

	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 100
	}

	list, total, err := service.GetAllKnowledgeForAdmin(page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     list,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// 获取待审核知识列表（管理员）
func GetPendingKnowledgeAdmin(c *gin.Context) {
	list, err := service.GetPendingKnowledgeForAdmin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"list": list})
}

// 获取已退回知识列表（管理员）
func GetRejectedKnowledgeAdmin(c *gin.Context) {
	list, err := service.GetRejectedKnowledgeForAdmin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"list": list})
}

// 获取已删除知识列表（管理员）
func GetDeletedKnowledgeAdmin(c *gin.Context) {
	list, err := service.GetDeletedKnowledgeForAdmin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"list": list})
}

// 审核知识（管理员）

func AuditKnowledge(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Approved bool `json:"approved"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if err := service.AuditKnowledge(uint(id), req.Approved); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "审核失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "审核完成"})
}

// 恢复知识（管理员）
func RestoreKnowledge(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := service.RestoreKnowledge(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "恢复失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "已恢复"})
}

// 点赞/取消点赞知识
func LikeKnowledge(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userIDVal, exists := c.Get("user_id")
	if !exists || userIDVal == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
		return
	}

	result, err := service.ToggleLike(uint(userID), "knowledge", uint(id))
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

// ========== 草稿相关 ==========

// 创建知识草稿
func CreateDraftKnowledge(c *gin.Context) {
	var k model.Knowledge
	if err := c.ShouldBindJSON(&k); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	userIDVal, exists := c.Get("user_id")
	if !exists || userIDVal == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
		return
	}
	k.UserID = uint(userID)

	if err := service.CreateDraftKnowledge(&k); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "草稿已保存", "data": k})
}

// 更新知识草稿
func UpdateDraftKnowledge(c *gin.Context) {
	var k model.Knowledge
	if err := c.ShouldBindJSON(&k); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	userIDVal, exists := c.Get("user_id")
	if !exists || userIDVal == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
		return
	}
	k.UserID = uint(userID)

	if err := service.UpdateDraftKnowledge(&k); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "草稿已更新"})
}

// 获取知识草稿列表
func GetDraftKnowledge(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists || userIDVal == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
		return
	}

	list, err := service.GetDraftKnowledge(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"list": list})
}

// 删除知识草稿
func DeleteDraftKnowledge(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userIDVal, exists := c.Get("user_id")
	if !exists || userIDVal == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
		return
	}

	if err := service.DeleteDraftKnowledge(uint(id), uint(userID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "草稿已删除"})
}

// 发布知识草稿（草稿转正式发布，执行敏感词检测）
func PublishDraftKnowledge(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userIDVal, exists := c.Get("user_id")
	if !exists || userIDVal == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
		return
	}
	uid := uint(userID)

	// 获取草稿
	draft, err := service.GetDraftKnowledgeByID(uint(id), uid)
	if err != nil || draft.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "草稿不存在"})
		return
	}

	// 敏感词检测
	hasImages := draft.ImageURL != "" || draft.VideoURL != ""
	hasSensitive := service.HasSensitiveWord(draft.Title + " " + draft.Content)
	autoPass, reason := service.ShouldAutoPass(hasImages, hasSensitive)

	var newStatus int
	if autoPass {
		newStatus = 1
	} else {
		newStatus = -1
	}

	// 更新状态发布
	if err := service.PublishKnowledgeFromDraft(uint(id), uid, newStatus); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发布失败"})
		return
	}

	if newStatus == 1 {
		c.JSON(http.StatusOK, gin.H{"msg": "发布成功"})
	} else {
		c.JSON(http.StatusOK, gin.H{"msg": "提交成功，内容需等待审核", "reason": reason})
	}
}

// 获取热门知识
func GetHotKnowledge(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	list, total, err := service.GetHotKnowledgeList(page, pageSize)
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
