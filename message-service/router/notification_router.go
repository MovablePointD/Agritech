package router

import (
	"rxtcloud/common/middleware"
	"rxtcloud/message-service/controller"

	"github.com/gin-gonic/gin"
)

func NotificationRouter(r *gin.RouterGroup) {
	auth := r.Group("/notification")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.GET("/", controller.GetMyNotifications)
		auth.GET("/unread-count", controller.GetNotificationUnreadCount)
		auth.POST("/read/:id", controller.MarkNotificationAsRead)
		auth.POST("/read-all", controller.MarkAllAsRead)
	}
}
