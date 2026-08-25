package router

import (
	"rxtcloud/knowledge-service/controller"

	"github.com/gin-gonic/gin"
)

// InitRouter 初始化 knowledge-service 路由
// 只注册【知识模块】和【知识评论模块】
func InitRouter(r *gin.Engine) {
	api := r.Group("/api")
	{
		// 公开路由：直接注册在 /api 下，避免与 KnowledgeRouter 内部的
		// r.Group("/knowledge") 产生路径冲突（Gin RedirectTrailingSlash 重定向死循环）
		api.GET("/knowledges", controller.GetKnowledgeList)
		api.GET("/knowledge", controller.GetKnowledgeList)
		api.GET("/knowledge/hot", controller.GetHotKnowledge)
		// 以下带参数的路由放在 /hot 后面避免冒号参数匹配
		api.GET("/knowledge/:id", controller.GetKnowledge)
		api.GET("/knowledge/:id/comments", controller.GetCommentKnowledge)

		KnowledgeRouter(api)
		CommentKnowledgeRouter(api)

		// 内部 API：接收 message-service 的点赞转发
		api.POST("/internal/like", controller.HandleInternalLike)
	}
}
