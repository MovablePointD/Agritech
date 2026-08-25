package controller

import (
	"net/http"
	"strconv"

	"rxtcloud/message-service/model"
	"rxtcloud/message-service/service"

	"github.com/gin-gonic/gin"
)

// GetConversations 获取我的会话列表
func GetConversations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	convs, err := service.GetConversations(uint(userID.(uint)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取会话列表失败"})
		return
	}

	// 构建响应（注入当前用户的未读数）
	type convResp struct {
		ID               uint            `json:"id"`
		OtherUser        *model.UserInfo `json:"other_user"`
		LastMessage      string          `json:"last_message"`
		LastMessageAt    string          `json:"last_message_at"`
		UnreadCount      int             `json:"unread_count"`
		ConversationType string          `json:"conversation_type"`
		ExpertName       string          `json:"expert_name"`
		UpdatedAt        string          `json:"updated_at"`
	}

	uid := uint(userID.(uint))
	var result []convResp
	for _, conv := range convs {
		unread := conv.User1Unread
		if conv.User1ID == uid {
			unread = conv.User1Unread
		} else {
			unread = conv.User2Unread
		}

		result = append(result, convResp{
			ID:               conv.ID,
			OtherUser:        conv.OtherUser,
			LastMessage:      conv.LastMessage,
			LastMessageAt:    conv.LastMessageAt.Format("2006-01-02 15:04:05"),
			UnreadCount:      unread,
			ConversationType: conv.ConversationType,
			ExpertName:       conv.ExpertName,
			UpdatedAt:        conv.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{"conversations": result})
}

// CreateConversation 创建或获取与指定用户的会�?
func CreateConversation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	var req struct {
		TargetUserID uint `json:"target_user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供目标用户ID"})
		return
	}

	conv, err := service.CreateOrGetConversation(uint(userID.(uint)), req.TargetUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"conversation_id": conv.ID})
}

// GetMessages 获取会话消息列表
func GetMessages(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	convID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的会话ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	msgs, total, err := service.GetMessages(uint(convID), uint(userID.(uint)), page, pageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages":  msgs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// SendMessage 发送消息
func SendMessage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	var req struct {
		ConversationID uint   `json:"conversation_id" binding:"required"`
		Content        string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数不完整"})
		return
	}

	msg, err := service.SendMessage(req.ConversationID, uint(userID.(uint)), req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": msg, "msg": "发送成功"})
}

// MarkMessageAsRead 标记会话已读
func MarkMessageAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	convID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的会话ID"})
		return
	}

	if err := service.MarkAsRead(uint(convID), uint(userID.(uint))); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "已标记已读"})
}

// GetConversationDetail 获取会话详情（含对方信息）
func GetConversationDetail(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	convID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的会话ID"})
		return
	}

	conv, err := service.GetConversationByID(uint(convID), uint(userID.(uint)))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	unread := conv.User1Unread
	if conv.User2ID == uint(userID.(uint)) {
		unread = conv.User2Unread
	} else {
		unread = conv.User1Unread
	}

	c.JSON(http.StatusOK, gin.H{
		"conversation": gin.H{
			"id":                conv.ID,
			"other_user":        conv.OtherUser,
			"last_message":      conv.LastMessage,
			"last_message_at":   conv.LastMessageAt,
			"unread_count":      unread,
			"conversation_type": conv.ConversationType,
			"expert_name":       conv.ExpertName,
		},
	})
}

// GetMessageUnreadCount 获取总未读消息数
func GetMessageUnreadCount(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	count := service.GetUnreadCount(uint(userID.(uint)))
	c.JSON(http.StatusOK, gin.H{"count": count})
}

// DeleteConversation 删除会话
func DeleteConversation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	convID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的会话ID"})
		return
	}

	if err := service.DeleteConversation(uint(convID), uint(userID.(uint))); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "已删除"})
}

// CreateConsultation 发起专家咨询
func CreateConsultation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	var req struct {
		ExpertID uint   `json:"expert_id" binding:"required"`
		Content  string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数不完整，请提供专家ID和咨询内容"})
		return
	}

	conv, _, err := service.CreateConsultation(req.ExpertID, uint(userID.(uint)), req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg":             "咨询已提交，等待专家回复",
		"conversation_id": conv.ID,
	})
}
