package controller

import (
	"errors"
	"net/http"
	"rxtcloud/product-service/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CartController struct{}

var CartService = service.CartServiceInstance

func getUserIDFromContext(ctx *gin.Context) (uint, error) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		return 0, errors.New("未登录")
	}
	return uint(userID.(uint)), nil
}

// GetMyCart 获取我的购物车列表
func (c *CartController) GetMyCart(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	items, err := CartService.GetMyCartItems(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"list":  items,
		"total": len(items),
	})
}

// GetCartItem 获取单个购物车项
func (c *CartController) GetCartItem(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	item, err := CartService.GetCartItem(uint(id), userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, item)
}

// AddToCart 添加到购物车
func (c *CartController) AddToCart(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		ProductID uint `json:"product_id" binding:"required"`
		Quantity  int  `json:"quantity" binding:"required,min=1"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	item, err := CartService.AddToCart(userID, req.ProductID, req.Quantity)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, item)
}

// UpdateCartItem 更新购物车项
func (c *CartController) UpdateCartItem(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req struct {
		Quantity int `json:"quantity" binding:"required,min=1"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	item, err := CartService.UpdateCartItem(uint(id), userID, req.Quantity)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, item)
}

// RemoveFromCart 从购物车移除
func (c *CartController) RemoveFromCart(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := CartService.RemoveFromCart(uint(id), userID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ClearCart 清空购物车
func (c *CartController) ClearCart(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if err := CartService.ClearCart(userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "清空成功"})
}

// GetCartCount 获取购物车数量
func (c *CartController) GetCartCount(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	count, err := CartService.GetCartCount(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"count": count})
}

// BatchAddToCart 批量添加商品到购物车
func (c *CartController) BatchAddToCart(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		Items []struct {
			ProductID uint `json:"product_id" binding:"required"`
			Quantity  int  `json:"quantity" binding:"required,min=1"`
		} `json:"items" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 转换�?service 需要的类型
	serviceItems := make([]struct {
		ProductID uint `json:"product_id"`
		Quantity  int  `json:"quantity"`
	}, len(req.Items))
	for i, item := range req.Items {
		serviceItems[i] = struct {
			ProductID uint `json:"product_id"`
			Quantity  int  `json:"quantity"`
		}{ProductID: item.ProductID, Quantity: item.Quantity}
	}

	items, err := CartService.BatchAddToCart(userID, serviceItems)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"list": items})
}

// CheckoutCart 结算购物车（将选中商品转为订单，每个商品单独创建一份订单）
func (c *CartController) CheckoutCart(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		ItemIDs   []uint `json:"item_ids" binding:"required"` // 要结算的购物车项ID
		AddressID uint   `json:"address_id" binding:"required"`
		Remark    string `json:"remark"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 结算购物车并创建订单（返回订单列表）
	orders, err := CartService.CheckoutCart(userID, req.ItemIDs, req.AddressID, req.Remark)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "结算成功",
		"orders":  orders,
	})
}

// DirectBuy 直接购买单个商品（不加入购物车）
func (c *CartController) DirectBuy(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		ProductID uint   `json:"product_id" binding:"required"`
		Quantity  int    `json:"quantity" binding:"required,min=1"`
		AddressID uint   `json:"address_id" binding:"required"`
		Remark    string `json:"remark"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	order, err := CartService.DirectBuy(userID, req.ProductID, req.Quantity, req.AddressID, req.Remark)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "购买成功",
		"order":   order,
	})
}
