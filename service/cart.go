package service

import (
	"context"
	"strconv"

	"mall/pkg/e"
	"mall/pkg/errorhandler"
	util "mall/pkg/utils"
	dao2 "mall/repository/db/dao"
	"mall/repository/db/model"
	"mall/serializer"

	"github.com/sirupsen/logrus"
)

// CartService 创建购物车
type CartService struct {
	Id        uint `form:"id" json:"id"`
	BossID    uint `form:"boss_id" json:"boss_id"`
	ProductId uint `form:"product_id" json:"product_id"`
	Num       int  `form:"num" json:"num"`
}

func (service *CartService) Create(ctx context.Context, uId uint) serializer.Response {
	var product *model.Product
	code := e.SUCCESS

	// 判断有无这个商品
	productDao := dao2.NewProductDao(ctx)
	product, err := productDao.GetProductById(service.ProductId)
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":   err,
			"context": "CartCreate", // 描述了日志记录发生的上下文或服务
		}).Error("Failed to get product by Id")
		code = e.ErrorDatabase
		return errorhandler.HandleError(err, code, nil)

	}

	// 检查并设置商品数量
	num := 1 // 商品数量默认为1
	if service.Num > 1 {
		num = service.Num
	}

	// 创建购物车
	cartDao := dao2.NewCartDao(ctx)
	cart, status, _ := cartDao.CreateCart(service.ProductId, uId, product.BossID, num)
	if status == e.ErrorProductMoreCart {
		return serializer.Response{
			Status: status,
			Msg:    e.GetMsg(status),
		}
	}

	userDao := dao2.NewUserDao(ctx)
	boss, _ := userDao.GetUserById(product.BossID)
	return serializer.Response{
		Status: status,
		Msg:    e.GetMsg(status),
		Data:   serializer.BuildCart(cart, product, boss),
	}
}

// Show 购物车
func (service *CartService) Show(ctx context.Context, uId uint) serializer.Response {
	code := e.SUCCESS
	cartDao := dao2.NewCartDao(ctx)
	carts, err := cartDao.ListCartByUserId(uId)
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":   err,
			"context": "CartShow", // 描述了日志记录发生的上下文或服务
		}).Error("Failed to get cart by user_id")
		code = e.ErrorDatabase
		return errorhandler.HandleError(err, code, nil)

	}
	return serializer.Response{
		Status: code,
		Data:   serializer.BuildCarts(carts),
		Msg:    e.GetMsg(code),
	}
}

// Update 修改购物车信息
func (service *CartService) Update(ctx context.Context, cId string) serializer.Response {
	code := e.SUCCESS
	cartId, _ := strconv.Atoi(cId)

	cartDao := dao2.NewCartDao(ctx)

	if service.Num <= 0 {
		code = e.InvalidParams
		return errorhandler.HandleError(nil, code, nil)

	} else if service.Num > 10 {
		code = e.ErrorProductMoreCart
		return errorhandler.HandleError(nil, code, nil)

	}

	err := cartDao.UpdateCartNumById(uint(cartId), service.Num)
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":   err,
			"context": "CartUpdate", // 描述了日志记录发生的上下文或服务
		}).Error("Failed to update cart by Id")
		code = e.ErrorDatabase
		return errorhandler.HandleError(err, code, nil)

	}

	return serializer.Response{
		Status: code,
		Msg:    e.GetMsg(code),
	}
}

// 删除购物车
func (service *CartService) Delete(ctx context.Context, cId string) serializer.Response {
	code := e.SUCCESS
	cartDao := dao2.NewCartDao(ctx)
	cartId, _ := strconv.Atoi(cId)

	// 使用新的 CartExistsById 方法检查是否存在该 cId 的购物车项
	exists, err := cartDao.CartExistsById(uint(cartId)) // 调用 CartExistsById 方法
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":   err,
			"context": "CartDelete", // 描述了日志记录发生的上下文或服务
		}).Error("Failed to check if cart item exists")
		code = e.ErrorDatabase
		return errorhandler.HandleError(err, code, nil)
	}
	if !exists {
		util.LogrusObj.WithFields(logrus.Fields{
			"cartId":  cartId,
			"context": "CartDelete",
		}).Warn("Cart item not found")
		code = e.ErrorCartNotExist
		return errorhandler.HandleError(nil, code, nil)
	}

	err = cartDao.DeleteCartById(uint(cartId))
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":   err,
			"context": "CartDelete", // 描述了日志记录发生的上下文或服务
		}).Error("Failed to delete cart by cart_id")
		code = e.ErrorDatabase
		// return serializer.Response{
		// 	Status: code,
		// 	Msg:    e.GetMsg(code),
		// 	Error:  err.Error(),
		// }
		return errorhandler.HandleError(err, code, nil)

	}
	return serializer.Response{
		Status: code,
		Msg:    e.GetMsg(code),
	}
}
