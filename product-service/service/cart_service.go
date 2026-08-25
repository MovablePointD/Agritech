// Package service 提供购物车与订单的完整业务流程
//
// 购物车 → 结算(CheckoutCart) → 订单 → 支付 → 发货 → 确认收货
// 订单状态：0-待付款 1-已支付 2-已发货 3-已完成 4-已取消
package service

import (
	"errors"
	"fmt"
	"rxtcloud/common/client"
	"rxtcloud/product-service/config"
	"rxtcloud/product-service/model"
)

// CartService 购物车服务，单例模式
type CartService struct{}

// CartServiceInstance 全局购物车服务实例
var CartServiceInstance = &CartService{}

// 获取用户购物车列表
func (s *CartService) GetMyCartItems(userID uint) ([]model.CartItem, error) {
	var items []model.CartItem
	err := config.DB.Preload("Product").Where("user_id = ?", userID).Order("created_at desc").Find(&items).Error
	return items, err
}

// 获取单个购物车项
func (s *CartService) GetCartItem(id uint, userID uint) (*model.CartItem, error) {
	var item model.CartItem
	err := config.DB.Preload("Product").Where("id = ? AND user_id = ?", id, userID).First(&item).Error
	if err != nil {
		return nil, errors.New("购物车项不存在")
	}
	return &item, nil
}

// AddToCart 添加商品到购物车，已存在则累加数量，不存在则新增
func (s *CartService) AddToCart(userID, productID uint, quantity int) (*model.CartItem, error) {
	// 检查商品是否存在
	var product model.Product
	if err := config.DB.First(&product, productID).Error; err != nil {
		return nil, errors.New("商品不存在")
	}

	// 检查是否已在购物车中
	var existing model.CartItem
	err := config.DB.Where("user_id = ? AND product_id = ?", userID, productID).First(&existing).Error
	if err == nil {
		// 已存在，增加数量
		existing.Quantity += quantity
		existing.TotalPrice = float64(existing.Quantity) * existing.Price
		if err := config.DB.Save(&existing).Error; err != nil {
			return nil, errors.New("更新购物车失败")
		}
		return &existing, nil
	}

	// 新增购物车项
	item := model.CartItem{
		UserID:     userID,
		ProductID:  productID,
		Quantity:   quantity,
		Price:      product.Price,
		TotalPrice: float64(quantity) * product.Price,
	}

	if err := config.DB.Create(&item).Error; err != nil {
		return nil, errors.New("添加到购物车失败")
	}

	// 重新加载关联的商品信息
	config.DB.Preload("Product").First(&item, item.ID)
	return &item, nil
}

// 更新购物车项数量
func (s *CartService) UpdateCartItem(id uint, userID uint, quantity int) (*model.CartItem, error) {
	var item model.CartItem
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).First(&item).Error; err != nil {
		return nil, errors.New("购物车项不存在")
	}

	if quantity <= 0 {
		return nil, errors.New("数量必须大于0")
	}

	item.Quantity = quantity
	item.TotalPrice = float64(quantity) * item.Price

	if err := config.DB.Save(&item).Error; err != nil {
		return nil, errors.New("更新失败")
	}

	config.DB.Preload("Product").First(&item, item.ID)
	return &item, nil
}

// 从购物车移除
func (s *CartService) RemoveFromCart(id uint, userID uint) error {
	result := config.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&model.CartItem{})
	if result.RowsAffected == 0 {
		return errors.New("购物车项不存在")
	}
	return nil
}

// 清空用户购物车
func (s *CartService) ClearCart(userID uint) error {
	return config.DB.Where("user_id = ?", userID).Delete(&model.CartItem{}).Error
}

// 获取购物车商品总数
func (s *CartService) GetCartCount(userID uint) (int64, error) {
	var count int64
	err := config.DB.Model(&model.CartItem{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// 批量添加商品到购物车
func (s *CartService) BatchAddToCart(userID uint, items []struct {
	ProductID uint `json:"product_id"`
	Quantity  int  `json:"quantity"`
}) ([]model.CartItem, error) {
	var results []model.CartItem
	for _, item := range items {
		result, err := s.AddToCart(userID, item.ProductID, item.Quantity)
		if err != nil {
			return nil, err
		}
		results = append(results, *result)
	}
	return results, nil
}

// CheckoutCart 购物车结算：每个选中商品独立创建订单，使用事务保证原子性
// 流程：验证地址 → 验证商品在售&库存 → 事务内创建订单+订单项+扣减库存+删除购物车项
// 异步通知卖家库存变化
func (s *CartService) CheckoutCart(userID uint, itemIDs []uint, addressID uint, remark string) ([]model.Order, error) {
	// 验证地址
	var address model.Address
	if err := config.DB.Where("id = ? AND user_id = ?", addressID, userID).First(&address).Error; err != nil {
		return nil, errors.New("收货地址不存在")
	}

	// 获取选中的购物车项
	var cartItems []model.CartItem
	if len(itemIDs) == 0 {
		return nil, errors.New("请选择要结算的商品")
	}

	err := config.DB.Preload("Product").Where("id IN ? AND user_id = ?", itemIDs, userID).Find(&cartItems).Error
	if err != nil || len(cartItems) == 0 {
		return nil, errors.New("购物车项不存在")
	}

	// 检查商品是否上架及库存
	for _, item := range cartItems {
		if item.Product == nil {
			return nil, errors.New("商品信息不存在")
		}
		if item.Product.Status != 1 {
			return nil, errors.New("商品 [" + item.Product.Title + "] 未上架，无法购买")
		}
		// 检查库存
		if item.Quantity > item.Product.Stock {
			if item.Product.Stock == 0 {
				return nil, errors.New("商品 [" + item.Product.Title + "] 库存不足，暂时无法购买")
			}
			return nil, errors.New("商品 [" + item.Product.Title + "] 库存不足（当前库存：" + fmt.Sprintf("%d", item.Product.Stock) + "）")
		}
	}

	// 收货地址信息
	fullAddress := addressProvince(&address) + address.City + address.District + address.Address

	// 使用事务创建订单
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var orders []model.Order

	// 为每个购物车项创建单独的订单
	for _, item := range cartItems {
		orderNo := GenerateOrderNo()

		// 构建订单
		var order = model.Order{
			OrderNo:    orderNo,
			UserID:     userID,
			OrderType:  "buy",
			TotalPrice: item.TotalPrice,
			Receiver:   address.Receiver,
			Phone:      address.Phone,
			Address:    fullAddress,
			Status:     0, // 待付款
			PayStatus:  0,
			Remark:     remark,
		}

		// 创建订单
		if err := tx.Create(&order).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("创建订单失败")
		}

		// 创建订单项
		orderItem := model.OrderItem{
			OrderID:     order.ID,
			ProductID:   item.ProductID,
			ProductName: item.Product.Title,
			Price:       item.Price,
			Quantity:    item.Quantity,
			Subtotal:    item.TotalPrice,
		}
		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("创建订单项失败")
		}

		// 扣减库存（购物车结算时预扣库存）
		newStock := item.Product.Stock - item.Quantity
		if newStock < 0 {
			newStock = 0
		}
		if err := tx.Model(&model.Product{}).Where("id = ?", item.ProductID).Update("stock", newStock).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("库存更新失败")
		}

		// 记录需要发送通知的商品（库存变化后的值）
		productForNotification := &model.Product{
			ID:        item.ProductID,
			Title:     item.Product.Title,
			Publisher: item.Product.Publisher,
			Stock:     newStock,
		}

		// 重新加载订单项
		tx.Preload("Items").First(&order, order.ID)
		orders = append(orders, order)

		// 在事务外处理通知
		defer func() {
			notifClient := client.NewServiceClient("product-service")
			notifClient.CallAsync("message-service", "POST", "/api/internal/notify/stock", map[string]interface{}{
				"product_id":    productForNotification.ID,
				"product_title": productForNotification.Title,
				"stock":         productForNotification.Stock,
				"publisher_id":  productForNotification.Publisher,
			})
		}()
	}

	// 删除已结算的购物车项
	if err := tx.Where("id IN ?", itemIDs).Delete(&model.CartItem{}).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("清除购物车失败")
	}

	tx.Commit()

	return orders, nil
}

// addressProvince 获取地址的省份信息
func addressProvince(addr *model.Address) string {
	return addr.Province
}

// DirectBuy 直接购买（不经过购物车，直接从商品详情页下单）
// 注意：此方法不扣减库存，库存扣减在 PayOrder 时执行
func (s *CartService) DirectBuy(userID, productID uint, quantity int, addressID uint, remark string) (*model.Order, error) {
	// 验证地址
	var address model.Address
	if err := config.DB.Where("id = ? AND user_id = ?", addressID, userID).First(&address).Error; err != nil {
		return nil, errors.New("收货地址不存在")
	}

	// 获取商品信息
	var product model.Product
	if err := config.DB.First(&product, productID).Error; err != nil {
		return nil, errors.New("商品不存在")
	}

	// 检查商品是否上架
	if product.Status != 1 {
		return nil, errors.New("商品未上架，无法购买")
	}

	// 检查库存
	if quantity > product.Stock {
		if product.Stock == 0 {
			return nil, errors.New("商品库存不足，暂时无法购买")
		}
		return nil, errors.New("商品库存不足（当前库存：" + fmt.Sprintf("%d", product.Stock) + "）")
	}

	// 计算总价
	subtotal := float64(quantity) * product.Price

	// 生成订单编号
	orderNo := GenerateOrderNo()

	// 构建订单项
	orderItems := []model.OrderItem{{
		ProductID:   productID,
		ProductName: product.Title,
		Price:       product.Price,
		Quantity:    quantity,
		Subtotal:    subtotal,
	}}

	// 构建订单
	var order = model.Order{
		OrderNo:    orderNo,
		UserID:     userID,
		OrderType:  "buy",
		TotalPrice: subtotal,
		Receiver:   address.Receiver,
		Phone:      address.Phone,
		Address:    addressProvince(&address) + address.City + address.District + address.Address,
		Status:     0, // 待付款
		PayStatus:  0,
		Remark:     remark,
		Items:      orderItems,
	}

	// 创建订单（使用 Omit 排除 Items，避免 GORM 嵌套创建导致重复）
	// 先清空 order 的 Items 关联，防止 GORM 意外创建
	order.Items = nil
	if err := config.DB.Omit("Items").Create(&order).Error; err != nil {
		return nil, errors.New("创建订单失败: " + err.Error())
	}

	// 单独创建订单项
	for _, item := range orderItems {
		newItem := model.OrderItem{
			OrderID:     order.ID,
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
			Subtotal:    item.Subtotal,
		}
		if err := config.DB.Create(&newItem).Error; err != nil {
			return nil, errors.New("创建订单项失败")
		}
	}

	// 重新加载订单及其关联
	config.DB.Preload("Items").First(&order, order.ID)
	return &order, nil
}
