// Package service 提供农村事务核心业务逻辑
// RuralAffair 状态流转说明：
//   1 - 待审核：用户提交事务后初始状态
//   2 - 审核通过（待处理）：审核通过，等待处理人员接取
//   3 - 审核不通过：被审核人员驳回
//   4 - 处理中：处理人员已接取并提交处理记录
//   5 - 已完成：用户确认事务处理完成
//   6 - 追问中：用户对处理结果有疑问，发起追问
//   7 - 申诉中：用户或处理人员发起申诉
package service

import (
	"errors"
	"rxtcloud/common/config"
	"rxtcloud/common/model"
	"time"
)

// CreateRuralAffair 创建农村事务，status 默认为 1（待审核）
func CreateRuralAffair(affair *model.RuralAffair) error {
	return config.DB.Create(affair).Error
}

// GetRuralAffairs 分页查询事务列表，支持按状态、类型、关键词筛选
// keyword 支持模糊匹配 title 和 content 字段
func GetRuralAffairs(page, pageSize int, status string, affairType string, keyword string) ([]model.RuralAffair, int64, error) {
	var affairs []model.RuralAffair
	var total int64

	query := config.DB.Model(&model.RuralAffair{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if affairType != "" {
		query = query.Where("type = ?", affairType)
	}
	if keyword != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("User").
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&affairs).Error

	return affairs, total, err
}

// GetRuralAffair 根据 ID 获取单个事务详情，预加载关联用户信息
func GetRuralAffair(id uint) (*model.RuralAffair, error) {
	var affair model.RuralAffair
	err := config.DB.Preload("User").First(&affair, id).Error
	if err != nil {
		return nil, err
	}
	return &affair, nil
}

// GetUserRuralAffairs 获取当前用户提交的所有事务（本人的事务列表）
func GetUserRuralAffairs(userID uint, page, pageSize int) ([]model.RuralAffair, int64, error) {
	var affairs []model.RuralAffair
	var total int64

	query := config.DB.Model(&model.RuralAffair{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("User").
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&affairs).Error

	return affairs, total, err
}

func UpdateRuralAffair(affair *model.RuralAffair) error {
	return config.DB.Save(affair).Error
}

func DeleteRuralAffair(id uint) error {
	return config.DB.Delete(&model.RuralAffair{}, id).Error
}

// 获取待审核的事务列表（供审核人员使用）
func GetPendingAuditAffairs(page, pageSize int) ([]model.RuralAffair, int64, error) {
	var affairs []model.RuralAffair
	var total int64

	query := config.DB.Model(&model.RuralAffair{}).Where("status = ?", 1)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("User").
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&affairs).Error

	return affairs, total, err
}

// 审核事务
func AuditRuralAffair(id uint, approved bool, auditorName string, rejectReason string) error {
	var affair model.RuralAffair
	if err := config.DB.First(&affair, id).Error; err != nil {
		return err
	}

	now := time.Now()
	if approved {
		affair.Status = 2 // 审核通过
	} else {
		affair.Status = 3 // 审核不通过
		affair.RejectReason = rejectReason
	}
	affair.AuditTime = &now
	affair.AuditName = auditorName

	// 发送通知
	if approved {
		NotifyAffairApproved(&affair)
	} else {
		NotifyAffairRejected(&affair)
	}

	return config.DB.Save(&affair).Error
}

// 获取待处理的事务列表（供处理人员使用）
func GetPendingProcessAffairs(page, pageSize int) ([]model.RuralAffair, int64, error) {
	var affairs []model.RuralAffair
	var total int64

	query := config.DB.Model(&model.RuralAffair{}).Where("status = ?", 2)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("User").
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&affairs).Error

	return affairs, total, err
}

// 处理事务 — 提交处理记录（首次提交后不直接完成，等待用户确认或追问）
func ProcessRuralAffair(id uint, handlerID uint, handlerName string, processContent string, processImages string) error {
	var affair model.RuralAffair
	if err := config.DB.First(&affair, id).Error; err != nil {
		return errors.New("事务不存在")
	}

	if affair.Status != 2 && affair.Status != 4 && affair.Status != 6 {
		return errors.New("该事务当前状态不允许处理")
	}

	now := time.Now()
	affair.HandlerID = handlerID
	affair.HandlerName = handlerName
	affair.ProcessContent = processContent
	affair.ProcessImages = processImages

	// 首次提交处理记录时记录时间（用于判断处理人员申诉条件）
	if affair.FirstResponseAt == nil {
		affair.FirstResponseAt = &now
	}

	// 保留在原状态（处理中4），由用户确认完成或追问
	// 如果是从审核通过(2)来的，转为处理中(4)
	if affair.Status == 2 {
		affair.Status = 4
		affair.ProcessTime = &now
	}
	// 追问中(6)状态下追答 → 已在 AddFollowUpAnswer 中处理状态切换
	// 这里统一保留为处理中，让用户确认完成
	if affair.Status == 6 {
		affair.Status = 4
	}

	if err := config.DB.Save(&affair).Error; err != nil {
		return err
	}

	// 通知事务提交者：处理报告已提交
	NotifyAffairProcessed(&affair)

	return nil
}

// 开始处理（从审核通过转为处理中）
func StartProcessRuralAffair(id uint, handlerID uint, handlerName string) error {
	var affair model.RuralAffair
	if err := config.DB.First(&affair, id).Error; err != nil {
		return errors.New("事务不存在")
	}

	if affair.Status != 2 {
		return errors.New("该事务当前状态不允许开始处理")
	}

	now := time.Now()
	affair.Status = 4 // 处理中
	affair.HandlerID = handlerID
	affair.HandlerName = handlerName
	affair.ProcessTime = &now

	if err := config.DB.Save(&affair).Error; err != nil {
		return err
	}

	// 通知事务提交者：事务已被接取
	NotifyAffairAccepted(&affair)

	return nil
}

// StartProcessRuralAffairWithCheck 带权限检查的开始处理
func StartProcessRuralAffairWithCheck(id uint, handlerID uint, handlerName string, userRole string) error {
	// 检查权限：只有processor或sysadmin角色可以处理
	if userRole != "processor" && userRole != "sysadmin" && userRole != "admin" {
		return errors.New("您没有处理权限")
	}
	return StartProcessRuralAffair(id, handlerID, handlerName)
}

// ProcessRuralAffairWithCheck 带权限检查的处理完成
func ProcessRuralAffairWithCheck(id uint, handlerID uint, handlerName string, processContent string, processImages string, userRole string) error {
	// 检查权限：只有processor或sysadmin角色可以处理
	if userRole != "processor" && userRole != "sysadmin" && userRole != "admin" {
		return errors.New("您没有处理权限")
	}
	return ProcessRuralAffair(id, handlerID, handlerName, processContent, processImages)
}

// 获取所有事务（管理员）
func GetAllRuralAffairs(page, pageSize int, status string) ([]model.RuralAffair, int64, error) {
	var affairs []model.RuralAffair
	var total int64

	query := config.DB.Model(&model.RuralAffair{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("User").
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&affairs).Error

	return affairs, total, err
}

// 获取审核不通过的事务（status=3）
func GetRejectedAffairs(page, pageSize int) ([]model.RuralAffair, int64, error) {
	var affairs []model.RuralAffair
	var total int64

	query := config.DB.Model(&model.RuralAffair{}).Where("status = ?", 3)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("User").
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&affairs).Error

	return affairs, total, err
}

// 获取处理中的事务（status=4）
func GetProcessingAffairs(page, pageSize int) ([]model.RuralAffair, int64, error) {
	var affairs []model.RuralAffair
	var total int64

	query := config.DB.Model(&model.RuralAffair{}).Where("status = ?", 4)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("User").
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&affairs).Error

	return affairs, total, err
}

// 获取已完成的事务（status=5）
func GetCompletedAffairs(page, pageSize int) ([]model.RuralAffair, int64, error) {
	var affairs []model.RuralAffair
	var total int64

	query := config.DB.Model(&model.RuralAffair{}).Where("status = ?", 5)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("User").
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&affairs).Error

	return affairs, total, err
}
