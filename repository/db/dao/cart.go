package dao

import (
	"context"

	"gorm.io/gorm"

	"mall/pkg/e"
	"mall/repository/db/model"
)

type CartDao struct {
	*gorm.DB
}

func NewCartDao(ctx context.Context) *CartDao {
	return &CartDao{NewDBClient(ctx)}
}

func NewCartDaoByDB(db *gorm.DB) *CartDao {
	return &CartDao{db}
}

// CreateCart 创建 cart pId(商品 id)、uId(用户id)、bId(店家id)
func (dao *CartDao) CreateCart(pId, uId, bId uint, num int) (cart *model.Cart, status int, err error) {
	// 查询有无此条商品
	cart, err = dao.GetCartById(pId, uId, bId)
	// 空的，第一次加入
	if err == gorm.ErrRecordNotFound {
		cart = &model.Cart{
			UserID:    uId,
			ProductID: pId,
			BossID:    bId,
			Num:       num,
			MaxNum:    10,
			Check:     false,
		}
		err = dao.DB.Create(&cart).Error
		if err != nil {
			return
		}
		return cart, e.SUCCESS, err
	} else if (cart.Num + num) <= cart.MaxNum { // 如果购物车中已有此商品，且数量未达到最大值，则将数量增加 num
		// 小于最大 num
		cart.Num += num
		err = dao.DB.Save(&cart).Error
		if err != nil {
			return
		}
		return cart, e.ErrorProductExistCart, err
	} else {
		// 大于最大num
		return cart, e.ErrorProductMoreCart, err
	}
}

// GetCartById 获取 Cart 通过 Id
func (dao *CartDao) GetCartById(pId, uId, bId uint) (cart *model.Cart, err error) {
	err = dao.DB.Model(&model.Cart{}).
		Where("user_id=? AND product_id=? AND boss_id=?", uId, pId, bId).
		First(&cart).Error
	return
}

// ListCartByUserId 获取 Cart 通过 user_id
func (dao *CartDao) ListCartByUserId(uId uint) (cart []*model.Cart, err error) {
	err = dao.DB.Model(&model.Cart{}).
		Where("user_id=?", uId).Find(&cart).Error
	return
}

// UpdateCartNumById 通过id更新Cart信息
func (dao *CartDao) UpdateCartNumById(cId uint, num int) error {
	return dao.DB.Model(&model.Cart{}).
		Where("id=?", cId).Update("num", num).Error
}

// DeleteCartById 通过 cart_id 删除 cart
func (dao *CartDao) DeleteCartById(cId uint) error {
	return dao.DB.Model(&model.Cart{}).
		Where("id=?", cId).Delete(&model.Cart{}).Error
}

func (dao *CartDao) CartExistsById(cId uint) (bool, error) {
	var cart model.Cart
	// 查询购物车项是否存在
	if err := dao.DB.Where("id = ?", cId).First(&cart).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil // 表示不存在
		}
		return false, err // 查询出错
	}
	return true, nil // 存在
}
