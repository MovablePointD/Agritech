// Package service 提供订单全生命周期管理
//
// 订单状态机：0-待付款 → 1-已支付(扣减库存) → 2-已发货 → 3-已完成(确认收货) → 4-已取消
// 支付状态独立管理：pay_status: 0-未支付 1-已支付
package service

import (
	"errors"
	"fmt"
	"rxtcloud/product-service/config"
	"rxtcloud/product-service/model"
	"time"

	"gorm.io/gorm"
)

// GenerateOrderNo 生成唯一订单编号：时间戳(秒)+4位纳秒随机数
func GenerateOrderNo() string {
	return fmt.Sprintf("%d%04d", time.Now().Unix(), time.Now().Nanosecond()%10000)
}

// CreateOrder 创建购买订单，事务内验证商品在售状态和库存
func CreateOrder(order *model.Order) error {
	// 仅允许创建购买类型订单
	if order.OrderType != "buy" {
		return errors.New("非法订单类型")
	}

	// 使用事务确保原子性
	return config.DB.Transaction(func(tx *gorm.DB) error {
		// 验证商品是否上架并检查库存
		if order.Items != nil && len(order.Items) > 0 {
			for _, item := range order.Items {
				var product model.Product
				if err := tx.First(&product, item.ProductID).Error; err != nil {
					return errors.New("商品不存在")
				}
				if product.Status != 1 {
					return errors.New("商品 [" + product.Title + "] 未上架，无法购买")
				}
				// 检查库存是否充足（购买数量 <= 库存）
				if item.Quantity > product.Stock {
					if product.Stock == 0 {
						return errors.New("商品 [" + product.Title + "] 库存不足，暂时无法购买")
					}
					return errors.New("商品 [" + product.Title + "] 库存不足（当前库存：" + fmt.Sprintf("%d", product.Stock) + "）")
				}
			}
		}

		order.OrderNo = GenerateOrderNo()
		order.Status = 0    // 待付款
		order.PayStatus = 0 // 未支付
		return tx.Create(order).Error
	})
}

// 获取订单列表（分页）
func GetOrders(page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	config.DB.Model(&model.Order{}).Count(&total)

	offset := (page - 1) * pageSize
	err := config.DB.Preload("Items").
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&orders).Error

	return orders, total, err
}

// 根据ID获取订单
func GetOrderByID(id uint) (*model.Order, error) {
	var order model.Order
	err := config.DB.Preload("Items").First(&order, id).Error
	return &order, err
}

// 根据订单号获取订单
func GetOrderByNo(orderNo string) (*model.Order, error) {
	var order model.Order
	err := config.DB.Where("order_no = ?", orderNo).First(&order).Error
	return &order, err
}

// 根据用户ID获取订单
func GetOrdersByUserID(userID uint) ([]model.Order, error) {
	var orders []model.Order
	err := config.DB.Preload("Items").
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&orders).Error
	return orders, err
}

// 更新订单
func UpdateOrder(order *model.Order) error {
	return config.DB.Save(order).Error
}

// 删除订单
func DeleteOrder(id uint) error {
	return config.DB.Delete(&model.Order{}, id).Error
}

// 更新订单状态
func UpdateOrderStatus(id uint, status int) error {
	return config.DB.Model(&model.Order{}).Where("id = ?", id).Update("status", status).Error
}

// PayOrder 支付订单：事务内扣减库存 + 更新支付状态（原子操作）
// 库存扣减调用 ReduceProductStock，会异步通知卖家
func PayOrder(id uint) error {
	return config.DB.Transaction(func(tx *gorm.DB) error {
		// 获取订单信息
		var order model.Order
		if err := tx.Preload("Items").First(&order, id).Error; err != nil {
			return errors.New("订单不存在")
		}

		// 扣减库存
		if order.Items != nil && len(order.Items) > 0 {
			for _, item := range order.Items {
				if err := ReduceProductStock(item.ProductID, item.Quantity); err != nil {
					return errors.New("库存更新失败: " + err.Error())
				}
			}
		}

		// 更新订单状态
		return tx.Model(&model.Order{}).Where("id = ?", id).
			Updates(map[string]interface{}{
				"pay_status": 1,
				"status":     1,
			}).Error
	})
}

// 发货（更新物流）
func ShipOrder(id uint, expressNo string) error {
	return config.DB.Model(&model.Order{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"express_no": expressNo,
			"status":     2,
		}).Error
}

// 确认收货
func ConfirmOrder(id uint) error {
	return config.DB.Model(&model.Order{}).Where("id = ?", id).Update("status", 3).Error
}

// 取消订单
func CancelOrder(id uint) error {
	return config.DB.Model(&model.Order{}).Where("id = ? AND status = 0", id).Update("status", 4).Error
}

// 根据状态获取订单
func GetOrdersByStatus(status int) ([]model.Order, error) {
	var orders []model.Order
	err := config.DB.Where("status = ?", status).
		Order("created_at desc").
		Find(&orders).Error
	return orders, err
}

// GetSellerOrders 获取卖家订单：通过 OrderItem → Product → Publisher 三表 JOIN 查询
// 排除已取消(status=4)的订单
func GetSellerOrders(userID uint) ([]model.Order, error) {
	var orders []model.Order

	// 查询所有包含该用户发布的商品的订单
	// 通过 OrderItem -> Product -> Publisher 关联
	err := config.DB.Preload("Items").
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Joins("JOIN products ON products.id = order_items.product_id").
		Where("products.publisher = ?", userID).
		Where("orders.status != ?", 4). // 排除已取消的订单
		Group("orders.id").
		Order("orders.created_at desc").
		Find(&orders).Error

	return orders, err
}
