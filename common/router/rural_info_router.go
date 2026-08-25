package router

import (
	"rxtcloud/common/controller"
	"rxtcloud/common/middleware"

	"github.com/gin-gonic/gin"
)

// RuralInfoRouter 农村信息介绍模块路由
// 微服务边界：农村基础信息 + 政策公告的CRUD、审核与评论
func RuralInfoRouter(r *gin.RouterGroup) {
	// ============ 农村基础信息 ============

	// 公开路由
	r.GET("/info", controller.GetRuralInfos)
	r.GET("/info/:id", controller.GetRuralInfo)

	// 发布/修改/删除：仅 admin / sysadmin / processor
	auth := r.Group("/info")
	auth.Use(middleware.AuthMiddleware())
	auth.Use(func(c *gin.Context) {
		role, _ := c.Get("role")
		roleStr := role.(string)
		if roleStr == "admin" || roleStr == "sysadmin" || roleStr == "processor" {
			c.Next()
			return
		}
		c.JSON(403, gin.H{"error": "仅管理员、审核人员和事务处理人员可操作"})
		c.Abort()
	})
	{
		auth.POST("/", controller.CreateRuralInfo)
		auth.PUT("/", controller.UpdateRuralInfo)
		auth.DELETE("/:id", controller.DeleteRuralInfo)
	}

	// 管理员审核路由
	admin := r.Group("/admin/info")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.RequireAdminOrSysAdmin())
	{
		admin.GET("/all", controller.AdminAllRuralInfos)
		admin.POST("/audit/:id", controller.AuditRuralInfo)
	}

	// ============ 农村政策公告 ============

	// 公开路由
	r.GET("/policy", controller.GetPolicyNotices)
	r.GET("/policy/:id", controller.GetPolicyNotice)

	// 发布/修改/删除：仅 admin / sysadmin / processor
	policy := r.Group("/policy")
	policy.Use(middleware.AuthMiddleware())
	policy.Use(func(c *gin.Context) {
		role, _ := c.Get("role")
		roleStr := role.(string)
		if roleStr == "admin" || roleStr == "sysadmin" || roleStr == "processor" {
			c.Next()
			return
		}
		c.JSON(403, gin.H{"error": "仅管理员、审核人员和事务处理人员可操作"})
		c.Abort()
	})
	{
		policy.POST("/", controller.CreatePolicyNotice)
		policy.PUT("/", controller.UpdatePolicyNotice)
		policy.DELETE("/:id", controller.DeletePolicyNotice)
	}

	// 管理员审核路由
	adminPolicy := r.Group("/admin/policy")
	adminPolicy.Use(middleware.AuthMiddleware())
	adminPolicy.Use(middleware.RequireAdminOrSysAdmin())
	{
		adminPolicy.GET("/all", controller.AdminAllPolicyNotices)
		adminPolicy.POST("/audit/:id", controller.AuditPolicyNotice)
	}

	// ============ 地址关联查询（公开） ============
	r.GET("/info/associated", controller.GetAssociatedItems)
}
