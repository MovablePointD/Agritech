package router

import (
	"github.com/gin-gonic/gin"
)

// InitRouter 初始化 message-service 路由
// 只注册【消息模块】和【通知模块】
// 点赞转发通过 internal API 实现
func InitRouter(r *gin.Engine) {
	api := r.Group("/api")
	{
		MessageRouter(api)
		NotificationRouter(api)
	}
}
