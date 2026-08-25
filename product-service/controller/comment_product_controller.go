// controller/comment_product_controller.go
package controller

import (
	"net/http"
	"strconv"

	"rxtcloud/product-service/model"
	"rxtcloud/product-service/service"

	"github.com/gin-gonic/gin"
)

// 创建商品评论
func CreateCommentProduct(c *gin.Context) {
	var comment model.CommentProduct

	println("添加商品评论")
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	comment.UserID = uint(userID.(uint))

	if err := service.CreateCommentProduct(&comment); err != nil {
		println("创建评论失败:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "评论成功"})
}

// 获取商品评论（树结构）
func GetCommentProduct(c *gin.Context) {
	productID, _ := strconv.Atoi(c.Param("product_id"))

	list, err := service.GetCommentProductByProductID(uint(productID))
	if err != nil {
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}

	tree := BuildCommentProductTree(list)

	c.JSON(200, tree)
}

// 更新商品评论
func UpdateCommentProduct(c *gin.Context) {
	var comment model.CommentProduct

	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := service.UpdateCommentProduct(&comment); err != nil {
		c.JSON(500, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(200, gin.H{"msg": "更新成功"})
}

// 删除商品评论
func DeleteCommentProduct(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	if err := service.DeleteCommentProduct(uint(id)); err != nil {
		c.JSON(500, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(200, gin.H{"msg": "删除成功"})
}

// 点赞/取消点赞商品评论
func LikeCommentProduct(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}

	result, err := service.ToggleLike(userID.(uint), "product_comment", uint(id))
	if err != nil {
		c.JSON(500, gin.H{"error": "点赞失败"})
		return
	}

	msg := "点赞成功"
	if !result.Liked {
		msg = "已取消点赞"
	}
	c.JSON(200, gin.H{"msg": msg, "liked": result.Liked, "likes": result.Likes})
}
