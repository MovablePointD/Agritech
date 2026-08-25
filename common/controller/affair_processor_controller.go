package controller

import (
	"rxtcloud/common/config"
	"rxtcloud/common/model"
	"rxtcloud/common/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ApplyProcessor 申请成为事务处理人员
// ApplyProcessor 申请成为事务处理人，需登录后提交申请信息
func ApplyProcessor(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}

	var processor model.AffairProcessor
	if err := c.ShouldBindJSON(&processor); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	processor.UserID = userID
	processor.Status = 0

	if err := service.ApplyProcessor(&processor); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"msg":  "申请提交成功",
		"data": processor,
	})
}

// GetMyProcessor 获取我的处理人员申请状态
func GetMyProcessor(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}

	processor, err := service.GetMyProcessor(userID)
	if err != nil {
		c.JSON(200, nil)
		return
	}

	c.JSON(200, processor)
}

// CheckIsProcessor 检查是否为事务处理人员
func CheckIsProcessor(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}

	isProcessor := service.IsProcessor(userID)
	c.JSON(200, gin.H{"is_processor": isProcessor})
}

// GetPendingProcessors 获取待审核的处理人员申请
func GetPendingProcessors(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := service.GetPendingProcessors(page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"error": "获取失败"})
		return
	}

	c.JSON(200, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ApproveProcessor 批准处理人员申请
func ApproveProcessorCtrl(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	var user model.User
	userID, _ := c.Get("user_id")
	config.DB.First(&user, userID)

	if err := service.ApproveProcessor(uint(id), user.Nickname); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"msg": "批准成功"})
}

// RejectProcessor 拒绝处理人员申请
func RejectProcessorCtrl(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	var input struct {
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	var user model.User
	userID, _ := c.Get("user_id")
	config.DB.First(&user, userID)

	if err := service.RejectProcessor(uint(id), user.Nickname, input.Reason); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"msg": "已拒绝"})
}

// GetRecommendedAffairs 获取推荐的事务列表
func GetRecommendedAffairs(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := service.GetRecommendedAffairs(userID, page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"error": "获取失败"})
		return
	}

	c.JSON(200, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetMyHandledAffairs 获取我处理过的事务
func GetMyHandledAffairs(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := service.GetMyHandledAffairs(userID, page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"error": "获取失败"})
		return
	}

	c.JSON(200, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetAllProcessors 获取所有处理人员申请（包含所有状态）
func GetAllProcessors(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "100"))

	list, total, err := service.GetAllProcessors(page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"error": "获取失败"})
		return
	}

	c.JSON(200, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// DisableProcessor 禁用处理人员
func DisableProcessorCtrl(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := service.DisableProcessor(uint(id)); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"msg": "已禁用"})
}
