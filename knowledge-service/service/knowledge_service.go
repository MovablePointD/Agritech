// Package service 提供农业知识核心业务逻辑
//
// 状态说明：1-已发布  -1-待审核  0-已退回  2-已删除  3-草稿
//
// 热门推荐算法：
//   HotScore = likes / (TIMESTAMPDIFF(DAY, created_at, NOW()) + 2)
//   综合点赞数和发布时间衰减
package service

import (
	"rxtcloud/knowledge-service/config"
	"rxtcloud/knowledge-service/model"
)

// 新增
func CreateKnowledge(k model.Knowledge) error {
	// 默认待审核状态
	if k.Status == 0 {
		k.Status = -1
	}
	return config.DB.Create(&k).Error
}

// 分页查询全部（返回已发布的+当前用户自己的所有状态）
func GetKnowledgeList(page int, pageSize int, keyword string, userID uint) ([]model.Knowledge, int64, error) {
	var list []model.Knowledge
	var total int64

	// status=1（已发布）+ 作者本人的待审核(-1)/退回(0)
	db := config.DB.Model(&model.Knowledge{})
	if userID > 0 {
		db = db.Where("status = 1 OR (status IN (-1, 0) AND user_id = ?)", userID)
	} else {
		db = db.Where("status = 1")
	}

	// 模糊搜索
	if keyword != "" {
		db = db.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	db.Count(&total)

	offset := (page - 1) * pageSize

	err := db.
		Preload("User").
		Limit(pageSize).
		Offset(offset).
		Order("created_at desc").
		Find(&list).Error

	return list, total, err
}

// 根据ID查询
func GetKnowledgeByID(id uint) (model.Knowledge, error) {
	var k model.Knowledge
	err := config.DB.Preload("User").First(&k, id).Error
	return k, err
}

// 更新
func UpdateKnowledge(k model.Knowledge) error {
	// 先查询原记录，保留不可修改的字段
	var existing model.Knowledge
	if err := config.DB.First(&existing, k.ID).Error; err != nil {
		return err
	}

	// 只更新允许修改的字段
	updates := map[string]interface{}{
		"title":      k.Title,
		"content":    k.Content,
		"image_url":  k.ImageURL,
	}

	// 保留原始字段
	updates["user_id"] = existing.UserID
	updates["created_at"] = existing.CreatedAt

	return config.DB.Model(&model.Knowledge{}).Where("id = ?", k.ID).Updates(updates).Error
}

// 删除（软删除）
func DeleteKnowledge(id uint) error {
	return config.DB.Model(&model.Knowledge{}).Where("id = ?", id).Update("status", 2).Error
}

// 获取当前用户发布的知识列表
func GetMyKnowledge(userID uint) ([]model.Knowledge, error) {
	var list []model.Knowledge
	err := config.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&list).Error
	return list, err
}

// 获取所有知识（管理员，包含已删除）
func GetAllKnowledgeForAdmin(page, pageSize int, status int) ([]model.Knowledge, int64, error) {
	var list []model.Knowledge
	var total int64

	query := config.DB.Model(&model.Knowledge{})
	if status != -999 { // -999 表示获取所有状态
		query = query.Where("status = ?", status)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := config.DB.Preload("User").
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&list).Error

	return list, total, err
}

// 获取已退回知识（审核不通过）
func GetRejectedKnowledgeForAdmin() ([]model.Knowledge, error) {
	var list []model.Knowledge
	err := config.DB.Preload("User").
		Where("status = 0"). // 0 = 隐藏/退回
		Order("created_at desc").
		Find(&list).Error
	return list, err
}

// 获取已删除知识
func GetDeletedKnowledgeForAdmin() ([]model.Knowledge, error) {
	var list []model.Knowledge
	err := config.DB.Preload("User").
		Where("status = 2"). // 2 = 删除
		Order("created_at desc").
		Find(&list).Error
	return list, err
}

// 获取待审核知识列表
func GetPendingKnowledgeForAdmin() ([]model.Knowledge, error) {
	var list []model.Knowledge
	err := config.DB.Preload("User").
		Where("status = -1").
		Order("created_at desc").
		Find(&list).Error
	return list, err
}

// 审核知识
func AuditKnowledge(id uint, approved bool) error {
	status := 1 // 通过
	if !approved {
		status = 0 // 隐藏
	}
	return config.DB.Model(&model.Knowledge{}).Where("id = ?", id).Update("status", status).Error
}

// 恢复知识（重新提交审核）
func RestoreKnowledge(id uint) error {
	return config.DB.Model(&model.Knowledge{}).Where("id = ?", id).Update("status", -1).Error
}

// ========== 草稿相关 ==========

// 创建知识草稿
func CreateDraftKnowledge(k *model.Knowledge) error {
	k.Status = 3
	return config.DB.Create(k).Error
}

// 更新知识草稿（仅允许本人更新 status=3 的记录）
func UpdateDraftKnowledge(k *model.Knowledge) error {
	updates := map[string]interface{}{
		"title":     k.Title,
		"content":   k.Content,
		"image_url": k.ImageURL,
	}
	return config.DB.Model(&model.Knowledge{}).
		Where("id = ? AND status = 3 AND user_id = ?", k.ID, k.UserID).
		Updates(updates).Error
}

// 获取用户的知识草稿列表
func GetDraftKnowledge(userID uint) ([]model.Knowledge, error) {
	var list []model.Knowledge
	err := config.DB.Preload("User").
		Where("user_id = ? AND status = 3", userID).
		Order("updated_at desc").
		Find(&list).Error
	return list, err
}

// 删除知识草稿（物理删除）
func DeleteDraftKnowledge(id uint, userID uint) error {
	return config.DB.Where("id = ? AND status = 3 AND user_id = ?", id, userID).
		Delete(&model.Knowledge{}).Error
}

// 获取知识草稿详情（仅本人可查看）
func GetDraftKnowledgeByID(id uint, userID uint) (*model.Knowledge, error) {
	var k model.Knowledge
	err := config.DB.Preload("User").
		Where("id = ? AND status = 3 AND user_id = ?", id, userID).
		First(&k).Error
	return &k, err
}

// 发布知识草稿（status: 3 → 1 或 -1）
func PublishKnowledgeFromDraft(id uint, userID uint, newStatus int) error {
	return config.DB.Model(&model.Knowledge{}).
		Where("id = ? AND status = 3 AND user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"status": newStatus,
		}).Error
}

// 获取热门知识（综合点赞数与发布时间的推荐算法）
// 热度 = likes / (TIMESTAMPDIFF(DAY, created_at, NOW()) + 2)
func GetHotKnowledgeList(page, pageSize int) ([]model.Knowledge, int64, error) {
	var list []model.Knowledge
	var total int64

	hotScore := "likes / (TIMESTAMPDIFF(DAY, created_at, NOW()) + 2)"

	query := config.DB.Model(&model.Knowledge{}).Where("status = 1")
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.
		Preload("User").
		Limit(pageSize).
		Offset(offset).
		Order(hotScore + " desc, created_at desc").
		Find(&list).Error

	return list, total, err
}
