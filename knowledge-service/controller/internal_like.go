package controller

import (
	"net/http"

	"rxtcloud/common/client"
	"rxtcloud/knowledge-service/service"

	"github.com/gin-gonic/gin"
)

// HandleInternalLike 内部接口：接收来自 message-service 的点赞转发
// POST /api/internal/like
func HandleInternalLike(c *gin.Context) {
	var req client.LikeCallbackBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if req.TargetType != "knowledge" && req.TargetType != "knowledge_comment" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的目标类型: " + req.TargetType})
		return
	}

	result, err := service.ToggleLike(req.UserID, req.TargetType, req.TargetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"liked": result.Liked,
		"likes": result.Likes,
	})
}
