package controller

import (
	"net/http"
	"rxtcloud/common/config"
	"rxtcloud/common/model"
	"rxtcloud/common/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetCommentRuralInfos 获取评论列表（游客可访问）
func GetCommentRuralInfos(c *gin.Context) {
	targetType := c.Query("target_type") // "rural_info" | "policy_notice"
	targetID, err := strconv.ParseUint(c.Query("target_id"), 10, 32)
	if err != nil || targetType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	list, err := service.GetCommentRuralInfos(targetType, uint(targetID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败"})
		return
	}

	// 构建评论树
	tree := buildCommentTree(list)
	c.JSON(http.StatusOK, gin.H{"list": tree})
}

type commentNode struct {
	model.CommentRuralInfo
	Children []*commentNode `json:"children,omitempty"`
}

func buildCommentTree(list []model.CommentRuralInfo) []*commentNode {
	nodes := make(map[uint]*commentNode)
	var roots []*commentNode

	for i := range list {
		node := &commentNode{CommentRuralInfo: list[i]}
		nodes[list[i].ID] = node
	}

	for _, node := range nodes {
		if node.ParentID == 0 {
			roots = append(roots, node)
		} else if parent, ok := nodes[node.ParentID]; ok {
			parent.Children = append(parent.Children, node)
		}
	}
	return roots
}

// CreateCommentRuralInfo 创建评论（需登录）
func CreateCommentRuralInfo(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	var comment model.CommentRuralInfo
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	comment.UserID = userID

	// 自动计算层级
	if comment.ParentID > 0 {
		var parent model.CommentRuralInfo
		if err := config.DB.First(&parent, comment.ParentID).Error; err == nil {
			comment.Level = parent.Level + 1
			if comment.Level > 3 {
				comment.Level = 3
			}
		} else {
			comment.Level = 1
		}
	} else {
		comment.Level = 1
	}

	if err := service.CreateCommentRuralInfo(&comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "评论失败"})
		return
	}

	// 异步通知
	go func() {
		var targetUserID uint
		if comment.ParentID > 0 {
			var p model.CommentRuralInfo
			if config.DB.First(&p, comment.ParentID).Error == nil {
				targetUserID = p.UserID
			}
		}
		if targetUserID > 0 && targetUserID != userID {
			uniqueKey := comment.TargetType + "_comment_" + strconv.FormatUint(uint64(comment.ID), 10)
			_ = service.CreateCommentNotification(targetUserID, "comment", "评论通知", comment.Content, comment.TargetID, comment.TargetType, userID, uniqueKey)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"msg": "评论成功", "data": comment})
}

// UpdateCommentRuralInfo 更新评论（仅本人）
func UpdateCommentRuralInfo(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	var req struct {
		ID      uint   `json:"id"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if err := service.UpdateCommentRuralInfo(req.ID, userID, req.Content); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权修改"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "修改成功"})
}

// DeleteCommentRuralInfo 删除评论（本人或管理员）
func DeleteCommentRuralInfo(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if err := service.DeleteCommentRuralInfo(uint(id), userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权删除"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "删除成功"})
}

// LikeCommentRuralInfo 点赞/取消点赞
func LikeCommentRuralInfo(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	result, err := service.ToggleLikeCommentRuralInfo(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"liked": result.Liked, "likes": result.Likes})
}
