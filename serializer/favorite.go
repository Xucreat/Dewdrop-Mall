package serializer

import (
	"context"

	// "mall/conf"
	dao2 "mall/repository/db/dao"
	model2 "mall/repository/db/model"
)

// type Favorite struct {
// 	UserID        uint   `json:"user_id"`
// 	ProductID     uint   `json:"product_id"`
// 	CreatedAt     int64  `json:"create_at"`
// 	Name          string `json:"name"`
// 	CategoryID    uint   `json:"category_id"`
// 	Title         string `json:"title"`
// 	Info          string `json:"info"`
// 	ImgPath       string `json:"img_path"`
// 	Price         string `json:"price"`
// 	DiscountPrice string `json:"discount_price"`
// 	BossID        uint   `json:"boss_id"`
// 	Num           int    `json:"num"`
// 	OnSale        bool   `json:"on_sale"`
// }

// // 序列化收藏夹
// func BuildFavorite(item1 *model2.Favorite, item2 *model2.Product, item3 *model2.User) Favorite {
// 	return Favorite{
// 		UserID:        item1.UserID,
// 		ProductID:     item1.ProductID,
// 		CreatedAt:     item1.CreatedAt.Unix(),
// 		Name:          item2.Name,
// 		CategoryID:    item2.CategoryID,
// 		Title:         item2.Title,
// 		Info:          item2.Info,
// 		ImgPath:       conf.PhotoHost + conf.HttpPort + conf.ProductPhotoPath + item2.ImgPath,
// 		Price:         item2.Price,
// 		DiscountPrice: item2.DiscountPrice,
// 		BossID:        item3.ID,
// 		Num:           item2.Num,
// 		OnSale:        item2.OnSale,
// 	}
// }

// // 收藏夹列表
// func BuildFavorites(ctx context.Context, items []*model2.Favorite) (favorites []Favorite) {
// 	productDao := dao2.NewProductDao(ctx)
// 	bossDao := dao2.NewUserDao(ctx)

// 	for _, fav := range items {
// 		product, err := productDao.GetProductById(fav.ProductID)
// 		if err != nil {
// 			continue
// 		}
// 		boss, err := bossDao.GetUserById(fav.UserID)
// 		if err != nil {
// 			continue
// 		}
// 		favorite := BuildFavorite(fav, product, boss)
// 		favorites = append(favorites, favorite)
// 	}
// 	return favorites
// }

type FavoriteStore struct {
	UserID     uint   `json:"user_id"`     // 用户 ID
	BossID     uint   `json:"boss_id"`     // 店铺（商家）ID
	BossName   string `json:"boss_name"`   // 店铺名称（商家名称）
	BossAvatar string `json:"boss_avatar"` // 店铺头像
	CreatedAt  int64  `json:"created_at"`  // 收藏创建时间
}

// 序列化店铺收藏
func BuildFavoriteStore(item1 *model2.FavoriteStore, item2 *model2.User) FavoriteStore {
	return FavoriteStore{
		UserID:     item1.UserID,
		BossID:     item2.ID,
		BossName:   item2.UserName,
		BossAvatar: item2.Avatar,
		CreatedAt:  item1.CreatedAt.Unix(),
	}
}

// 收藏店铺列表
func BuildFavoriteStores(ctx context.Context, items []*model2.FavoriteStore) (favorites []FavoriteStore) {
	bossDao := dao2.NewUserDao(ctx)

	for _, fav := range items {
		boss, err := bossDao.GetUserById(fav.BossID) // 根据商家 ID 获取商家信息
		if err != nil {
			continue
		}
		favorite := BuildFavoriteStore(fav, boss)
		favorites = append(favorites, favorite)
	}
	return favorites
}
