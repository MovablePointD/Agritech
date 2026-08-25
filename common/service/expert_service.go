// Package service 提供农业专家申请、审核与查询服务
//
// 专家状态：0-待审核  1-审核通过  2-审核不通过
// 审核通过时自动更新用户角色为 expert（除非已是 processor/sysadmin/admin）
// SearchExperts 按姓名/专业/单位模糊搜索
package service

import (
	"rxtcloud/common/config"
	"rxtcloud/common/model"
)

// 创建专家申请
func CreateExpert(expert *model.Expert) error {
	expert.Status = 0 // 待审核
	return config.DB.Create(expert).Error
}

// 获取所有专家（已审核通过的）
func GetExperts(page, pageSize int) ([]model.Expert, int64, error) {
	var experts []model.Expert
	var total int64

	config.DB.Model(&model.Expert{}).Where("status = 1").Count(&total)

	offset := (page - 1) * pageSize
	err := config.DB.Where("status = 1").
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&experts).Error

	return experts, total, err
}

// 根据ID获取专家
func GetExpertByID(id uint) (*model.Expert, error) {
	var expert model.Expert
	err := config.DB.First(&expert, id).Error
	return &expert, err
}

// 根据用户ID获取专家信息
func GetExpertByUserID(userID uint) (*model.Expert, error) {
	var expert model.Expert
	err := config.DB.Where("user_id = ?", userID).First(&expert).Error
	if err != nil {
		return nil, err
	}
	return &expert, nil
}

// 获取所有待审核专家
func GetPendingExperts() ([]model.Expert, error) {
	var experts []model.Expert
	err := config.DB.Where("status = 0").Order("created_at asc").Find(&experts).Error
	return experts, err
}

// 更新专家信息
func UpdateExpert(expert *model.Expert) error {
	// 先查询原记录，保留不可修改的字段
	var existing model.Expert
	if err := config.DB.First(&existing, expert.ID).Error; err != nil {
		return err
	}

	// 只更新允许修改的字段
	updates := map[string]interface{}{
		"real_name":  expert.RealName,
		"phone":      expert.Phone,
		"profession": expert.Profession,
		"title":      expert.Title,
		"company":    expert.Company,
		"intro":      expert.Intro,
		"cert_no":    expert.CertNo,
	}

	// 保留原始字段
	updates["user_id"] = existing.UserID
	updates["status"] = existing.Status
	updates["created_at"] = existing.CreatedAt

	return config.DB.Model(&model.Expert{}).Where("id = ?", expert.ID).Updates(updates).Error
}

// 删除专家
func DeleteExpert(id uint) error {
	return config.DB.Delete(&model.Expert{}, id).Error
}

// 审核通过专家
func ApproveExpert(id uint) error {
	// 先获取专家信息
	var expert model.Expert
	if err := config.DB.First(&expert, id).Error; err != nil {
		return err
	}
	// 获取用户当前角色
	var user model.User
	if err := config.DB.First(&user, expert.UserID).Error; err != nil {
		return err
	}
	// 如果用户不是processor或sysadmin，则更新为expert；否则保留processor/sysadmin角色（专家权限通过expert表status判断）
	currentRole := user.Role
	if currentRole != "processor" && currentRole != "sysadmin" && currentRole != "admin" {
		if err := config.DB.Model(&model.User{}).Where("id = ?", expert.UserID).Update("role", "expert").Error; err != nil {
			return err
		}
	}
	return config.DB.Model(&model.Expert{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":    1,
			"verify_at": config.DB.NowFunc(),
		}).Error
}

// 审核拒绝专家
func RejectExpert(id uint) error {
	return config.DB.Model(&model.Expert{}).Where("id = ?", id).Update("status", 2).Error
}

// 根据专业筛选专家
func GetExpertsByProfession(profession string) ([]model.Expert, error) {
	var experts []model.Expert
	err := config.DB.Where("profession = ? AND status = 1", profession).
		Order("created_at desc").
		Find(&experts).Error
	return experts, err
}

// 搜索专家（姓名/专业/单位）
func SearchExperts(keyword string) ([]model.Expert, error) {
	var experts []model.Expert
	err := config.DB.Where("status = 1 AND (real_name LIKE ? OR profession LIKE ? OR company LIKE ?)",
		"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%").
		Order("created_at desc").
		Find(&experts).Error
	return experts, err
}

// 获取所有专家申请（包含所有状态）
func GetAllExperts() ([]model.Expert, error) {
	var experts []model.Expert
	err := config.DB.Preload("User").Order("created_at desc").Find(&experts).Error
	return experts, err
}
