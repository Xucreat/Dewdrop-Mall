package dao

import (
	"context"

	"gorm.io/gorm"

	"mall/repository/db/model"
)

type AddressDao struct {
	*gorm.DB
}

func NewAddressDao(ctx context.Context) *AddressDao {
	return &AddressDao{NewDBClient(ctx)}
}

func NewAddressDaoByDB(db *gorm.DB) *AddressDao {
	return &AddressDao{db}
}

// GetAddressByAid 根据 Address Id 获取 Address
func (dao *AddressDao) GetAddressByAid(aId uint) (address *model.Address, err error) {
	err = dao.DB.Model(&model.Address{}).
		Where("id = ?", aId).First(&address).
		Error
	return
}

// ListAddressByUid 根据 User Id 获取User
func (dao *AddressDao) ListAddressByUid(uid uint) (addressList []*model.Address, err error) {
	err = dao.DB.Model(&model.Address{}).
		Where("user_id=?", uid).Order("created_at desc").
		Find(&addressList).Error
	return
}

// CreateAddress 创建地址
func (dao *AddressDao) CreateAddress(address *model.Address) (err error) {
	err = dao.DB.Model(&model.Address{}).Create(&address).Error
	return
}

// DeleteAddressById 根据 id 删除地址
func (dao *AddressDao) DeleteAddressById(aId uint) (err error) {
	err = dao.DB.Where("id=?", aId).Delete(&model.Address{}).Error
	return
}

// UpdateAddressFieldsById 只更新指定的字段
func (dao *AddressDao) UpdateAddressFieldsById(aId uint, fields map[string]interface{}) (err error) {
	err = dao.DB.Model(&model.Address{}).
		Where("id=?", aId).Updates(fields).Error
	return
}
