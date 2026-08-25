package router

import (
	"rxtcloud/common/controller"
	"rxtcloud/common/middleware"

	"github.com/gin-gonic/gin"
)

func AffairProcessorRouter(r *gin.RouterGroup) {
	// 需要登录的路由
	auth := r.Group("/affair-processor")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.POST("/apply", controller.ApplyProcessor)             // 申请成为处理人员
		auth.GET("/my", controller.GetMyProcessor)                 // 获取我的申请状态
		auth.GET("/check", controller.CheckIsProcessor)            // 检查是否为处理人员
		auth.GET("/recommended", controller.GetRecommendedAffairs) // 获取推荐的事务
		auth.GET("/handled", controller.GetMyHandledAffairs)       // 获取我处理过的事务
	}

	// 管理员路由
	admin := r.Group("/admin/processor")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.RequireAdminOrSysAdmin())
	{
		admin.GET("/pending", controller.GetPendingProcessors)      // 获取待审核申请
		admin.GET("/all", controller.GetAllProcessors)              // 获取所有处理人员申请
		admin.POST("/approve/:id", controller.ApproveProcessorCtrl) // 批准申请
		admin.POST("/reject/:id", controller.RejectProcessorCtrl)   // 拒绝申请
		admin.POST("/disable/:id", controller.DisableProcessorCtrl) // 禁用处理人员
	}
}
