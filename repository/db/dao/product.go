package dao

import (
	"context"

	"gorm.io/gorm"

	model2 "mall/repository/db/model"
)

type ProductDao struct {
	*gorm.DB
}

func NewProductDao(ctx context.Context) *ProductDao {
	return &ProductDao{NewDBClient(ctx)}
}

func NewProductDaoByDB(db *gorm.DB) *ProductDao {
	return &ProductDao{db}
}

// GetProductById 通过 id 获取product
func (dao *ProductDao) GetProductById(id uint) (product *model2.Product, err error) {
	err = dao.DB.Model(&model2.Product{}).Where("id=?", id).
		First(&product).Error
	return
}

// ListProductByCondition 获取商品列表
func (dao *ProductDao) ListProductByCondition(condition map[string]interface{}, page, pageSize int) (products []*model2.Product, err error) {
	// 计算偏移量
	offset := (page - 1) * pageSize

	// 分页查询
	// Go 支持在函数签名中声明返回值，避免在函数体内重复定义这些变量
	err = dao.DB.Where(condition).Offset(offset).Limit(pageSize).Find(&products).Error
	if err != nil {
		return nil, err
	}
	return products, nil
}

// CreateProduct 创建商品
func (dao *ProductDao) CreateProduct(product *model2.Product) error {
	return dao.DB.Model(&model2.Product{}).Create(&product).Error
}

// CountProductByCondition 根据情况获取商品的数量
func (dao *ProductDao) CountProductByCondition(condition map[string]interface{}) (total int64, err error) {
	err = dao.DB.Model(&model2.Product{}).Where(condition).Count(&total).Error
	return
}

// DeleteProduct 删除商品
func (dao *ProductDao) DeleteProduct(pId uint) error {
	return dao.DB.Model(&model2.Product{}).Delete(&model2.Product{}).Error
}

// UpdateProduct 更新商品
func (dao *ProductDao) UpdateProduct(pId uint, product *model2.Product) error {
	return dao.DB.Model(&model2.Product{}).Where("id=?", pId).
		Updates(&product).Error
}

// SearchProduct 搜索商品
func (dao *ProductDao) SearchProduct(info string, page, pageSize, categoryID int) (products []*model2.Product, err error) {
	err = dao.DB.Model(&model2.Product{}).
		Where("name LIKE ? OR info LIKE ? OR category_id = ?", "%"+info+"%", "%"+info+"%", categoryID).
		Offset((page - 1) * pageSize).
		Limit(pageSize).Find(&products).Error
	return
}
