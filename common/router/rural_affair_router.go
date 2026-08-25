package router

import (
	"rxtcloud/common/controller"
	"rxtcloud/common/middleware"

	"github.com/gin-gonic/gin"
)

func RuralAffairRouter(r *gin.RouterGroup) {
	// 公开路由
	r.GET("/affairs", controller.GetRuralAffairs)
	r.GET("/affair/:id", controller.GetRuralAffair)

	// ========== 需要登录的路由 ==========
	auth := r.Group("/affair")
	auth.Use(middleware.AuthMiddleware())
	{
		// 基础操作
		auth.POST("/", controller.CreateRuralAffair)
		auth.PUT("/", controller.UpdateRuralAffair)
		auth.DELETE("/:id", controller.DeleteRuralAffair)
		auth.GET("/my", controller.GetMyRuralAffairs)

		// 完整详情（含修改记录、追问、申诉、权限判断）
		auth.GET("/:id/full", controller.GetAffairFullDetail)

		// 修改（待审核状态）
		auth.PUT("/:id/modify", controller.ModifyRuralAffair)
		auth.GET("/:id/modifications", controller.GetAffairModifications)

		// 追问追答
		auth.POST("/:id/follow-up", controller.AddFollowUpQuestion)
		auth.POST("/:id/follow-up/answer", controller.AddFollowUpAnswer)
		auth.GET("/:id/follow-ups", controller.GetAffairFollowUps)

		// 确认完成（用户）
		auth.POST("/:id/confirm", controller.ConfirmCompleteAffair)

		// 申诉
		auth.POST("/:id/appeal", controller.CreateAppeal)
		auth.POST("/appeal/:id/material", controller.AddAppealMaterial)
	}

	// ========== 审核操作（admin和sysadmin） ==========
	review := r.Group("/affair")
	review.Use(middleware.AuthMiddleware())
	review.Use(middleware.RequireAdminOrSysAdmin())
	{
		review.POST("/audit/:id", controller.AuditRuralAffair)
	}

	// ========== 处理人员操作（processor、admin、sysadmin） ==========
	processor := r.Group("/affair")
	processor.Use(middleware.AuthMiddleware())
	{
		processor.POST("/process/:id", controller.ProcessRuralAffair)
		processor.POST("/start-process/:id", controller.StartProcessRuralAffair)
	}

	// ========== 管理员路由 ==========
	admin := r.Group("/admin/affair")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.RequireAdminOrSysAdmin())
	{
		admin.GET("/pending-audit", controller.GetPendingAuditAffairs)
		admin.GET("/pending-process", controller.GetPendingProcessAffairs)
		admin.GET("/rejected", controller.GetRejectedAffairs)
		admin.GET("/processing", controller.GetProcessingAffairs)
		admin.GET("/completed", controller.GetCompletedAffairs)
		admin.GET("/all", controller.GetAllRuralAffairs)

		// 申诉管理
		admin.POST("/appeal/:id/process", controller.ProcessAppeal)
		admin.GET("/appeals", controller.GetAllAppeals)
	}
}
