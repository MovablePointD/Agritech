package router

import (
	"rxtcloud/common/controller"
	"rxtcloud/common/middleware"

	"github.com/gin-gonic/gin"
)

func SensitiveWordRouter(r *gin.RouterGroup) {
	// 需要管理员权限
	admin := r.Group("/admin/sensitive-word")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.RequireSysAdmin())
	{
		admin.GET("/", controller.GetSensitiveWords)
		admin.POST("/", controller.AddSensitiveWord)
		admin.PUT("/", controller.UpdateSensitiveWord)
		admin.DELETE("/", controller.DeleteSensitiveWord)
	}

	// 自动审核设置（需要管理员权限）
	adminSetting := r.Group("/admin/auto-review")
	adminSetting.Use(middleware.AuthMiddleware())
	adminSetting.Use(middleware.RequireSysAdmin())
	{
		adminSetting.GET("/setting", controller.GetAutoReviewSetting)
		adminSetting.POST("/setting", controller.SetAutoReviewSetting)
	}

	// 内容检测（公开接口）
	r.GET("/check-content", controller.CheckContent)
}
