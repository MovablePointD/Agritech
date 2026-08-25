// Package service 提供农村信息(RuralInfo)与政策公告(PolicyNotice)服务
//
// RuralInfo：农村资讯，按 type 分类（新闻/通知/公告），status 控制发布
// PolicyNotice：政策公告，支持置顶(is_top)，按 category 分类
// 地址关联服务：GetAssociatedItemsByAddress 根据地址查询同地址的公告和事务
package service

import (
	"errors"
	"rxtcloud/common/config"
	"rxtcloud/common/model"
	"time"
)

// ==================== RuralInfo 服务 ====================

func CreateRuralInfo(info *model.RuralInfo) error {
	return config.DB.Create(info).Error
}

func GetRuralInfo(id uint) (*model.RuralInfo, error) {
	var info model.RuralInfo
	err := config.DB.Preload("User").First(&info, id).Error
	if err != nil {
		return nil, err
	}
	// 浏览数+1
	config.DB.Model(&info).UpdateColumn("views", info.Views+1)
	info.Views++
	return &info, nil
}

func GetRuralInfoList(page, pageSize int, infoType string, keyword string) ([]model.RuralInfo, int64, error) {
	var list []model.RuralInfo
	var total int64

	query := config.DB.Model(&model.RuralInfo{}).Where("status = ?", 1)
	if infoType != "" {
		query = query.Where("type = ?", infoType)
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
		Find(&list).Error
	return list, total, err
}

func GetAllRuralInfos(page, pageSize int, status string, keyword string) ([]model.RuralInfo, int64, error) {
	var list []model.RuralInfo
	var total int64

	query := config.DB.Model(&model.RuralInfo{})
	if status != "" {
		query = query.Where("status = ?", status)
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
		Find(&list).Error
	return list, total, err
}

func UpdateRuralInfo(info *model.RuralInfo) error {
	return config.DB.Model(&model.RuralInfo{}).Where("id = ?", info.ID).Updates(map[string]interface{}{
		"title":      info.Title,
		"content":    info.Content,
		"type":       info.Type,
		"address":    info.Address,
		"images":     info.Images,
		"status":     info.Status,
		"updated_at": time.Now(),
	}).Error
}

func DeleteRuralInfo(id uint) error {
	return config.DB.Model(&model.RuralInfo{}).Where("id = ?", id).Update("status", 2).Error
}

func HardDeleteRuralInfo(id uint) error {
	return config.DB.Delete(&model.RuralInfo{}, id).Error
}

func AuditRuralInfo(id uint, status int, rejectReason string) error {
	if status != 1 && status != 0 {
		return errors.New("无效的审核状态")
	}
	return config.DB.Model(&model.RuralInfo{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}).Error
}

// ==================== PolicyNotice 服务 ====================

func CreatePolicyNotice(notice *model.PolicyNotice) error {
	return config.DB.Create(notice).Error
}

func GetPolicyNotice(id uint) (*model.PolicyNotice, error) {
	var notice model.PolicyNotice
	err := config.DB.Preload("User").First(&notice, id).Error
	if err != nil {
		return nil, err
	}
	config.DB.Model(&notice).UpdateColumn("views", notice.Views+1)
	notice.Views++
	return &notice, nil
}

func GetPolicyNoticeList(page, pageSize int, category string, keyword string) ([]model.PolicyNotice, int64, error) {
	var list []model.PolicyNotice
	var total int64

	query := config.DB.Model(&model.PolicyNotice{}).Where("status = ?", 1)
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if keyword != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	err := query.Preload("User").
		Order("is_top desc, created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&list).Error
	return list, total, err
}

func GetAllPolicyNotices(page, pageSize int, status string, keyword string) ([]model.PolicyNotice, int64, error) {
	var list []model.PolicyNotice
	var total int64

	query := config.DB.Model(&model.PolicyNotice{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	err := query.Preload("User").
		Order("is_top desc, created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&list).Error
	return list, total, err
}

func UpdatePolicyNotice(notice *model.PolicyNotice) error {
	return config.DB.Model(&model.PolicyNotice{}).Where("id = ?", notice.ID).Updates(map[string]interface{}{
		"title":        notice.Title,
		"content":      notice.Content,
		"category":     notice.Category,
		"address":      notice.Address,
		"publish_dept": notice.PublishDept,
		"publish_date": notice.PublishDate,
		"images":       notice.Images,
		"attachment":   notice.Attachment,
		"is_top":       notice.IsTop,
		"status":       notice.Status,
		"updated_at":   time.Now(),
	}).Error
}

func DeletePolicyNotice(id uint) error {
	return config.DB.Model(&model.PolicyNotice{}).Where("id = ?", id).Update("status", 2).Error
}

func HardDeletePolicyNotice(id uint) error {
	return config.DB.Delete(&model.PolicyNotice{}, id).Error
}

func AuditPolicyNotice(id uint, status int) error {
	if status != 1 && status != 0 {
		return errors.New("无效的审核状态")
	}
	return config.DB.Model(&model.PolicyNotice{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}).Error
}

// ==================== 地址关联服务 ====================

// AssociatedItemsResult 地址关联结果：同地址的公告和事务
type AssociatedItemsResult struct {
	Policies []model.PolicyNotice `json:"policies"` // 同地址的已发布公告
	Affairs  []model.RuralAffair  `json:"affairs"`  // 同地址的已发布事务
}

// GetAssociatedItemsByAddress 根据地址（忽略大小写）查询关联的公告和事务
func GetAssociatedItemsByAddress(address string) (*AssociatedItemsResult, error) {
	result := &AssociatedItemsResult{
		Policies: []model.PolicyNotice{},
		Affairs:  []model.RuralAffair{},
	}
	if address == "" {
		return result, nil
	}

	// 查询同地址的已发布公告（status=1，忽略大小写）
	if err := config.DB.
		Where("LOWER(address) = LOWER(?) AND status = ?", address, 1).
		Preload("User").
		Order("is_top desc, created_at desc").
		Find(&result.Policies).Error; err != nil {
		return nil, err
	}

	// 查询同地址的所有事务（忽略大小写，不限定状态）
	if err := config.DB.
		Where("LOWER(address) = LOWER(?)", address).
		Preload("User").
		Order("created_at desc").
		Find(&result.Affairs).Error; err != nil {
		return nil, err
	}

	return result, nil
}
