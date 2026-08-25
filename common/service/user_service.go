// Package service 提供用户账户管理服务（注册、登录、封禁、注销等）
package service

import (
	"rxtcloud/common/config"
	"rxtcloud/common/model"
	"time"
)

// 创建用户
func CreateUser(user model.User) error {
	return config.DB.Create(&user).Error
}

// 根据用户名查询
func GetUserByUsername(username string) (model.User, error) {
	var user model.User
	err := config.DB.Where("username = ?", username).First(&user).Error
	return user, err
}

// 根据ID查询用户
func GetUserByID(id uint) (model.User, error) {
	var user model.User
	err := config.DB.First(&user, id).Error
	return user, err
}

// 查询全部用户
func GetUserList() ([]model.User, error) {
	var users []model.User
	err := config.DB.Find(&users).Error
	return users, err
}

// 搜索用户（支持关键字和分页）
func SearchUsers(keyword string, page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	db := config.DB.Model(&model.User{})
	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("username LIKE ? OR nickname LIKE ?", like, like)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if pageSize <= 0 {
		pageSize = 10
	}
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * pageSize
	err := db.Select("id, username, nickname, avatar_url, role").Offset(offset).Limit(pageSize).Order("id ASC").Find(&users).Error
	return users, total, err
}

// 更新用户
func UpdateUser(user model.User) error {
	return config.DB.Save(&user).Error
}

// 更新用户资料（只更新部分字段，不更新username）
func UpdateUserProfile(id uint, nickname, phone, email, signature, avatarURL string) error {
	updates := map[string]interface{}{}
	if nickname != "" {
		updates["nickname"] = nickname
	}
	if phone != "" {
		updates["phone"] = phone
	}
	if email != "" {
		updates["email"] = email
	}
	if signature != "" {
		updates["signature"] = signature
	}
	if avatarURL != "" {
		updates["avatar_url"] = avatarURL
	}
	return config.DB.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error
}

// 删除用户
func DeleteUser(id uint) error {
	return config.DB.Delete(&model.User{}, id).Error
}

// 重置密码
func ResetPassword(username, newPassword string) error {
	var user model.User
	if err := config.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return err
	}
	return config.DB.Model(&user).Update("password", newPassword).Error
}

// BanUser 封禁用户：设置 banned=true，banned_until 到期时间（nil=永久封禁），status=0
func BanUser(userID uint, until *time.Time) error {
	updates := map[string]interface{}{
		"banned":       true,
		"banned_until": until,
		"status":       0,
	}
	return config.DB.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}

// 解封用户
func UnbanUser(userID uint) error {
	updates := map[string]interface{}{
		"banned":       false,
		"banned_until": nil,
		"status":       1,
	}
	return config.DB.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}

// 软删除用户（用户自助注销）
func SoftDeleteUser(userID uint) error {
	return config.DB.Delete(&model.User{}, userID).Error
}

// GetDeletedUsers 查询已注销用户列表（Unscoped 绕过 GORM 软删除过滤）
func GetDeletedUsers(page, pageSize int, keyword string) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	db := config.DB.Unscoped().Model(&model.User{}).Where("deleted_at IS NOT NULL")
	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("username LIKE ? OR nickname LIKE ?", like, like)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if pageSize <= 0 {
		pageSize = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	err := db.Order("deleted_at DESC").Offset(offset).Limit(pageSize).Find(&users).Error
	return users, total, err
}

// 更新已注销用户信息
func UpdateDeletedUser(userID uint, updates map[string]interface{}) error {
	return config.DB.Unscoped().Model(&model.User{}).Where("id = ? AND deleted_at IS NOT NULL", userID).Updates(updates).Error
}

// 获取全部用户列表（管理端用，分页+搜索，不含软删除用户）
func GetAllUsersForAdmin(page, pageSize int, keyword string, role string, banned *bool) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	db := config.DB.Model(&model.User{})

	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("username LIKE ? OR nickname LIKE ? OR phone LIKE ? OR email LIKE ?", like, like, like, like)
	}
	if role != "" {
		db = db.Where("role = ?", role)
	}
	if banned != nil {
		db = db.Where("banned = ?", *banned)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if pageSize <= 0 {
		pageSize = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	err := db.Order("id DESC").Offset(offset).Limit(pageSize).Find(&users).Error
	return users, total, err
}
