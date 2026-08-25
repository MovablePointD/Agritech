package router

import (
	"rxtcloud/common/middleware"
	"rxtcloud/product-service/controller"

	"github.com/gin-gonic/gin"
)

func CartRouter(r *gin.RouterGroup) {
	cart := r.Group("/cart")
	cart.Use(middleware.AuthMiddleware())
	{
		cc := &controller.CartController{}
		cart.GET("/", cc.GetMyCart)
		cart.GET("/count", cc.GetCartCount)
		cart.GET("/:id", cc.GetCartItem)
		cart.POST("/", cc.AddToCart)
		cart.POST("/batch", cc.BatchAddToCart)
		cart.POST("/checkout", cc.CheckoutCart)
		cart.POST("/direct_buy", cc.DirectBuy)
		cart.PUT("/:id", cc.UpdateCartItem)
		cart.DELETE("/:id", cc.RemoveFromCart)
		cart.DELETE("/", cc.ClearCart)
	}
}
