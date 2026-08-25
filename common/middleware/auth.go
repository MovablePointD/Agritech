package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var secret = []byte("secret_key")

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")

		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token无效"})
			c.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)

		// JWT claims 中的数值会被 JSON 解析为 float64，需要转换为 uint
		if uid, ok := claims["user_id"].(float64); ok {
			c.Set("user_id", uint(uid))
		} else {
			c.Set("user_id", claims["user_id"])
		}
		c.Set("role", claims["role"])

		c.Next()
	}
}

// RequireRole 角色权限中间件
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
			c.Abort()
			return
		}

		userRole := role.(string)
		for _, r := range roles {
			if userRole == r {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "无权限访问"})
		c.Abort()
	}
}

// RequireSysAdmin 只允许系统管理员
func RequireSysAdmin() gin.HandlerFunc {
	return RequireRole("sysadmin")
}

// RequireAdminOrSysAdmin 允许系统管理员或审核人员
func RequireAdminOrSysAdmin() gin.HandlerFunc {
	return RequireRole("sysadmin", "admin")
}
