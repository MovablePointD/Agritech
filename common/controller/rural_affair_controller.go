package controller

import (
	"rxtcloud/common/config"
	"rxtcloud/common/model"
	"rxtcloud/common/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetRuralAffairs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")
	affairType := c.Query("type")
	keyword := c.Query("keyword")

	list, total, err := service.GetRuralAffairs(page, pageSize, status, affairType, keyword)
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

func GetRuralAffair(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	affair, err := service.GetRuralAffair(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"error": "不存在"})
		return
	}

	c.JSON(200, affair)
}

func CreateRuralAffair(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}

	var affair model.RuralAffair
	if err := c.ShouldBindJSON(&affair); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	affair.UserID = userID
	affair.Status = 1 // 待审核

	if err := service.CreateRuralAffair(&affair); err != nil {
		c.JSON(500, gin.H{"error": "创建失败"})
		return
	}

	c.JSON(200, gin.H{
		"msg":  "提交成功",
		"data": affair,
	})
}

func UpdateRuralAffair(c *gin.Context) {
	var affair model.RuralAffair
	if err := c.ShouldBindJSON(&affair); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := service.UpdateRuralAffair(&affair); err != nil {
		c.JSON(500, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(200, gin.H{"msg": "更新成功"})
}

func DeleteRuralAffair(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := service.DeleteRuralAffair(uint(id)); err != nil {
		c.JSON(500, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(200, gin.H{"msg": "删除成功"})
}

// 获取用户自己的事务列表
func GetMyRuralAffairs(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := service.GetUserRuralAffairs(userID, page, pageSize)
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

// 获取待审核的事务（供审核人员/管理员使用）
func GetPendingAuditAffairs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := service.GetPendingAuditAffairs(page, pageSize)
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

// 审核事务
func AuditRuralAffair(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	var input struct {
		Approved     bool   `json:"approved"`
		AuditorName  string `json:"auditor_name"`
		RejectReason string `json:"reject_reason"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := service.AuditRuralAffair(uint(id), input.Approved, input.AuditorName, input.RejectReason); err != nil {
		c.JSON(500, gin.H{"error": "审核失败"})
		return
	}

	c.JSON(200, gin.H{"msg": "审核成功"})
}

// 获取待处理的事务列表（供处理人员使用）
func GetPendingProcessAffairs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := service.GetPendingProcessAffairs(page, pageSize)
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

// 开始处理事务
func StartProcessRuralAffair(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	userID, _ := GetUserID(c)
	userRole, _ := c.Get("role")
	var user model.User
	config.DB.First(&user, userID)

	var input struct {
		HandlerID   uint   `json:"handler_id"`
		HandlerName string `json:"handler_name"`
	}
	input.HandlerID = userID
	input.HandlerName = user.Nickname

	// 检查权限
	roleStr := ""
	if r, ok := userRole.(string); ok {
		roleStr = r
	}

	if err := service.StartProcessRuralAffairWithCheck(uint(id), input.HandlerID, input.HandlerName, roleStr); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"msg": "已开始处理"})
}

// 处理完成事务
func ProcessRuralAffair(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	userID, _ := GetUserID(c)
	userRole, _ := c.Get("role")
	var user model.User
	config.DB.First(&user, userID)

	var input struct {
		ProcessContent string `json:"process_content"`
		ProcessImages  string `json:"process_images"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	// 检查权限
	roleStr := ""
	if r, ok := userRole.(string); ok {
		roleStr = r
	}

	if err := service.ProcessRuralAffairWithCheck(uint(id), userID, user.Nickname, input.ProcessContent, input.ProcessImages, roleStr); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"msg": "处理完成"})
}

// 获取所有事务（管理员）
func GetAllRuralAffairs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")

	list, total, err := service.GetAllRuralAffairs(page, pageSize, status)
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

// 获取审核不通过的事务（status=3）
func GetRejectedAffairs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := service.GetRejectedAffairs(page, pageSize)
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

// 获取处理中的事务（status=4）
func GetProcessingAffairs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := service.GetProcessingAffairs(page, pageSize)
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

// 获取已完成的事务（status=5）
func GetCompletedAffairs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := service.GetCompletedAffairs(page, pageSize)
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

// ============================================================
// 扩展功能：修改、追问、申诉
// ============================================================

// GetAffairFullDetail 获取事务完整详情（含修改记录、追问、申诉和权限判断）
func GetAffairFullDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	userID, _ := GetUserID(c)
	userRole, _ := c.Get("role")
	roleStr := ""
	if r, ok := userRole.(string); ok {
		roleStr = r
	}

	detail, err := service.GetAffairFullDetail(uint(id), userID, roleStr)
	if err != nil {
		c.JSON(404, gin.H{"error": "事务不存在"})
		return
	}

	c.JSON(200, detail)
}

// ModifyRuralAffair 修改待审核状态的事务
func ModifyRuralAffair(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}

	var user model.User
	config.DB.First(&user, userID)

	var input model.RuralAffair
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := service.ModifyRuralAffair(uint(id), userID, user.Nickname, input); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"msg": "修改成功"})
}

// GetAffairModifications 获取事务修改记录
func GetAffairModifications(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	list, err := service.GetAffairModifications(uint(id))
	if err != nil {
		c.JSON(500, gin.H{"error": "获取修改记录失败"})
		return
	}

	c.JSON(200, gin.H{"list": list})
}

// AddFollowUpQuestion 用户添加追问
func AddFollowUpQuestion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}

	var user model.User
	config.DB.First(&user, userID)

	var input struct {
		Content string `json:"content"`
		Images  string `json:"images"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if input.Content == "" {
		c.JSON(400, gin.H{"error": "追问内容不能为空"})
		return
	}

	followUp, err := service.AddFollowUpQuestion(uint(id), userID, user.Nickname, input.Content, input.Images)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"msg": "追问成功", "data": followUp})
}

// AddFollowUpAnswer 处理人员添加追答
func AddFollowUpAnswer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}

	var user model.User
	config.DB.First(&user, userID)

	var input struct {
		Content string `json:"content"`
		Images  string `json:"images"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if input.Content == "" {
		c.JSON(400, gin.H{"error": "追答内容不能为空"})
		return
	}

	followUp, err := service.AddFollowUpAnswer(uint(id), userID, user.Nickname, input.Content, input.Images)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"msg": "追答成功", "data": followUp})
}

// GetAffairFollowUps 获取事务的追问追答列表
func GetAffairFollowUps(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	list, err := service.GetAffairFollowUps(uint(id))
	if err != nil {
		c.JSON(500, gin.H{"error": "获取失败"})
		return
	}

	c.JSON(200, gin.H{"list": list})
}

// ConfirmCompleteAffair 用户确认事务完成
func ConfirmCompleteAffair(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}

	if err := service.ConfirmCompleteRuralAffair(uint(id), userID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"msg": "已确认完成"})
}

// CreateAppeal 创建申诉
func CreateAppeal(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}

	var user model.User
	config.DB.First(&user, userID)

	var input struct {
		ApplicantType string `json:"applicant_type"` // "user" 或 "handler"
		Reason        string `json:"reason"`
		Images        string `json:"images"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if input.Reason == "" {
		c.JSON(400, gin.H{"error": "申诉理由不能为空"})
		return
	}

	if input.ApplicantType != "user" && input.ApplicantType != "handler" {
		c.JSON(400, gin.H{"error": "无效的申诉人类型"})
		return
	}

	appeal, err := service.CreateAppeal(uint(id), input.ApplicantType, userID, user.Nickname, input.Reason, input.Images)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"msg": "申诉已提交", "data": appeal})
}

// AddAppealMaterial 提交申诉辅助材料
func AddAppealMaterial(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}

	var user model.User
	config.DB.First(&user, userID)

	var input struct {
		SubmitterType string `json:"submitter_type"` // "user" 或 "handler"
		Content       string `json:"content"`
		Images        string `json:"images"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	material, err := service.AddAppealMaterial(uint(id), input.SubmitterType, userID, user.Nickname, input.Content, input.Images)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"msg": "材料已提交", "data": material})
}

// admin/sysadmin 处理申诉
func ProcessAppeal(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	userID, _ := GetUserID(c)
	var user model.User
	config.DB.First(&user, userID)

	var input struct {
		Result  string `json:"result"`  // 处理结果说明
		Approve bool   `json:"approve"` // 是否批准申诉
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := service.ProcessAppeal(uint(id), userID, user.Nickname, input.Result, input.Approve); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"msg": "申诉已处理"})
}

// GetAllAppeals admin获取所有申诉
func GetAllAppeals(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")

	list, total, err := service.GetAllAppeals(status, page, pageSize)
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
