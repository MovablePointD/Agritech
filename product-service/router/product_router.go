package router

import (
	"rxtcloud/common/middleware"
	"rxtcloud/product-service/controller"

	"github.com/gin-gonic/gin"
)

func ProductRouter(r *gin.RouterGroup) {

	// 游客可访问
	r.GET("/product/:id", controller.GetProduct)
	r.GET("/products", controller.GetProducts)
	r.GET("/products/on_sale", controller.GetOnSaleProducts)
	r.GET("/products/type/:type", controller.GetProductsByType)

	// 需要登录
	auth := r.Group("/product")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.POST("/", controller.CreateProduct)
		auth.PUT("/", controller.UpdateProduct)
		auth.DELETE("/:id", controller.DeleteProduct)
		auth.GET("/my", controller.GetMyProducts)
	}

	// 管理员操作
	admin := r.Group("/admin/product")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.RequireAdminOrSysAdmin())
	{
		admin.GET("/all", controller.GetAllProductsAdmin)           // 获取所有商品（含下架）
		admin.GET("/pending", controller.GetPendingProductsAdmin)   // 获取待审核商品
		admin.GET("/rejected", controller.GetRejectedProductsAdmin) // 获取已退回商品
		admin.GET("/hidden", controller.GetHiddenProductsAdmin)     // 获取已隐藏商品
		admin.POST("/audit/:id", controller.AuditProduct)           // 审核商品
		admin.POST("/restore/:id", controller.RestoreProductAdmin)  // 恢复商品（上架）
		admin.POST("/offshelf/:id", controller.OffShelfProduct)     // 下架商品
	}
}
