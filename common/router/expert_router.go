package router

import (
	"rxtcloud/common/controller"
	"rxtcloud/common/middleware"

	"github.com/gin-gonic/gin"
)

func ExpertRouter(r *gin.RouterGroup) {

	// 游客可访问
	r.GET("/experts", controller.GetExperts)
	r.GET("/expert/:id", controller.GetExpert)
	r.GET("/experts/profession/:profession", controller.GetExpertsByProfession)
	r.GET("/experts/search", controller.SearchExperts)

	// 用户操作
	user := r.Group("/expert")
	user.Use(middleware.AuthMiddleware())
	{
		user.POST("/apply", controller.ApplyExpert)
		user.PUT("/", controller.UpdateExpert)
		user.DELETE("/:id", controller.DeleteExpert)
		user.GET("/my", controller.GetMyExpert)
	}

	// 管理员操作（sysadmin和admin均可访问）
	admin := r.Group("/admin/expert")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.RequireAdminOrSysAdmin())
	{
		admin.GET("/pending", controller.GetPendingExperts)
		admin.GET("/all", controller.GetAllExpertsCtrl)
		admin.POST("/approve/:id", controller.ApproveExpert)
		admin.POST("/reject/:id", controller.RejectExpert)
	}
}
