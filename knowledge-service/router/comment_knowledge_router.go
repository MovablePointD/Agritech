package router

import (
	"rxtcloud/common/middleware"
	"rxtcloud/knowledge-service/controller"

	"github.com/gin-gonic/gin"
)

func CommentKnowledgeRouter(r *gin.RouterGroup) {

	// 游客可访问
	r.GET("/comment_knowledge/:knowledge_id", controller.GetCommentKnowledge)

	// 需要登录
	auth := r.Group("/comment_knowledge")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.POST("/", controller.CreateCommentKnowledge)
		auth.PUT("/", controller.UpdateCommentKnowledge)
		auth.DELETE("/:id", controller.DeleteCommentKnowledge)
		auth.POST("/like/:id", controller.LikeCommentKnowledge)
	}
}
