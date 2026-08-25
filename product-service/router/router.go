package router

import (
	"rxtcloud/product-service/controller"

	"github.com/gin-gonic/gin"
)

// InitRouter 初始化 product-service 路由
// 只注册【商品模块】【订单模块】【购物车模块】【商品评论模块】
func InitRouter(r *gin.Engine) {
	api := r.Group("/api")
	{
		ProductRouter(api)
		OrderRouter(api)
		CartRouter(api)
		CommentProductRouter(api)

		// 内部 API：接收 message-service 的点赞转发
		api.POST("/internal/like", controller.HandleInternalLike)
	}
}
