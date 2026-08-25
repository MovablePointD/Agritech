package service

import (
	"fmt"
	"rxtcloud/common/config"
	"rxtcloud/common/model"
	"time"
)

// createAffairNotification 创建事务相关通知（异步，不阻塞主流程）
func createAffairNotification(userID uint, notifType, title, content string, affairID uint, uniqueKey string) {
	go func() {
		// 去重检查
		if uniqueKey != "" {
			var existing model.Notification
			if err := config.DB.Where("unique_key = ?", uniqueKey).First(&existing).Error; err == nil {
				return // 已存在，不重复创建
			}
		}

		notification := model.Notification{
			UserID:      userID,
			Type:        notifType,
			Title:       title,
			Content:     content,
			RelatedID:   affairID,
			RelatedType: "affair",
			IsRead:      0,
			UniqueKey:   uniqueKey,
			CreatedAt:   time.Now(),
		}
		config.DB.Create(&notification)
	}()
}

// ============================================================
// 事务审核通知
// ============================================================

// NotifyAffairApproved 通知事务提交者：事务审核通过
func NotifyAffairApproved(affair *model.RuralAffair) {
	key := fmt.Sprintf("affair_approved_%d", affair.ID)
	content := fmt.Sprintf("您提交的事务「%s」已审核通过，等待事务处理人员接取。", affair.Title)
	createAffairNotification(affair.UserID, model.NotificationTypeAffairApproved, "事务审核通过", content, affair.ID, key)
}

// NotifyAffairRejected 通知事务提交者：事务审核不通过
func NotifyAffairRejected(affair *model.RuralAffair) {
	key := fmt.Sprintf("affair_rejected_%d", affair.ID)
	reason := ""
	if affair.RejectReason != "" {
		reason = fmt.Sprintf("，原因：%s", affair.RejectReason)
	}
	content := fmt.Sprintf("您提交的事务「%s」审核未通过%s。", affair.Title, reason)
	createAffairNotification(affair.UserID, model.NotificationTypeAffairApproved, "事务审核未通过", content, affair.ID, key)
}

// ============================================================
// 事务接取通知
// ============================================================

// NotifyAffairAccepted 通知事务提交者：事务已被处理人员接取
func NotifyAffairAccepted(affair *model.RuralAffair) {
	key := fmt.Sprintf("affair_accepted_%d", affair.ID)
	content := fmt.Sprintf("您的事务「%s」已被%s接取，正在处理中。", affair.Title, affair.HandlerName)
	createAffairNotification(affair.UserID, model.NotificationTypeAffairAccepted, "事务已被接取", content, affair.ID, key)
}

// ============================================================
// 处理报告通知
// ============================================================

// NotifyAffairProcessed 通知事务提交者：处理人员已提交处理报告
func NotifyAffairProcessed(affair *model.RuralAffair) {
	content := fmt.Sprintf("事务处理人员%s已针对您的事务「%s」提交了处理报告，请查看并确认完成。", affair.HandlerName, affair.Title)
	createAffairNotification(affair.UserID, model.NotificationTypeAffairProcessed, "处理报告已提交", content, affair.ID, "")
}

// NotifyAffairFollowUpAnswered 通知事务提交者：处理人员已回复追问
func NotifyAffairFollowUpAnswered(affair *model.RuralAffair) {
	content := fmt.Sprintf("事务处理人员%s已对您的事务「%s」的追问进行了回复，请查看。", affair.HandlerName, affair.Title)
	createAffairNotification(affair.UserID, model.NotificationTypeAffairProcessed, "追问已回复", content, affair.ID, "")
}

// ============================================================
// 事务完成通知
// ============================================================

// NotifyAffairCompleted 通知事务提交者和处理人员：事务已完成
func NotifyAffairCompleted(affair *model.RuralAffair) {
	// 通知事务提交者
	keyUser := fmt.Sprintf("affair_completed_user_%d", affair.ID)
	contentUser := fmt.Sprintf("您的事务「%s」已完成，感谢您的使用。", affair.Title)
	createAffairNotification(affair.UserID, model.NotificationTypeAffairCompleted, "事务已完成", contentUser, affair.ID, keyUser)

	// 通知事务处理人员
	if affair.HandlerID > 0 {
		keyHandler := fmt.Sprintf("affair_completed_handler_%d", affair.ID)
		contentHandler := fmt.Sprintf("您处理的事务「%s」已被用户确认完成。", affair.Title)
		createAffairNotification(affair.HandlerID, model.NotificationTypeAffairCompleted, "事务已完成", contentHandler, affair.ID, keyHandler)
	}
}

// ============================================================
// 追问通知
// ============================================================

// NotifyAffairFollowUp 通知事务处理人员：事务被用户追问
func NotifyAffairFollowUp(affair *model.RuralAffair) {
	if affair.HandlerID == 0 {
		return
	}
	content := fmt.Sprintf("您处理的事务「%s」收到用户的追问，请及时查收并回复。", affair.Title)
	createAffairNotification(affair.HandlerID, model.NotificationTypeAffairFollowUp, "事务收到追问", content, affair.ID, "")
}

// ============================================================
// 申诉通知
// ============================================================

// NotifyAffairAppeal 通知所有 admin 和 sysadmin：有新的申诉需要处理
func NotifyAffairAppeal(appeal *model.AffairAppeal, affair *model.RuralAffair) {
	// 查找所有拥有 admin 或 sysadmin 角色的用户
	var admins []model.User
	config.DB.Where("role IN ?", []string{"admin", "sysadmin"}).Find(&admins)

	applicantLabel := "用户"
	if appeal.ApplicantType == "handler" {
		applicantLabel = "事务处理人员"
	}

	for _, admin := range admins {
		content := fmt.Sprintf("%s%s对事务「%s」发起了申诉，请及时处理。申诉理由：%s",
			applicantLabel, appeal.ApplicantName, affair.Title, appeal.Reason)
		createAffairNotification(admin.ID, model.NotificationTypeAffairAppeal, "收到新申诉", content, affair.ID, "")
	}
}
