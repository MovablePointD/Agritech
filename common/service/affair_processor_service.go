// Package service 提供事务处理人员申请、审核、事务推荐服务
//
// 申请状态：0-待审核  1-审核通过  2-已拒绝/已禁用
// 审核通过时自动更新用户角色为 processor
// 事务推荐：基于处理人员的负责地址关键词匹配，同地址事务优先展示
package service

import (
	"errors"
	"rxtcloud/common/config"
	"rxtcloud/common/model"
	"strings"
	"time"
)

// ApplyProcessor 申请成为事务处理人员，已申请(0)→提示等待审核，已通过(1)→提示无需重复，已拒绝(2)→更新后重新申请
func ApplyProcessor(processor *model.AffairProcessor) error {
	// 检查是否已有申请记录
	var existing model.AffairProcessor
	err := config.DB.Where("user_id = ?", processor.UserID).First(&existing).Error
	if err == nil {
		// 已有申请，检查状态
		if existing.Status == 0 {
			return errors.New("您已提交过申请，请等待审核")
		}
		if existing.Status == 1 {
			return errors.New("您已是事务处理人员，无需重复申请")
		}
		// 之前被拒绝，更新申请
		existing.Status = 0
		existing.RealName = processor.RealName
		existing.Phone = processor.Phone
		existing.Address = processor.Address
		existing.Intro = processor.Intro
		existing.CertImages = processor.CertImages
		existing.RejectReason = ""
		return config.DB.Save(&existing).Error
	}

	return config.DB.Create(processor).Error
}

// GetMyProcessor 获取我的处理人员申请
func GetMyProcessor(userID uint) (*model.AffairProcessor, error) {
	var processor model.AffairProcessor
	err := config.DB.Preload("User").Where("user_id = ?", userID).First(&processor).Error
	if err != nil {
		return nil, err
	}
	return &processor, nil
}

// IsProcessor 检查用户是否为事务处理人员（processor或sysadmin角色）
func IsProcessor(userID uint) bool {
	var user model.User
	err := config.DB.First(&user, userID).Error
	if err != nil {
		return false
	}
	return user.Role == "processor" || user.Role == "sysadmin" || user.Role == "admin"
}

// GetPendingProcessors 获取待审核的处理人员申请
func GetPendingProcessors(page, pageSize int) ([]model.AffairProcessor, int64, error) {
	var processors []model.AffairProcessor
	var total int64

	query := config.DB.Model(&model.AffairProcessor{}).Where("status = ?", 0)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("User").
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&processors).Error

	return processors, total, err
}

// ApproveProcessor 批准处理人员申请
func ApproveProcessor(id uint, auditorName string) error {
	var processor model.AffairProcessor
	if err := config.DB.First(&processor, id).Error; err != nil {
		return errors.New("申请不存在")
	}

	if processor.Status != 0 {
		return errors.New("该申请已被处理")
	}

	// 更新申请状态
	now := time.Now()
	processor.Status = 1
	processor.AuditTime = &now
	processor.AuditName = auditorName

	// 更新用户角色
	if err := config.DB.Model(&model.User{}).Where("id = ?", processor.UserID).Update("role", "processor").Error; err != nil {
		return errors.New("更新用户角色失败")
	}

	return config.DB.Save(&processor).Error
}

// RejectProcessor 拒绝处理人员申请
func RejectProcessor(id uint, auditorName string, reason string) error {
	var processor model.AffairProcessor
	if err := config.DB.First(&processor, id).Error; err != nil {
		return errors.New("申请不存在")
	}

	if processor.Status != 0 {
		return errors.New("该申请已被处理")
	}

	now := time.Now()
	processor.Status = 2
	processor.AuditTime = &now
	processor.AuditName = auditorName
	processor.RejectReason = reason

	return config.DB.Save(&processor).Error
}

// GetAllProcessors 获取所有处理人员申请（包含所有状态）
func GetAllProcessors(page, pageSize int) ([]model.AffairProcessor, int64, error) {
	var processors []model.AffairProcessor
	var total int64

	query := config.DB.Model(&model.AffairProcessor{})
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("User").
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&processors).Error

	return processors, total, err
}

// DisableProcessor 禁用处理人员
func DisableProcessor(id uint) error {
	var processor model.AffairProcessor
	if err := config.DB.First(&processor, id).Error; err != nil {
		return errors.New("申请不存在")
	}

	if processor.Status != 1 {
		return errors.New("只能禁用已通过的处理人员")
	}

	// 更新申请状态为已拒绝（禁用）
	processor.Status = 2

	// 更新用户角色为普通用户
	if err := config.DB.Model(&model.User{}).Where("id = ?", processor.UserID).Update("role", "normal").Error; err != nil {
		return errors.New("更新用户角色失败")
	}

	return config.DB.Save(&processor).Error
}

// GetRecommendedAffairs 为处理人员推荐事务：基于负责地址关键词匹配
// admin/sysadmin 返回所有待处理事务；普通处理人员按地址匹配度排序（匹配的在前）
func GetRecommendedAffairs(userID uint, page, pageSize int) ([]model.RuralAffair, int64, error) {
	var affairs []model.RuralAffair
	var total int64

	// 检查用户角色
	var user model.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return nil, 0, errors.New("用户不存在")
	}

	// sysadmin和admin不需要推荐，直接获取所有待处理事务
	if user.Role == "sysadmin" || user.Role == "admin" {
		return GetPendingProcessAffairs(page, pageSize)
	}

	// 获取用户的负责区域
	var processor model.AffairProcessor
	if err := config.DB.Where("user_id = ? AND status = ?", userID, 1).First(&processor).Error; err != nil {
		// 用户不是处理人员，返回普通待处理列表
		return GetPendingProcessAffairs(page, pageSize)
	}

	userAddress := processor.Address

	// 查询待处理事务（状态为2：审核通过）
	query := config.DB.Model(&model.RuralAffair{}).Where("status = ?", 2)
	query.Count(&total)

	offset := (page - 1) * pageSize

	// 如果有地址信息，按地址相似度排序
	if userAddress != "" {
		// 使用地址关键词匹配
		keywords := extractKeywords(userAddress)
		if len(keywords) > 0 {
			// 创建临时表存储匹配度
			var matchedAffairs []model.RuralAffair
			var unmatchedAffairs []model.RuralAffair

			err := query.Preload("User").Order("created_at asc").Find(&affairs).Error
			if err != nil {
				return nil, 0, err
			}

			// 分离匹配和不匹配的事务
			for _, affair := range affairs {
				if matchAddress(affair.Address, keywords) {
					matchedAffairs = append(matchedAffairs, affair)
				} else {
					unmatchedAffairs = append(unmatchedAffairs, affair)
				}
			}

			// 合并结果：匹配的在前，按创建时间排序；不匹配的按创建时间由老到新排序
			affairs = append(matchedAffairs, unmatchedAffairs...)

			// 分页
			total = int64(len(affairs))
			start := offset
			end := offset + pageSize
			if start > len(affairs) {
				return []model.RuralAffair{}, total, nil
			}
			if end > len(affairs) {
				end = len(affairs)
			}
			affairs = affairs[start:end]

			return affairs, total, nil
		}
	}

	// 默认按创建时间由老到新排序
	err := query.Preload("User").
		Order("created_at asc").
		Offset(offset).
		Limit(pageSize).
		Find(&affairs).Error

	return affairs, total, err
}

// extractKeywords 提取地址关键词
func extractKeywords(address string) []string {
	var keywords []string
	// 简单按常见分隔符分割
	splitChars := []string{"省", "市", "区", "县", "镇", "乡", "村", "路", "街", "道"}
	temp := address
	for _, char := range splitChars {
		parts := strings.Split(temp, char)
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if len(part) >= 2 {
				keywords = append(keywords, part)
			}
		}
		temp = strings.Join(parts, "")
	}
	return keywords
}

// matchAddress 检查地址是否匹配关键词
func matchAddress(address string, keywords []string) bool {
	if address == "" {
		return false
	}
	for _, keyword := range keywords {
		if len(keyword) >= 2 && strings.Contains(address, keyword) {
			return true
		}
	}
	return false
}

// GetMyHandledAffairs 获取我处理过的事务
func GetMyHandledAffairs(userID uint, page, pageSize int) ([]model.RuralAffair, int64, error) {
	var affairs []model.RuralAffair
	var total int64

	query := config.DB.Model(&model.RuralAffair{}).Where("handler_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("User").
		Order("completed_time desc").
		Offset(offset).
		Limit(pageSize).
		Find(&affairs).Error

	return affairs, total, err
}
