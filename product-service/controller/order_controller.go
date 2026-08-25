package controller

import (
	"net/http"
	"strconv"

	commonCtrl "rxtcloud/common/controller"
	"rxtcloud/product-service/model"
	"rxtcloud/product-service/service"

	"github.com/gin-gonic/gin"
)

// 创建订单
func CreateOrder(c *gin.Context) {
	var order model.Order

	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	userID, ok := commonCtrl.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	order.UserID = userID

	if err := service.CreateOrder(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "创建成功", "data": order})
}

// 获取订单列表（分页）
func GetOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	list, total, err := service.GetOrders(page, pageSize)
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

// 获取单个订单
func GetOrder(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	order, err := service.GetOrderByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}

	c.JSON(http.StatusOK, order)
}

// 获取我的订单
func GetMyOrders(c *gin.Context) {
	userID, ok := commonCtrl.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	list, err := service.GetOrdersByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"list": list})
}

// 更新订单（仅限未付款订单）
func UpdateOrder(c *gin.Context) {
	var order model.Order

	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 检查订单状态
	existing, _ := service.GetOrderByID(order.ID)
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}
	if existing.Status != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅可修改待付款订单"})
		return
	}

	if err := service.UpdateOrder(&order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "更新成功"})
}

// 删除订单（仅限已取消订单）
func DeleteOrder(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	order, err := service.GetOrderByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}
	if order.Status != 4 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅可删除已取消订单"})
		return
	}

	if err := service.DeleteOrder(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "删除成功"})
}

// 支付订单
func PayOrder(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	order, _ := service.GetOrderByID(uint(id))
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}
	if order.Status != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "订单状态不允许支付"})
		return
	}

	if err := service.PayOrder(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "支付失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "支付成功"})
}

// 取消订单
func CancelOrder(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	order, _ := service.GetOrderByID(uint(id))
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}
	if order.Status != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅可取消待付款订单"})
		return
	}

	if err := service.CancelOrder(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取消失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "取消成功"})
}

// 确认收货（用户）
func ConfirmReceiveOrder(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	order, _ := service.GetOrderByID(uint(id))
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}
	if order.Status != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅可确认待收货订单"})
		return
	}

	userID, ok := commonCtrl.GetUserID(c)
	if !ok || order.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作"})
		return
	}

	if err := service.ConfirmOrder(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "确认失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "确认收货成功"})
}

// 发货（商家/管理员）
func ShipOrder(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var input struct {
		ExpressNo string `json:"express_no"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	order, _ := service.GetOrderByID(uint(id))
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}
	if order.Status != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅可对待发货订单发货"})
		return
	}

	if err := service.ShipOrder(uint(id), input.ExpressNo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发货失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "发货成功"})
}

// 获取我的销售订单
func GetMySellerOrders(c *gin.Context) {
	userID, ok := commonCtrl.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	list, err := service.GetSellerOrders(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"list": list})
}
