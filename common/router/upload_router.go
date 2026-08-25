package router

import (
	"rxtcloud/common/controller"
	"rxtcloud/common/middleware"

	"github.com/gin-gonic/gin"
)

func UploadRouter(r *gin.RouterGroup) {

	// 通用上传（需登录）
	auth := r.Group("")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.POST("/upload", controller.UploadFile)
		auth.POST("/uploads", controller.UploadFiles)
		auth.POST("/upload/avatar", controller.UploadAvatar)
		auth.POST("/upload/knowledge", controller.UploadKnowledgeImage)
		auth.POST("/upload/product", controller.UploadProductImage)
		auth.POST("/upload/post", controller.UploadPostImage)
		auth.POST("/upload/video", controller.UploadVideo)
		auth.DELETE("/upload", controller.DeleteFile)
		auth.GET("/files", controller.ListFiles)
	}
}
