package service

import (
	"rxtcloud/common/config"
	"rxtcloud/common/model"
)

// 创建地址
func CreateAddress(address *model.Address) error {
	// 如果设置为默认，先取消其他默认地址
	if address.IsDefault == 1 {
		config.DB.Model(&model.Address{}).Where("user_id = ?", address.UserID).Update("is_default", 0)
	}
	return config.DB.Create(address).Error
}

// 获取用户所有地址
func GetAddressesByUserID(userID uint) ([]model.Address, error) {
	var addresses []model.Address
	err := config.DB.Where("user_id = ?", userID).
		Order("is_default desc, updated_at desc").
		Find(&addresses).Error
	return addresses, err
}

// 根据ID获取地址
func GetAddressByID(id uint) (*model.Address, error) {
	var address model.Address
	err := config.DB.First(&address, id).Error
	return &address, err
}

// 更新地址
func UpdateAddress(address *model.Address) error {
	// 先查询原记录，保留不可修改的字段
	var existing model.Address
	if err := config.DB.First(&existing, address.ID).Error; err != nil {
		return err
	}

	// 如果设置为默认，先取消其他默认地址
	if address.IsDefault == 1 {
		config.DB.Model(&model.Address{}).Where("user_id = ? AND id != ?", address.UserID, address.ID).Update("is_default", 0)
	}

	// 只更新允许修改的字段
	updates := map[string]interface{}{
		"receiver":   address.Receiver,
		"phone":      address.Phone,
		"province":   address.Province,
		"city":       address.City,
		"district":   address.District,
		"detail":     address.Detail,
		"is_default": address.IsDefault,
		"type":       address.Type,
		"label":      address.Label,
	}

	// 保留原始字段
	updates["user_id"] = existing.UserID
	updates["created_at"] = existing.CreatedAt

	return config.DB.Model(&model.Address{}).Where("id = ?", address.ID).Updates(updates).Error
}

// 删除地址
func DeleteAddress(id uint) error {
	return config.DB.Delete(&model.Address{}, id).Error
}

// 设置默认地址
func SetDefaultAddress(userID, addressID uint) error {
	// 取消该用户所有默认地址
	config.DB.Model(&model.Address{}).Where("user_id = ?", userID).Update("is_default", 0)
	// 设置新的默认地址
	return config.DB.Model(&model.Address{}).Where("id = ?", addressID).Update("is_default", 1).Error
}

// 获取默认地址
func GetDefaultAddress(userID uint) (*model.Address, error) {
	var address model.Address
	err := config.DB.Where("user_id = ? AND is_default = 1", userID).First(&address).Error
	if err != nil {
		// 没有默认地址则返回第一个
		err = config.DB.Where("user_id = ?", userID).First(&address).Error
	}
	return &address, err
}
