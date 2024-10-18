package dao

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"mall/repository/db/model"
)

type UserDao struct {
	*gorm.DB
}

func NewUserDao(ctx context.Context) *UserDao {
	return &UserDao{NewDBClient(ctx)}
}

func NewUserDaoByDB(db *gorm.DB) *UserDao {
	return &UserDao{db}
}

// GetUserById 根据 id 获取用户
func (dao *UserDao) GetUserById(uId uint) (user *model.User, err error) {
	err = dao.DB.Model(&model.User{}).Where("id=?", uId).
		First(&user).Error
	return
}

// UpdateUserById 根据 id 更新用户信息
func (dao *UserDao) UpdateUserById(uId uint, user *model.User) (err error) {
	return dao.DB.Model(&model.User{}).Where("id=?", uId).
		Updates(&user).Error
}

func (dao *UserDao) UpdateUserEmailById(uId uint, user *model.User) (err error) {
	// 使用 `Update` 显式更新 Email 字段
	return dao.DB.Model(&model.User{}).Where("id = ?", uId).
		Select("Email", "Password", "UpdatedAt").Updates(&user).Error
}

// ExistOrNotByUserName 根据username判断是否存在该名字
// 优化:减少数据库查询次数
func (dao UserDao) ExistOrNotByUserName(userName string) (user *model.User, exist bool, err error) {
	user = &model.User{} // 提前分配user对象，避免空指针问题
	err = dao.DB.Where("user_name = ?", userName).First(user).Error

	if err != nil {
		// 如果查询不到记录，返回 false，并忽略记录未找到的错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		// 返回其他可能的错误，比如数据库连接错误等
		return nil, false, err
	}
	// 如果查询成功，返回用户信息和存在标记
	return user, true, nil
}

// func (dao *UserDao) ExistOrNotByUserName(userName string) (user *model.User, exist bool, err error) {
// 	var count int64
// 	err = dao.DB.Model(&model.User{}).Where("user_name=?", userName).Count(&count).Error
// 	if count == 0 {
// 		return user, false, err
// 	}
// 	err = dao.DB.Model(&model.User{}).Where("user_name=?", userName).First(&user).Error
// 	if err != nil {
// 		return user, false, err
// 	}
// 	return user, true, nil
// }

// CreateUser 创建用户
func (dao *UserDao) CreateUser(user *model.User) error {
	return dao.DB.Model(&model.User{}).Create(&user).Error
}
