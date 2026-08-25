// Package service 提供商品核心业务逻辑
//
// 状态说明：1-已上架  -1-待审核  0-已下架/审核不通过  -2-已隐藏
//
// 商品发布流程：CreateProduct → 敏感词检测 → 有图且无敏感词自动通过(1) / 否则待审核(-1)
// 库存管理：结算/支付时扣减库存，异步通知卖家库存变化
package service

import (
	"rxtcloud/common/client"
	"rxtcloud/product-service/config"
	"rxtcloud/product-service/model"
)

// 创建商品
func CreateProduct(product *model.Product) error {
	// 默认待审核状态
	if product.Status == 0 {
		product.Status = -1
	}
	return config.DB.Create(product).Error
}

// 获取所有商品（分页，返回已上架的+当前用户自己的所有状态）
func GetProducts(page, pageSize int, keyword string, productType string, userID uint) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64

	// status=1（已上架）+ 作者本人的待审核(-1)/下架(0)
	query := config.DB.Model(&model.Product{})
	if userID > 0 {
		query = query.Where("status = 1 OR (status IN (-1, 0) AND publisher = ?)", userID)
	} else {
		query = query.Where("status = 1")
	}
	if productType != "" {
		query = query.Where("type = ?", productType)
	}
	if keyword != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("PublisherUser").
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&products).Error

	return products, total, err
}

// 根据ID获取商品
func GetProductByID(id uint) (*model.Product, error) {
	var product model.Product
	err := config.DB.Preload("PublisherUser").First(&product, id).Error
	return &product, err
}

// 更新商品（只更新非零值字段）
func UpdateProduct(product *model.Product) error {
	// 先查询原记录
	var existing model.Product
	if err := config.DB.First(&existing, product.ID).Error; err != nil {
		return err
	}

	// 构建更新映射，只更新非零值的字段
	updates := map[string]interface{}{}

	// 检查每个字段，只有非零值才更新
	if product.Title != "" {
		updates["title"] = product.Title
	}
	if product.Type != "" {
		updates["type"] = product.Type
	}
	if product.Price > 0 {
		updates["price"] = product.Price
	}
	if product.Content != "" {
		updates["content"] = product.Content
	}
	if product.ImageURL != "" {
		updates["image_url"] = product.ImageURL
	}
	if product.Stock >= 0 {
		updates["stock"] = product.Stock
	}
	if product.Address != "" {
		updates["address"] = product.Address
	}
	// status 为 0 也是有效值（下架）
	if product.Status != 0 || (product.ID > 0) {
		// 只有显式传入 status 才更新
	}

	// 保留不可修改的字段
	updates["publisher"] = existing.Publisher
	updates["created_at"] = existing.CreatedAt

	if len(updates) == 2 { // 只有保留字段，没有要更新的
		return nil
	}

	err := config.DB.Model(&model.Product{}).Where("id = ?", product.ID).Updates(updates).Error
	if err == nil {
		// 重新查询更新后的商品用于检查库存，异步通知 message-service
		var updated model.Product
		config.DB.First(&updated, product.ID)
		notifClient := client.NewServiceClient("product-service")
		notifClient.CallAsync("message-service", "POST", "/api/internal/notify/stock", map[string]interface{}{
			"product_id":    updated.ID,
			"product_title": updated.Title,
			"stock":         updated.Stock,
			"publisher_id":  updated.Publisher,
		})
	}
	return err
}

// 更新商品状态（仅更新status字段）
func UpdateProductStatus(id uint, status int) error {
	return config.DB.Model(&model.Product{}).Where("id = ?", id).Update("status", status).Error
}

// ReduceProductStock 扣减商品库存（支付时调用），库存<0 时取 0，异步通知卖家
func ReduceProductStock(productID uint, quantity int) error {
	var product model.Product
	if err := config.DB.First(&product, productID).Error; err != nil {
		return err
	}

	newStock := product.Stock - quantity
	if newStock < 0 {
		newStock = 0
	}

	if err := config.DB.Model(&product).Update("stock", newStock).Error; err != nil {
		return err
	}

	// 检查库存并发送通知
	product.Stock = newStock
	notifClient := client.NewServiceClient("product-service")
	notifClient.CallAsync("message-service", "POST", "/api/internal/notify/stock", map[string]interface{}{
		"product_id":    product.ID,
		"product_title": product.Title,
		"stock":         product.Stock,
		"publisher_id":  product.Publisher,
	})

	return nil
}

// 删除商品
func DeleteProduct(id uint) error {
	return config.DB.Delete(&model.Product{}, id).Error
}

// 根据发布者获取商品
func GetProductsByPublisher(publisher uint) ([]model.Product, error) {
	var products []model.Product
	err := config.DB.Where("publisher = ?", publisher).
		Order("created_at desc").
		Find(&products).Error
	return products, err
}

// 根据状态获取商品列表
func GetProductsByStatus(status int) ([]model.Product, error) {
	var products []model.Product
	err := config.DB.Where("status = ?", status).
		Order("created_at desc").
		Find(&products).Error
	return products, err
}

// 获取待审核商品列表（管理员）
func GetPendingProductsForAdmin() ([]model.Product, error) {
	var products []model.Product
	err := config.DB.Preload("PublisherUser").
		Where("status = -1").
		Order("created_at desc").
		Find(&products).Error
	return products, err
}

// 审核商品
func AuditProduct(id uint, approved bool) error {
	status := 1 // 通过，上架
	if !approved {
		status = 0 // 不通过，下架
	}
	return config.DB.Model(&model.Product{}).Where("id = ?", id).Update("status", status).Error
}

// 恢复商品（重新提交审核）
func RestoreProduct(id uint) error {
	return config.DB.Model(&model.Product{}).Where("id = ?", id).Update("status", -1).Error
}

// 根据类型获取商品
func GetProductsByType(productType string) ([]model.Product, error) {
	var products []model.Product
	err := config.DB.Where("type = ?", productType).
		Order("created_at desc").
		Find(&products).Error
	return products, err
}

// 获取所有商品（管理员用，包含下架商品）
func GetAllProductsForAdmin(page, pageSize int, status int) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64

	query := config.DB.Model(&model.Product{})
	if status != -999 { // -999 表示获取所有状态
		query = query.Where("status = ?", status)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := config.DB.Preload("PublisherUser").
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&products).Error

	return products, total, err
}

// 获取已退回商品（审核不通过）
func GetRejectedProductsForAdmin() ([]model.Product, error) {
	var products []model.Product
	err := config.DB.Preload("PublisherUser").
		Where("status = 0"). // 0 = 下架/退回
		Order("created_at desc").
		Find(&products).Error
	return products, err
}

// 获取已隐藏商品
func GetHiddenProductsForAdmin() ([]model.Product, error) {
	var products []model.Product
	err := config.DB.Preload("PublisherUser").
		Where("status = -2"). // -2 = 隐藏
		Order("created_at desc").
		Find(&products).Error
	return products, err
}
