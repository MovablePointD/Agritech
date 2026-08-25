package router

import (
	"rxtcloud/common/middleware"
	"rxtcloud/post-service/controller"

	"github.com/gin-gonic/gin"
)

func PostRouter(r *gin.RouterGroup) {

	// 游客可访问
	r.GET("/posts", controller.GetPosts)
	r.GET("/posts/hot", controller.GetHotPosts)
	// 注意：/post/:id/comments 必须放在 /post/:id 前面，否则会被 :id 匹配
	r.GET("/post/:id/comments", controller.GetCommentPosts)
	r.GET("/post/:id", controller.GetPost)

	// 用户操作
	user := r.Group("/post")
	user.Use(middleware.AuthMiddleware())
	{
		user.POST("/", controller.CreatePost)
		user.PUT("/", controller.UpdatePost)
		user.DELETE("/:id", controller.DeletePost)
		user.POST("/like/:id", controller.LikePost)
		user.POST("/comment", controller.CreateCommentPost)
		user.PUT("/comment", controller.UpdateCommentPost)
		user.DELETE("/comment/:id", controller.DeleteCommentPost)
		user.POST("/comment/like/:id", controller.LikeCommentPost)
		user.GET("/my", controller.GetMyPosts)
		// 草稿
		user.POST("/draft", controller.CreateDraftPost)
		user.PUT("/draft", controller.UpdateDraftPost)
		user.GET("/drafts", controller.GetDraftPosts)
		user.DELETE("/draft/:id", controller.DeleteDraftPost)
		user.POST("/draft/publish/:id", controller.PublishDraftPost)
	}

	// 管理员操作
	admin := r.Group("/admin/post")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.RequireAdminOrSysAdmin())
	{
		admin.GET("/all", controller.GetAllPostsAdmin)           // 获取所有动态（含已删除）
		admin.GET("/pending", controller.GetPendingPostsAdmin)   // 获取待审核动态
		admin.GET("/rejected", controller.GetRejectedPostsAdmin) // 获取已退回动态
		admin.GET("/deleted", controller.GetDeletedPostsAdmin)   // 获取已删除动态
		admin.POST("/audit/:id", controller.AuditPost)           // 审核动态
		admin.POST("/restore/:id", controller.RestorePost)       // 恢复动态
	}
}
