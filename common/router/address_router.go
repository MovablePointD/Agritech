package router

import (
	"rxtcloud/common/controller"
	"rxtcloud/common/middleware"

	"github.com/gin-gonic/gin"
)

func AddressRouter(r *gin.RouterGroup) {

	// 需要登录
	auth := r.Group("/address")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.POST("/", controller.CreateAddress)
		auth.GET("/", controller.GetMyAddresses)
		auth.GET("/default", controller.GetDefaultAddress)
		auth.GET("/:id", controller.GetAddress)
		auth.PUT("/", controller.UpdateAddress)
		auth.DELETE("/:id", controller.DeleteAddress)
		auth.POST("/default/:id", controller.SetDefaultAddress)
	}
}
