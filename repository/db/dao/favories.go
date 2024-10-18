package dao

import (
	// "context"

	"context"

	"gorm.io/gorm"

	"mall/repository/db/model"
)

type FavoriteStoreDao struct {
	*gorm.DB
}

func NewFavoriteStoreDao(ctx context.Context) *FavoriteStoreDao {
	return &FavoriteStoreDao{NewDBClient(ctx)}
}

// func NewFavoritesDaoByDB(db *gorm.DB) *FavoritesDao {
// 	return &FavoritesDao{db}
// }

// // ListFavoriteByUserId 通过 user_id 获取收藏夹列表
// func (dao *FavoritesDao) ListFavoriteByUserId(uId uint, pageSize, pageNum int) (favorites []*model.Favorite, total int64, err error) {
// 	// 总数
// 	err = dao.DB.Model(&model.Favorite{}).Preload("User").
// 		Where("user_id=?", uId).Count(&total).Error
// 	if err != nil {
// 		return
// 	}
// 	// 分页
// 	err = dao.DB.Model(model.Favorite{}).Preload("User").Where("user_id=?", uId).
// 		Offset((pageNum - 1) * pageSize).
// 		Limit(pageSize).Find(&favorites).Error
// 	return
// }

// // CreateFavorite 创建收藏夹
// func (dao *FavoritesDao) CreateFavorite(favorite *model.Favorite) (err error) {
// 	err = dao.DB.Create(&favorite).Error
// 	return
// }

// // FavoriteExistOrNot 判断是否存在
// func (dao *FavoritesDao) FavoriteStoreExistOrNot(pId, uId uint) (exist bool, err error) {
// 	var count int64
// 	err = dao.DB.Model(&model.Favorite{}).
// 		Where("product_id=? AND user_id=?", pId, uId).Count(&count).Error
// 	if count == 0 || err != nil {
// 		return false, err
// 	}
// 	return true, err

// }

// // DeleteFavoriteById 删除收藏夹
// func (dao *FavoritesDao) DeleteFavoriteById(fId uint) error {
// 	return dao.DB.Where("id=?", fId).Delete(&model.Favorite{}).Error
// }

// FavoriteStoreExistOrNot 判断店铺是否已经被收藏
func (dao *FavoriteStoreDao) FavoriteStoreExistOrNot(bossId, uId uint) (exist bool, err error) {
	var count int64
	err = dao.DB.Model(&model.FavoriteStore{}).
		Where("boss_id=? AND user_id=?", bossId, uId).Count(&count).Error
	if count == 0 || err != nil {
		return false, err
	}
	return true, err
}

// CreateFavorite 创建收藏店铺
func (dao *FavoriteStoreDao) CreateFavoriteStore(favorite *model.FavoriteStore) (err error) {
	err = dao.DB.Create(&favorite).Error
	return
}

// ListFavoriteByUserId 通过 user_id 获取收藏的店铺列表
func (dao *FavoriteStoreDao) ListFavoriteStoresByUserId(uId uint, pageSize, pageNum int) (favorites []*model.FavoriteStore, total int64, err error) {
	// 获取总数
	err = dao.DB.Model(&model.FavoriteStore{}).Preload("Boss"). // 确保预加载店铺信息
									Where("user_id=?", uId).Count(&total).Error
	if err != nil {
		return
	}
	// 分页查询
	err = dao.DB.Model(&model.FavoriteStore{}).Preload("Boss").Where("user_id=?", uId).
		Offset((pageNum - 1) * pageSize).
		Limit(pageSize).Find(&favorites).Error
	return
}

// DeleteFavoriteByBossAndUserId 通过 boss_id 和 user_id 删除收藏的店铺
func (dao *FavoriteStoreDao) DeleteFavoriteByBossAndUserId(bossId, userId uint) error {
	return dao.DB.Where("boss_id=? AND user_id=?", bossId, userId).Delete(&model.FavoriteStore{}).Error
}

// DeleteFavoritesByBossIdsAndUserId 根据多个 boss_id 和 user_id 批量删除收藏的店铺
func (dao *FavoriteStoreDao) DeleteFavoriteStoresByBossAndUserId(bossIds []uint, userId uint) error {
	return dao.DB.Where("boss_id IN (?) AND user_id=?", bossIds, userId).Delete(&model.FavoriteStore{}).Error
}
