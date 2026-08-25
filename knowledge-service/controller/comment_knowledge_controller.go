// controller/comment_knowledge_controller.go
package controller

import (
	"net/http"
	"strconv"

	"rxtcloud/knowledge-service/model"
	"rxtcloud/knowledge-service/service"

	"github.com/gin-gonic/gin"
)

// 创建知识评论
func CreateCommentKnowledge(c *gin.Context) {
	var comment model.CommentKnowledge

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

	if err := service.CreateCommentKnowledge(&comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "评论成功"})
}

// 获取知识评论（树结构�?
func GetCommentKnowledge(c *gin.Context) {
	knowledgeID, _ := strconv.Atoi(c.Param("knowledge_id"))

	list, err := service.GetCommentKnowledgeByKnowledgeID(uint(knowledgeID))
	if err != nil {
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}

	tree := BuildCommentKnowledgeTree(list)

	c.JSON(200, tree)
}

// 更新知识评论
func UpdateCommentKnowledge(c *gin.Context) {
	var comment model.CommentKnowledge

	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := service.UpdateCommentKnowledge(&comment); err != nil {
		c.JSON(500, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(200, gin.H{"msg": "更新成功"})
}

// 删除知识评论
func DeleteCommentKnowledge(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	if err := service.DeleteCommentKnowledge(uint(id)); err != nil {
		c.JSON(500, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(200, gin.H{"msg": "删除成功"})
}

// 点赞/取消点赞知识评论
func LikeCommentKnowledge(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}

	result, err := service.ToggleLike(uint(userID.(uint)), "knowledge_comment", uint(id))
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
