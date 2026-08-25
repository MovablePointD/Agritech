package router

import (
	"rxtcloud/common/middleware"
	"rxtcloud/product-service/controller"

	"github.com/gin-gonic/gin"
)

func OrderRouter(r *gin.RouterGroup) {

	// 游客可访问
	r.GET("/order/:id", controller.GetOrder)
	r.GET("/orders", controller.GetOrders)

	// 用户操作（需登录）
	user := r.Group("/order")
	user.Use(middleware.AuthMiddleware())
	{
		user.POST("/", controller.CreateOrder)
		user.PUT("/", controller.UpdateOrder)
		user.DELETE("/:id", controller.DeleteOrder)
		user.POST("/pay/:id", controller.PayOrder)
		user.POST("/cancel/:id", controller.CancelOrder)
		user.POST("/confirm/:id", controller.ConfirmReceiveOrder)
		user.GET("/my", controller.GetMyOrders)
	}

	// 管理员/商家操作
	admin := r.Group("/order")
	admin.Use(middleware.AuthMiddleware())
	{
		admin.POST("/ship/:id", controller.ShipOrder)
	}

	// 卖家订单（需登录）
	seller := r.Group("/order")
	seller.Use(middleware.AuthMiddleware())
	{
		seller.GET("/seller", controller.GetMySellerOrders)
	}
}
