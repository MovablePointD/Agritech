// Package service 提供私信会话与消息核心服务
//
// 会话设计：
//   - user1_id < user2_id 保证两个用户之间只有唯一会话
//   - 支持 normal(普通私信) 和 consultation(专家咨询) 两种类型
//   - user1_unread/user2_unread 分别记录双方未读数
//
// 消息已读标记：MarkAsRead 事务内标记消息+清零未读
package service

import (
	"errors"
	"fmt"
	"time"

	"rxtcloud/message-service/config"
	"rxtcloud/message-service/model"
)

// CreateOrGetConversation 创建或获取两用户间的普通私信会话（user1_id < user2_id 防重复）
func CreateOrGetConversation(userID, targetUserID uint) (*model.Conversation, error) {
	return CreateOrGetConversationWithType(userID, targetUserID, "normal", "")
}

// CreateOrGetConversationWithType 创建或获取指定类型的会话
func CreateOrGetConversationWithType(userID, targetUserID uint, convType string, expertName string) (*model.Conversation, error) {
	if userID == targetUserID {
		return nil, errors.New("不能和自己创建会话")
	}

	// 保证 user1_id < user2_id，避免重复会话
	u1, u2 := userID, targetUserID
	if u1 > u2 {
		u1, u2 = u2, u1
	}

	var conv model.Conversation
	result := config.DB.Where("user1_id = ? AND user2_id = ? AND conversation_type = ?", u1, u2, convType).First(&conv)
	if result.Error == nil {
		// 更新专家姓名（咨询场景）
		if convType == "consultation" && expertName != "" {
			config.DB.Model(&conv).Update("expert_name", expertName)
		}
		return &conv, nil
	}

	// 不存在则创建
	conv = model.Conversation{
		User1ID:          u1,
		User2ID:          u2,
		ConversationType: convType,
		ExpertName:       expertName,
	}
	if err := config.DB.Create(&conv).Error; err != nil {
		return nil, err
	}
	return &conv, nil
}

// CreateConsultation 创建专家咨询：验证专家状态 → 创建 consultation 会话 → 发送首条消息 → 通知专家
func CreateConsultation(expertID uint, userID uint, content string) (*model.Conversation, *model.Message, error) {
	if content == "" {
		return nil, nil, errors.New("咨询内容不能为空")
	}

	// 查找专家信息
	var expert model.Expert
	if err := config.DB.First(&expert, expertID).Error; err != nil {
		return nil, nil, errors.New("专家不存在")
	}
	if expert.Status != 1 {
		return nil, nil, errors.New("该专家尚未通过认证")
	}
	if expert.UserID == userID {
		return nil, nil, errors.New("不能咨询自己")
	}

	// 创建咨询类型的会话
	conv, err := CreateOrGetConversationWithType(userID, expert.UserID, "consultation", expert.RealName)
	if err != nil {
		return nil, nil, err
	}

	// 发送首条咨询消息
	msg := model.Message{
		ConversationID: conv.ID,
		SenderID:       userID,
		Content:        content,
		IsRead:         0,
	}

	tx := config.DB.Begin()

	if err := tx.Create(&msg).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	// 更新会话最后消息
	updates := map[string]interface{}{
		"last_message":    truncateStr(content, 50),
		"last_message_at": time.Now(),
		"updated_at":      time.Now(),
	}

	// 更新专家接收者未读数
	if conv.User1ID == userID {
		updates["user2_unread"] = conv.User2Unread + 1
	} else {
		updates["user1_unread"] = conv.User1Unread + 1
	}

	if err := tx.Model(&conv).Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	tx.Commit()

	// 发送通知给专家
	go func() {
		var sender model.User
		if err := config.DB.First(&sender, userID).Error; err == nil {
			senderName := sender.Nickname
			if senderName == "" {
				senderName = sender.Username
			}
			uniqueKey := fmt.Sprintf("consultation_%d_%d_%d", conv.ID, userID, msg.ID)
			CreateCommentNotification(
				expert.UserID,
				model.NotificationTypeNewMessage,
				"新的专家咨询",
				fmt.Sprintf("%s 向您发起专家咨询：%s", senderName, truncateStr(content, 30)),
				conv.ID,
				"conversation",
				userID,
				uniqueKey,
			)
		}
	}()

	return conv, &msg, nil
}

// GetConversations 获取用户的所有会话列表（含对方用户信息和未读数）
func GetConversations(userID uint) ([]model.Conversation, error) {
	var convs []model.Conversation
	config.DB.Where("user1_id = ? OR user2_id = ?", userID, userID).
		Order("updated_at DESC").
		Find(&convs)

	// 填充对方用户信息
	for i := range convs {
		conv := &convs[i]
		var otherID uint
		if conv.User1ID == userID {
			otherID = conv.User2ID
		} else {
			otherID = conv.User1ID
		}

		var user model.User
		if err := config.DB.First(&user, otherID).Error; err == nil {
			conv.OtherUser = &model.UserInfo{
				ID:        user.ID,
				Username:  user.Username,
				Nickname:  user.Nickname,
				AvatarURL: user.AvatarURL,
			}
		}
	}

	return convs, nil
}

// GetUnreadCount 获取用户总未读消息数
func GetUnreadCount(userID uint) int {
	var total int
	convs, err := GetConversations(userID)
	if err != nil {
		return 0
	}
	for _, c := range convs {
		if c.User1ID == userID {
			total += c.User1Unread
		} else {
			total += c.User2Unread
		}
	}
	return total
}

// GetMessages 获取会话的消息列表（分页）
func GetMessages(conversationID, userID uint, page, pageSize int) ([]model.Message, int64, error) {
	if pageSize <= 0 {
		pageSize = 50
	}
	if page <= 0 {
		page = 1
	}

	var conv model.Conversation
	if err := config.DB.First(&conv, conversationID).Error; err != nil {
		return nil, 0, errors.New("会话不存在")
	}
	if conv.User1ID != userID && conv.User2ID != userID {
		return nil, 0, errors.New("无权访问该会话")
	}

	var total int64
	config.DB.Model(&model.Message{}).Where("conversation_id = ?", conversationID).Count(&total)

	var msgs []model.Message
	offset := (page - 1) * pageSize
	config.DB.Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&msgs)

	return msgs, total, nil
}

// SendMessage 发送私信：事务内创建消息 + 更新会话最后消息+接收者未读数 + 异步通知接收者
func SendMessage(conversationID, senderID uint, content string) (*model.Message, error) {
	if content == "" {
		return nil, errors.New("消息不能为空")
	}

	var conv model.Conversation
	if err := config.DB.First(&conv, conversationID).Error; err != nil {
		return nil, errors.New("会话不存在")
	}
	if conv.User1ID != senderID && conv.User2ID != senderID {
		return nil, errors.New("无权在该会话中发送消息")
	}

	// 确定接收者
	var receiverID uint
	if conv.User1ID == senderID {
		receiverID = conv.User2ID
	} else {
		receiverID = conv.User1ID
	}

	msg := model.Message{
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        content,
		IsRead:         0,
	}

	tx := config.DB.Begin()

	if err := tx.Create(&msg).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 更新会话最后消息
	updates := map[string]interface{}{
		"last_message":    truncateStr(content, 50),
		"last_message_at": time.Now(),
		"updated_at":      time.Now(),
	}
	// 更新接收者未读数
	if conv.User1ID == senderID {
		updates["user2_unread"] = conv.User2Unread + 1
	} else {
		updates["user1_unread"] = conv.User1Unread + 1
	}

	if err := tx.Model(&conv).Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	// 发送通知给接收者
	go func() {
		var sender model.User
		if err := config.DB.First(&sender, senderID).Error; err == nil {
			senderName := sender.Nickname
			if senderName == "" {
				senderName = sender.Username
			}
			uniqueKey := fmt.Sprintf("message_%d_%d_%d", conversationID, senderID, msg.ID)
			CreateCommentNotification(
				receiverID,
				model.NotificationTypeNewMessage,
				"新的私信",
				fmt.Sprintf("%s 给您发来一条私信：%s", senderName, truncateStr(content, 30)),
				conversationID,
				"conversation",
				senderID,
				uniqueKey,
			)
		}
	}()

	return &msg, nil
}

// MarkAsRead 标记已读：事务内将非本人消息置为已读 + 清零当前用户在该会话的未读计数
func MarkAsRead(conversationID, userID uint) error {
	var conv model.Conversation
	if err := config.DB.First(&conv, conversationID).Error; err != nil {
		return errors.New("会话不存在")
	}
	if conv.User1ID != userID && conv.User2ID != userID {
		return errors.New("无权操作")
	}

	tx := config.DB.Begin()

	// 标记消息已读
	if err := tx.Model(&model.Message{}).
		Where("conversation_id = ? AND sender_id != ? AND is_read = 0", conversationID, userID).
		Update("is_read", 1).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 清零未读数
	if conv.User1ID == userID {
		tx.Model(&conv).Update("user1_unread", 0)
	} else {
		tx.Model(&conv).Update("user2_unread", 0)
	}

	tx.Commit()
	return nil
}

// GetConversationByID 获取单个会话详情
func GetConversationByID(conversationID, userID uint) (*model.Conversation, error) {
	var conv model.Conversation
	if err := config.DB.First(&conv, conversationID).Error; err != nil {
		return nil, errors.New("会话不存在")
	}
	if conv.User1ID != userID && conv.User2ID != userID {
		return nil, errors.New("无权访问")
	}

	// 填充对方信息
	var otherID uint
	if conv.User1ID == userID {
		otherID = conv.User2ID
	} else {
		otherID = conv.User1ID
	}

	var user model.User
	if err := config.DB.First(&user, otherID).Error; err == nil {
		conv.OtherUser = &model.UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			Nickname:  user.Nickname,
			AvatarURL: user.AvatarURL,
		}
	}

	return &conv, nil
}

// DeleteConversation 删除会话（软删除）
func DeleteConversation(conversationID, userID uint) error {
	var conv model.Conversation
	if err := config.DB.First(&conv, conversationID).Error; err != nil {
		return errors.New("会话不存在")
	}
	if conv.User1ID != userID && conv.User2ID != userID {
		return errors.New("无权删除")
	}

	tx := config.DB.Begin()
	tx.Where("conversation_id = ?", conversationID).Delete(&model.Message{})
	tx.Delete(&conv)
	tx.Commit()

	return nil
}

func truncateStr(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
