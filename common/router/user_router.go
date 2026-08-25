package router

import (
	"rxtcloud/common/controller"
	"rxtcloud/common/middleware"

	"github.com/gin-gonic/gin"
)

func UserRouter(r *gin.RouterGroup) {
	api := r.Group("/user")
	{
		api.POST("/register", controller.Register)
		api.POST("/login", controller.Login)
		api.POST("/reset-password", controller.ResetPassword)

		// 需要认证的路由
		auth := api.Group("")
		auth.Use(middleware.AuthMiddleware())
		{
			auth.GET("/info", controller.GetCurrentUser)
			auth.GET("/list", controller.SearchUsers)
			auth.PUT("/", controller.UpdateUser)
			auth.POST("/delete-account", controller.SoftDeleteUser)
		}

		// 管理员路由
		api.GET("/", controller.GetUserList)
		api.DELETE("/:id", controller.DeleteUser)

		// 管理员用户管理路由
		adminGroup := r.Group("/admin/user")
		adminGroup.Use(middleware.AuthMiddleware(), middleware.RequireAdminOrSysAdmin())
		{
			adminGroup.GET("/all", controller.GetAllUsersForAdmin)       // 全部用户列表
			adminGroup.GET("/deleted", controller.GetDeletedUsers)       // 已注销用户列表
			adminGroup.PUT("/deleted/:id", controller.UpdateDeletedUser) // 修改已注销用户
			adminGroup.POST("/:id/ban", controller.BanUser)              // 封禁用户
			adminGroup.POST("/:id/unban", controller.UnbanUser)          // 解封用户
		}

		// 内部 API：供其他微服务通过 Nacos 服务发现调用用户信息
		// 无需 JWT 认证（仅限内部网络）
		api.GET("/info/:id", controller.GetUserInfoAPI)
	}
}
