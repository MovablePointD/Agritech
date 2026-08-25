package router

import (
	"rxtcloud/common/controller"
	"rxtcloud/common/middleware"

	"github.com/gin-gonic/gin"
)

func CommentRuralRouter(r *gin.RouterGroup) {
	// 游客可访问
	r.GET("/comment_rural", controller.GetCommentRuralInfos)

	// 需登录
	auth := r.Group("/comment_rural")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.POST("/", controller.CreateCommentRuralInfo)
		auth.PUT("/", controller.UpdateCommentRuralInfo)
		auth.DELETE("/:id", controller.DeleteCommentRuralInfo)
		auth.POST("/like/:id", controller.LikeCommentRuralInfo)
	}
}
