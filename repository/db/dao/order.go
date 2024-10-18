package dao

import (
	"context"

	"gorm.io/gorm"

	model2 "mall/repository/db/model"
)

type OrderDao struct {
	*gorm.DB
}

func NewOrderDao(ctx context.Context) *OrderDao {
	return &OrderDao{NewDBClient(ctx)}
}

func NewOrderDaoByDB(db *gorm.DB) *OrderDao {
	return &OrderDao{db}
}

// CreateOrder 创建订单
// 虑将订单创建操作放在一个事务中，尤其是当需要在创建订单时执行其他操作（例如扣减库存）时
func (dao *OrderDao) CreateOrder(order *model2.Order) error {
	tx := dao.DB.Begin()
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback() // 回滚事务
		return err
	}
	return tx.Commit().Error // 提交事务
}

// ListOrderByCondition 获取订单List
func (dao *OrderDao) ListOrderByCondition(condition map[string]interface{}, page model2.BasePage) (orders []*model2.Order, total int64, err error) {
	err = dao.DB.Model(&model2.Order{}).Where(condition).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = dao.DB.Model(&model2.Order{}).Where(condition).
		Offset((page.PageNum - 1) * page.PageSize).
		Limit(page.PageSize).Order("created_at desc").Find(&orders).Error
	return
}

// GetOrderById 获取订单详情
func (dao *OrderDao) GetOrderById(id uint) (order *model2.Order, err error) {
	// err = dao.DB.Model(&model2.Order{}).Where("id=?", id).
	// 	First(&order).Error
	// return

	// 使用 Unscoped() 查询所有记录，包括被软删除的
	if err := dao.DB.Unscoped().Where("id = ?", id).First(&order).Error; err != nil {
		return nil, err
	}
	return order, nil
}

// GetOrderByOrderNum 获取订单详情
func (dao *OrderDao) GetOrderByOrderNum(orderNum uint64) (order *model2.Order, err error) {
	err = dao.DB.Model(&model2.Order{}).Where("order_num=?", orderNum).
		First(&order).Error
	return
}

// DeleteOrderById 删除订单详情
func (dao *OrderDao) DeleteOrderById(id uint) error {
	return dao.DB.Where("id=?", id).Delete(&model2.Order{}).Error
}

// UpdateOrderById 更新订单详情
func (dao *OrderDao) UpdateOrderById(id uint, order *model2.Order) error {
	return dao.DB.Where("id=?", id).Updates(order).Error
}
