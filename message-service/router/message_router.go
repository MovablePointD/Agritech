package router

import (
	"rxtcloud/common/middleware"
	"rxtcloud/message-service/controller"

	"github.com/gin-gonic/gin"
)

func MessageRouter(r *gin.RouterGroup) {
	auth := r.Group("/message")
	auth.Use(middleware.AuthMiddleware())
	{
		// 会话
		auth.GET("/conversations", controller.GetConversations)
		auth.POST("/conversations", controller.CreateConversation)
		auth.GET("/conversation/:id", controller.GetConversationDetail)
		auth.DELETE("/conversation/:id", controller.DeleteConversation)

		// 咨询
		auth.POST("/consultation", controller.CreateConsultation)

		// 消息
		auth.GET("/messages/:id", controller.GetMessages)
		auth.POST("/send", controller.SendMessage)
		auth.POST("/read/:id", controller.MarkMessageAsRead)

		// 未读数
		auth.GET("/unread-count", controller.GetMessageUnreadCount)
	}
}
