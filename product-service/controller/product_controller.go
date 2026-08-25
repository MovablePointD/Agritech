package controller

import (
	"net/http"
	"strconv"

	"rxtcloud/product-service/model"
	"rxtcloud/product-service/service"

	"github.com/gin-gonic/gin"
)

// 创建商品
func CreateProduct(c *gin.Context) {
	var product model.Product

	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 从中间件获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	product.Publisher = uint(userID.(uint))

	// 检查敏感词
	hasImages := product.ImageURL != ""
	hasSensitive := service.HasSensitiveWord(product.Title + " " + product.Content)
	autoPass, reason := service.ShouldAutoPass(hasImages, hasSensitive)

	if autoPass {
		product.Status = 1 // 自动通过
	} else {
		product.Status = -1 // 待审核
	}

	if err := service.CreateProduct(&product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	if product.Status == 1 {
		c.JSON(http.StatusOK, gin.H{"msg": "创建成功", "data": product})
	} else {
		c.JSON(http.StatusOK, gin.H{"msg": "提交成功，内容需等待审核", "reason": reason, "data": product})
	}
}

// 获取商品列表（分页）
func GetProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.DefaultQuery("keyword", "")
	productType := c.DefaultQuery("type", "")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 获取当前用户ID（可选，未登录也可查看列表）
	var userID uint
	if uid, exists := c.Get("user_id"); exists {
		userID = uint(uid.(uint))
	}

	list, total, err := service.GetProducts(page, pageSize, keyword, productType, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// 获取单个商品
func GetProduct(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	product, err := service.GetProductByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "商品不存在"})
		return
	}

	// 获取当前用户ID
	var userID uint
	if uid, exists := c.Get("user_id"); exists {
		userID = uint(uid.(uint))
	}

	// 只有已上架(1)或发布者本人才能查看
	if product.Status != 1 && product.Publisher != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "商品不存在"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// 更新商品
func UpdateProduct(c *gin.Context) {
	var product model.Product

	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if err := service.UpdateProduct(&product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "更新成功"})
}

// 删除商品
func DeleteProduct(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	if err := service.DeleteProduct(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "删除成功"})
}

// 获取我发布的商品
func GetMyProducts(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	list, err := service.GetProductsByPublisher(uint(userID.(uint)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"list": list})
}

// 获取上架商品
func GetOnSaleProducts(c *gin.Context) {
	list, err := service.GetProductsByStatus(1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"list": list})
}

// 获取商品（按类型）
func GetProductsByType(c *gin.Context) {
	productType := c.Param("type")

	list, err := service.GetProductsByType(productType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"list": list})
}

// 获取所有商品（管理员，包含下架商品）
func GetAllProductsAdmin(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "100"))
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-999"))

	list, total, err := service.GetAllProductsForAdmin(page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// 获取已退回商品列表（管理员）
func GetRejectedProductsAdmin(c *gin.Context) {
	list, err := service.GetRejectedProductsForAdmin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"list": list})
}

// 获取已隐藏商品列表（管理员）
func GetHiddenProductsAdmin(c *gin.Context) {
	list, err := service.GetHiddenProductsForAdmin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"list": list})
}

// 获取待审核商品列表（管理员）
func GetPendingProductsAdmin(c *gin.Context) {
	list, err := service.GetPendingProductsForAdmin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"list": list})
}

// 审核商品（管理员）
func AuditProduct(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Approved bool `json:"approved"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if err := service.AuditProduct(uint(id), req.Approved); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "审核失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "审核完成"})
}

// 恢复商品（上架）
func RestoreProductAdmin(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := service.RestoreProduct(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "恢复失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "已重新上架"})
}

// 下架商品
func OffShelfProduct(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := service.UpdateProductStatus(uint(id), 0); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "下架失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "已下架"})
}
