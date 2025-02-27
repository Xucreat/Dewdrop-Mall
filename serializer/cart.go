package serializer

import (
	"context"

	"mall/conf"
	"mall/consts"
	dao2 "mall/repository/db/dao"
	model2 "mall/repository/db/model"
)

// 购物车
type Cart struct {
	ID            uint   `json:"id"`
	UserID        uint   `json:"user_id"`
	ProductID     uint   `json:"product_id"`
	CreateAt      int64  `json:"create_at"`
	Num           int    `json:"num"`
	MaxNum        int    `json:"max_num"`
	Check         bool   `json:"check"`
	Name          string `json:"name"`
	ImgPath       string `json:"img_path"`
	DiscountPrice string `json:"discount_price"`
	BossId        uint   `json:"boss_id"`
	BossName      string `json:"boss_name"`
	Desc          string `json:"desc"`
}

func BuildCart(cart *model2.Cart, product *model2.Product, boss *model2.User) Cart {
	c := Cart{
		ID:            cart.ID,
		UserID:        cart.UserID,
		ProductID:     cart.ProductID,
		CreateAt:      cart.CreatedAt.Unix(),
		Num:           cart.Num,
		MaxNum:        cart.MaxNum,
		Check:         cart.Check,
		Name:          product.Name,
		ImgPath:       conf.PhotoHost + conf.HttpPort + conf.ProductPhotoPath + product.ImgPath,
		DiscountPrice: product.DiscountPrice,
		BossId:        boss.ID,
		BossName:      boss.UserName,
		Desc:          product.Info,
	}
	if conf.UploadModel == consts.UploadModelOss {
		c.ImgPath = product.ImgPath
	}

	return c
}

func BuildCarts(items []*model2.Cart) (carts []Cart) {
	for _, item1 := range items {
		product, err := dao2.NewProductDao(context.Background()).
			GetProductById(item1.ProductID)
		if err != nil {
			continue
		}
		boss, err := dao2.NewUserDao(context.Background()).
			GetUserById(item1.BossID)
		if err != nil {
			continue
		}
		cart := BuildCart(item1, product, boss)
		carts = append(carts, cart)
	}
	return carts
}
