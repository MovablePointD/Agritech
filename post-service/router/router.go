package router

import (
	"rxtcloud/post-service/controller"

	"github.com/gin-gonic/gin"
)

// InitRouter 初始化 post-service 路由
// 只注册【动态模块】（含动态评论、点赞）
func InitRouter(r *gin.Engine) {
	api := r.Group("/api")
	{
		PostRouter(api)

		// 内部 API：接收 message-service 的点赞转发
		api.POST("/internal/like", controller.HandleInternalLike)
	}
}
