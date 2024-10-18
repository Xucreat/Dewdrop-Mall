package service

import (
	"context"
	"mall/pkg/e"
	util "mall/pkg/utils"

	dao2 "mall/repository/db/dao"
	"mall/repository/db/model"
	"mall/serializer"

	"github.com/sirupsen/logrus"
)

// import (
// 	"context"

// 	"mall/pkg/e"
// 	util "mall/pkg/utils"
// 	dao2 "mall/repository/db/dao"
// 	"mall/repository/db/model"
// 	"mall/serializer"

// 	"github.com/sirupsen/logrus"
// )

// type FavoritesService struct {
// 	ProductId  uint `form:"product_id" json:"product_id"`
// 	BossId     uint `form:"boss_id" json:"boss_id"`
// 	FavoriteId uint `form:"favorite_id" json:"favorite_id"`
// 	PageNum    int  `form:"pageNum"`
// 	PageSize   int  `form:"pageSize"`
// }

// // Show 商品收藏夹
// func (service *FavoritesService) Show(ctx context.Context, uId uint) serializer.Response {
// 	favoritesDao := dao2.NewFavoritesDao(ctx)
// 	code := e.SUCCESS
// 	if service.PageSize == 0 {
// 		service.PageSize = 15
// 	}
// 	favorites, total, err := favoritesDao.ListFavoriteByUserId(uId, service.PageSize, service.PageNum)
// 	if err != nil {
// 		util.LogrusObj.WithFields(logrus.Fields{
// 			"error":   err,
// 			"context": "FavoritesShow", // 描述了日志记录发生的服务
// 		}).Error("Failed to get product by Id")
// 		code = e.ErrorDatabase
// 		return serializer.Response{
// 			Status: code,
// 			Msg:    e.GetMsg(code),
// 			Error:  err.Error(),
// 		}
// 	}
// 	return serializer.BuildListResponse(serializer.BuildFavorites(ctx, favorites), uint(total))
// }

// // Create 创建收藏夹
// func (service *FavoritesService) Create(ctx context.Context, uId uint) serializer.Response {
// 	code := e.SUCCESS
// 	favoriteDao := dao2.NewFavoritesDao(ctx)
// 	exist, _ := favoriteDao.FavoriteExistOrNot(service.ProductId, uId)
// 	if exist {
// 		code = e.ErrorExistFavorite
// 		return serializer.Response{
// 			Status: code,
// 			Msg:    e.GetMsg(code),
// 		}
// 	}

// 	userDao := dao2.NewUserDao(ctx)
// 	user, err := userDao.GetUserById(uId)
// 	if err != nil {
// 		code = e.ErrorDatabase
// 		return serializer.Response{
// 			Status: code,
// 			Msg:    e.GetMsg(code),
// 		}
// 	}

// 	bossDao := dao2.NewUserDaoByDB(userDao.DB)
// 	boss, err := bossDao.GetUserById(service.BossId)
// 	if err != nil {
// 		code = e.ErrorDatabase
// 		return serializer.Response{
// 			Status: code,
// 			Msg:    e.GetMsg(code),
// 		}
// 	}

// 	productDao := dao2.NewProductDao(ctx)
// 	product, err := productDao.GetProductById(service.ProductId)
// 	if err != nil {
// 		code = e.ErrorDatabase
// 		return serializer.Response{
// 			Status: code,
// 			Msg:    e.GetMsg(code),
// 		}
// 	}

// 	favorite := &model.Favorite{
// 		UserID:    uId,
// 		User:      *user,
// 		ProductID: service.ProductId,
// 		Product:   *product,
// 		BossID:    service.BossId,
// 		Boss:      *boss,
// 	}
// 	// favoriteDao = dao2.NewFavoritesDaoByDB(favoriteDao.DB)
// 	err = favoriteDao.CreateFavorite(favorite)
// 	if err != nil {
// 		code = e.ErrorDatabase
// 		return serializer.Response{
// 			Status: code,
// 			Msg:    e.GetMsg(code),
// 		}
// 	}

// 	return serializer.Response{
// 		Status: code,
// 		Msg:    e.GetMsg(code),
// 	}
// }

// // Delete 删除收藏夹
// func (service *FavoritesService) Delete(ctx context.Context) serializer.Response {
// 	code := e.SUCCESS

// 	favoriteDao := dao2.NewFavoritesDao(ctx)
// 	err := favoriteDao.DeleteFavoriteById(service.FavoriteId)
// 	if err != nil {
// 		util.LogrusObj.WithFields(logrus.Fields{
// 			"error":   err,
// 			"context": "FavoritesDelete", // 描述了日志记录发生的服务
// 		}).Error("Failed to delete favorite by Id")
// 		code = e.ErrorDatabase
// 		return serializer.Response{
// 			Status: code,
// 			Data:   e.GetMsg(code),
// 			Error:  err.Error(),
// 		}
// 	}

// 	return serializer.Response{
// 		Status: code,
// 		Data:   e.GetMsg(code),
// 	}
// }

// FavoritesService 收藏店铺服务
type FavoritesService struct {
	BossId   uint `form:"boss_id" json:"boss_id"` // 要收藏的店铺（商家）ID
	PageNum  int  `form:"page_num" json:"page_num"`
	PageSize int  `form:"page_size" json:"page_size"`
}

// Create 收藏店铺
func (service *FavoritesService) Collect(ctx context.Context, uId uint) serializer.Response {
	code := e.SUCCESS

	// 检查是否已收藏该店铺
	favoriteDao := dao2.NewFavoriteStoreDao(ctx)
	exist, _ := favoriteDao.FavoriteStoreExistOrNot(service.BossId, uId)
	if exist {
		code = e.ErrorExistFavorite
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}
	}

	// 获取店铺（商家）信息
	bossDao := dao2.NewUserDao(ctx)
	boss, err := bossDao.GetUserById(service.BossId)
	if err != nil {
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}
	}

	// 创建收藏记录
	favorite := &model.FavoriteStore{
		UserID: uId,
		BossID: service.BossId,
		Boss:   *boss,
	}
	err = favoriteDao.CreateFavoriteStore(favorite)
	if err != nil {
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}
	}

	return serializer.Response{
		Status: code,
		Msg:    e.GetMsg(code),
	}
}

// Show 店铺收藏夹
func (service *FavoritesService) Show(ctx context.Context, uId uint) serializer.Response {
	favoritesDao := dao2.NewFavoriteStoreDao(ctx)
	code := e.SUCCESS

	// 设置分页大小
	if service.PageSize == 0 {
		service.PageSize = 15
	}

	// 获取用户收藏的店铺列表
	favorites, total, err := favoritesDao.ListFavoriteStoresByUserId(uId, service.PageSize, service.PageNum) // 更新为获取店铺收藏的方法
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":   err,
			"context": "FavoritesStoreShow", // 描述了日志记录发生的服务
		}).Error("Failed to get favorite stores by user Id")
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}

	// 构建响应数据，返回店铺收藏列表
	return serializer.BuildListResponse(serializer.BuildFavoriteStores(ctx, favorites), uint(total)) // 使用新的构建方法
}

// Delete 删除单个店铺收藏
func (service *FavoritesService) Delete(ctx context.Context, bossId uint, userId uint) serializer.Response {
	code := e.SUCCESS

	// 从收藏夹中删除店铺
	favoriteDao := dao2.NewFavoriteStoreDao(ctx)                     // 修改为正确的 DAO 名称
	err := favoriteDao.DeleteFavoriteByBossAndUserId(bossId, userId) // 使用正确的方法删除店铺收藏
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":   err,
			"context": "FavoritesStoreDelete", // 描述了日志记录发生的服务
		}).Error("Failed to delete favorite store by Boss ID and User ID")
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}

	return serializer.Response{
		Status: code,
		Msg:    e.GetMsg(code),
	}
}

// DeleteBatch 批量删除店铺收藏
func (service *FavoritesService) DeleteBatch(ctx context.Context, bossIds []uint, userId uint) serializer.Response {
	code := e.SUCCESS

	// 批量从收藏夹中删除店铺
	favoriteDao := dao2.NewFavoriteStoreDao(ctx)
	err := favoriteDao.DeleteFavoriteStoresByBossAndUserId(bossIds, userId) // 调用批量删除的方法
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":   err,
			"context": "FavoritesStoreDeleteBatch", // 描述了日志记录发生的服务
		}).Error("Failed to batch delete favorite stores by Boss IDs and User ID")
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}

	return serializer.Response{
		Status: code,
		Msg:    e.GetMsg(code),
	}
}
