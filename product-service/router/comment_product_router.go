// router/comment_product_router.go
package router

import (
	"rxtcloud/common/middleware"
	"rxtcloud/product-service/controller"

	"github.com/gin-gonic/gin"
)

func CommentProductRouter(r *gin.RouterGroup) {

	// 游客可访问
	r.GET("/comment_product/:product_id", controller.GetCommentProduct)

	// 需要登录
	auth := r.Group("/comment_product")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.POST("/", controller.CreateCommentProduct)
		auth.PUT("/", controller.UpdateCommentProduct)
		auth.DELETE("/:id", controller.DeleteCommentProduct)
		auth.POST("/like/:id", controller.LikeCommentProduct)
	}
}
