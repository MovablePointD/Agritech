package controller

import (
	"net/http"
	"strconv"

	"rxtcloud/common/model"
	"rxtcloud/common/service"

	"github.com/gin-gonic/gin"
)

// 申请成为专家
func ApplyExpert(c *gin.Context) {
	var expert model.Expert
	println("申请成为专家")
	if err := c.ShouldBindJSON(&expert); err != nil {
		println("参数错误")
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		println("未登录")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	// 检查是否已是专家或申请中（status=0待审核或status=1已通过
	existing, _ := service.GetExpertByUserID(uint(userID.(uint)))
	println("检查是否已是专家或申请中")
	if existing != nil && (existing.Status == 0 || existing.Status == 1) {
		println("已是专家或申请中")
		c.JSON(http.StatusBadRequest, gin.H{"error": "已是专家或申请中"})
		return
	}

	expert.UserID = uint(userID.(uint))
	expert.Status = 0 // 重置为待审核

	if err := service.CreateExpert(&expert); err != nil {
		println("申请失败:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "申请失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "申请已提交，请等待审核"})
}

// 获取专家列表
func GetExperts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	list, total, err := service.GetExperts(page, pageSize)
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

// 获取单个专家
func GetExpert(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	expert, err := service.GetExpertByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "专家不存在"})
		return
	}

	c.JSON(http.StatusOK, expert)
}

// 获取我的专家信息
func GetMyExpert(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	expert, err := service.GetExpertByUserID(uint(userID.(uint)))
	if err != nil {
		c.JSON(http.StatusOK, nil)
		return
	}

	c.JSON(http.StatusOK, expert)
}

// 更新专家信息
func UpdateExpert(c *gin.Context) {
	var expert model.Expert

	if err := c.ShouldBindJSON(&expert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	userID, _ := c.Get("user_id")
	existing, _ := service.GetExpertByUserID(uint(userID.(uint)))
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "您还不是专家"})
		return
	}

	// 只有待审核或已通过才能更新
	if existing.Status != 0 && existing.Status != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "当前状态不允许修改"})
		return
	}

	expert.ID = existing.ID
	expert.UserID = existing.UserID
	expert.Status = existing.Status

	if err := service.UpdateExpert(&expert); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "更新成功"})
}

// 删除专家（注销专家身份�?
func DeleteExpert(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	userID, _ := c.Get("user_id")
	expert, _ := service.GetExpertByID(uint(id))

	if expert == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "专家不存在"})
		return
	}

	// 只能本人删除
	if expert.UserID != uint(userID.(uint)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作"})
		return
	}

	if err := service.DeleteExpert(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "已注销专家身份"})
}

// 获取待审核专家（管理员）
func GetPendingExperts(c *gin.Context) {
	list, err := service.GetPendingExperts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"list": list})
}

// 审核通过专家（管理员）
func ApproveExpert(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	if err := service.ApproveExpert(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "审核失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "审核通过"})
}

// 审核拒绝专家（管理员）
func RejectExpert(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	if err := service.RejectExpert(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "审核失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "已拒绝"})
}

// 按专业获取专家
func GetExpertsByProfession(c *gin.Context) {
	profession := c.Param("profession")

	list, err := service.GetExpertsByProfession(profession)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"list": list})
}

// 搜索专家
func SearchExperts(c *gin.Context) {
	keyword := c.DefaultQuery("keyword", "")

	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入关键词"})
		return
	}

	list, err := service.SearchExperts(keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"list": list})
}

// 获取所有专家申请（管理员，包含所有状态）
func GetAllExpertsCtrl(c *gin.Context) {
	list, err := service.GetAllExperts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"list": list})
}
