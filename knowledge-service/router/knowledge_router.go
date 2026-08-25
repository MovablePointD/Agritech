package router

import (
	"rxtcloud/common/middleware"
	"rxtcloud/knowledge-service/controller"

	"github.com/gin-gonic/gin"
)

func KnowledgeRouter(r *gin.RouterGroup) {
	// 公开路由（GET /knowledge, GET /knowledges, GET /knowledge/:id 等）
	// 已迁移到 router.go InitRouter 中直接注册到 /api 下，
	// 避免与下方 r.Group("/knowledge") 发生 Gin RedirectTrailingSlash 冲突

	// 用户操作
	user := r.Group("/knowledge")
	user.Use(middleware.AuthMiddleware())
	{
		user.POST("/", controller.CreateKnowledge)
		user.PUT("/", controller.UpdateKnowledge)
		user.DELETE("/:id", controller.DeleteKnowledge)
		user.POST("/like/:id", controller.LikeKnowledge)
		user.POST("/comment", controller.CreateCommentKnowledge)
		user.PUT("/comment", controller.UpdateCommentKnowledge)
		user.DELETE("/comment/:id", controller.DeleteCommentKnowledge)
		user.POST("/comment/like/:id", controller.LikeCommentKnowledge)
		user.GET("/my", controller.GetMyKnowledge)
		// 草稿
		user.POST("/draft", controller.CreateDraftKnowledge)
		user.PUT("/draft", controller.UpdateDraftKnowledge)
		user.GET("/drafts", controller.GetDraftKnowledge)
		user.DELETE("/draft/:id", controller.DeleteDraftKnowledge)
		user.POST("/draft/publish/:id", controller.PublishDraftKnowledge)
	}

	// 管理员操作
	admin := r.Group("/admin/knowledge")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.RequireAdminOrSysAdmin())
	{
		admin.GET("/all", controller.GetAllKnowledgeAdmin)
		admin.GET("/pending", controller.GetPendingKnowledgeAdmin)
		admin.GET("/rejected", controller.GetRejectedKnowledgeAdmin)
		admin.GET("/deleted", controller.GetDeletedKnowledgeAdmin)
		admin.POST("/audit/:id", controller.AuditKnowledge)
		admin.POST("/restore/:id", controller.RestoreKnowledge)
	}
}
