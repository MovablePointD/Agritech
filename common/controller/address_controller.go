package controller

import (
	"net/http"
	"strconv"

	"rxtcloud/common/model"
	"rxtcloud/common/service"

	"github.com/gin-gonic/gin"
)

// 创建地址
func CreateAddress(c *gin.Context) {
	var address model.Address

	if err := c.ShouldBindJSON(&address); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	address.UserID = uint(userID.(uint))

	if err := service.CreateAddress(&address); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "创建成功", "data": address})
}

// 获取我的地址列表
func GetMyAddresses(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	list, err := service.GetAddressesByUserID(uint(userID.(uint)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"list": list})
}

// 获取单个地址
func GetAddress(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	address, err := service.GetAddressByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "地址不存在"})
		return
	}

	c.JSON(http.StatusOK, address)
}

// 更新地址
func UpdateAddress(c *gin.Context) {
	var address model.Address

	if err := c.ShouldBindJSON(&address); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	userID, _ := c.Get("user_id")
	existing, _ := service.GetAddressByID(address.ID)

	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "地址不存在"})
		return
	}

	if existing.UserID != uint(userID.(uint)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作"})
		return
	}

	if err := service.UpdateAddress(&address); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "更新成功"})
}

// 删除地址
func DeleteAddress(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	address, _ := service.GetAddressByID(uint(id))
	if address == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "地址不存在"})
		return
	}

	userID, _ := c.Get("user_id")
	if address.UserID != uint(userID.(uint)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作"})
		return
	}

	if err := service.DeleteAddress(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "删除成功"})
}

// 设置默认地址
func SetDefaultAddress(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	address, _ := service.GetAddressByID(uint(id))
	if address == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "地址不存在"})
		return
	}

	userID, _ := c.Get("user_id")
	if address.UserID != uint(userID.(uint)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作"})
		return
	}

	if err := service.SetDefaultAddress(uint(userID.(uint)), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "设置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "设置成功"})
}

// 获取默认地址
func GetDefaultAddress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	address, err := service.GetDefaultAddress(uint(userID.(uint)))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "暂无默认地址"})
		return
	}

	c.JSON(http.StatusOK, address)
}
