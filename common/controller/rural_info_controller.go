package controller

import (
	"log"
	"net/http"
	"rxtcloud/common/model"
	"rxtcloud/common/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ==================== RuralInfo 控制器 ====================

// GetRuralInfos 公开列表（仅status=1）
func GetRuralInfos(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	infoType := c.Query("type")
	keyword := c.Query("keyword")

	list, total, err := service.GetRuralInfoList(page, pageSize, infoType, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetRuralInfo 详情
func GetRuralInfo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	info, err := service.GetRuralInfo(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "不存在"})
		return
	}
	c.JSON(http.StatusOK, info)
}

// CreateRuralInfo 创建（需登录）
func CreateRuralInfo(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	var info model.RuralInfo
	if err := c.ShouldBindJSON(&info); err != nil {
		log.Printf("[RuralInfo] JSON绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	info.UserID = userID
	info.Status = -1 // 默认待审核
	if err := service.CreateRuralInfo(&info); err != nil {
		log.Printf("[RuralInfo] 创建失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "创建成功", "data": info})
}

// UpdateRuralInfo 更新（需登录）
func UpdateRuralInfo(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	var info model.RuralInfo
	if err := c.ShouldBindJSON(&info); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	// 验证所有权
	existing, err := service.GetRuralInfo(info.ID)
	if err != nil || existing.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}
	if err := service.UpdateRuralInfo(&info); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "更新成功"})
}

// DeleteRuralInfo 软删除（需登录）
func DeleteRuralInfo(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	existing, err := service.GetRuralInfo(uint(id))
	if err != nil || existing.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}
	if err := service.DeleteRuralInfo(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "删除成功"})
}

// AdminAllRuralInfos 管理端全部列表
func AdminAllRuralInfos(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")
	keyword := c.Query("keyword")

	list, total, err := service.GetAllRuralInfos(page, pageSize, status, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// AuditRuralInfo 审核
func AuditRuralInfo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	var req struct {
		Status       int    `json:"status"`        // 1通过 0退回
		RejectReason string `json:"reject_reason"` // 退回原因
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if err := service.AuditRuralInfo(uint(id), req.Status, req.RejectReason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "审核完成"})
}

// ==================== PolicyNotice 控制器 ====================

// GetPolicyNotices 公开列表
func GetPolicyNotices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	category := c.Query("category")
	keyword := c.Query("keyword")

	list, total, err := service.GetPolicyNoticeList(page, pageSize, category, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetPolicyNotice 详情
func GetPolicyNotice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	notice, err := service.GetPolicyNotice(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "不存在"})
		return
	}
	c.JSON(http.StatusOK, notice)
}

// CreatePolicyNotice 创建
func CreatePolicyNotice(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	var notice model.PolicyNotice
	if err := c.ShouldBindJSON(&notice); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	notice.UserID = userID
	notice.Status = -1 // 默认待审核
	if err := service.CreatePolicyNotice(&notice); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "创建成功", "data": notice})
}

// UpdatePolicyNotice 更新
func UpdatePolicyNotice(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	var notice model.PolicyNotice
	if err := c.ShouldBindJSON(&notice); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	existing, err := service.GetPolicyNotice(notice.ID)
	if err != nil || existing.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}
	if err := service.UpdatePolicyNotice(&notice); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "更新成功"})
}

// DeletePolicyNotice 软删除
func DeletePolicyNotice(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	existing, err := service.GetPolicyNotice(uint(id))
	if err != nil || existing.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		return
	}
	if err := service.DeletePolicyNotice(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "删除成功"})
}

// AdminAllPolicyNotices 管理端全部列表
func AdminAllPolicyNotices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")
	keyword := c.Query("keyword")

	list, total, err := service.GetAllPolicyNotices(page, pageSize, status, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// AuditPolicyNotice 审核
func AuditPolicyNotice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	var req struct {
		Status int `json:"status"` // 1通过 0退回
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if err := service.AuditPolicyNotice(uint(id), req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "审核完成"})
}

// ==================== 地址关联控制器 ====================

// GetAssociatedItems 根据农村地址查询关联的公告和事务
func GetAssociatedItems(c *gin.Context) {
	address := c.Query("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "地址参数不能为空"})
		return
	}
	result, err := service.GetAssociatedItemsByAddress(address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, result)
}
