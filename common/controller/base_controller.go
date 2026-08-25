package controller

import "github.com/gin-gonic/gin"

// getUserID 从context获取用户ID（处理float64类型转换）
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	// JWT claims中的user_id是float64类型
	switch v := userID.(type) {
	case float64:
		return uint(v), true
	case uint:
		return v, true
	case int:
		return uint(v), true
	default:
		return 0, false
	}
}
