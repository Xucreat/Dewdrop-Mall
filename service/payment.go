package service

/*1.代码结构清晰:将支付和收款逻辑封装成一个独立的函数 processPayment，提高了代码的可读性和可维护性。*/
import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"mall/pkg/e"
	"mall/pkg/errorhandler"
	util "mall/pkg/utils"
	"mall/repository/cache"
	dao2 "mall/repository/db/dao"
	model2 "mall/repository/db/model"
	"mall/serializer"

	"gorm.io/gorm"
)

type OrderPay struct {
	OrderId   uint    `form:"order_id" json:"order_id"`
	Money     float64 `form:"money" json:"money"`
	OrderNo   string  `form:"orderNo" json:"orderNo"`
	ProductID int     `form:"product_id" json:"product_id"`
	PayTime   string  `form:"payTime" json:"payTime" `
	Sign      string  `form:"sign" json:"sign" `
	BossID    int     `form:"boss_id" json:"boss_id"`
	BossName  string  `form:"boss_name" json:"boss_name"`
	Num       int     `form:"num" json:"num"`
	Key       string  `form:"key" json:"key"`
}

func (service *OrderPay) PayDown(ctx context.Context, uId uint) serializer.Response {
	code := e.SUCCESS

	err := dao2.NewOrderDao(ctx).Transaction(func(tx *gorm.DB) error {
		util.Encrypt.SetKey(service.Key)
		orderDao := dao2.NewOrderDaoByDB(tx)

		// 获取订单信息
		cacheKey := fmt.Sprintf("order:%d", service.OrderId)
		var order *model2.Order

		// 先从 Redis 缓存中获取订单信息
		orderData, err := cache.RedisClient.Get(cacheKey).Result()
		if err == nil && orderData != "" {
			// 反序列化订单数据
			order, err = serializer.ParseOrderFromCache(orderData)
			if err != nil {
				code = e.ErrorDatabase
				return fmt.Errorf("failed to parse order from cache: %v", err)
			}
		} else {
			// 如果缓存中没有，则从数据库中获取订单
			order, err = orderDao.GetOrderById(service.OrderId)
			if err != nil {
				code = e.ErrorDatabase
				return fmt.Errorf("failed to get order from database: %v", err)
			}
		}

		// 获取商品信息
		productDao := dao2.NewProductDaoByDB(tx)
		product, err := productDao.GetProductById(uint(order.ProductID))
		if err != nil {
			code = e.ErrorDatabase
			return fmt.Errorf("failed to get product: %v", err)
		}

		// 使用订单中的金额，确保与订单创建时一致
		totalAmount := order.Money

		// 获取用户和商家信息
		userDao := dao2.NewUserDaoByDB(tx)
		user, err := userDao.GetUserById(uId)
		if err != nil {
			code = e.ErrorDatabase
			return fmt.Errorf("failed to get user: %v", err)
		}

		boss, err := userDao.GetUserById(order.BossID)
		if err != nil {
			code = e.ErrorDatabase
			return fmt.Errorf("failed to get boss: %v", err)
		}

		// 3. 处理用户支付和商家收款
		if err := processPayment(user, boss, totalAmount, util.Encrypt, tx); err != nil {
			code = e.ErrorPay
			return fmt.Errorf("failed to process payment: %v", err)
		}

		// 4. 更新商品数量
		product.Num -= order.Num
		if err := productDao.UpdateProduct(uint(order.ProductID), product); err != nil {
			code = e.ErrorDatabase
			return fmt.Errorf("failed to update product quantity: %v", err)
		}

		// 5. 更新订单状态
		order.Type = 2 // 订单状态更新为已支付
		if err := orderDao.UpdateOrderById(order.ID, order); err != nil {
			code = e.ErrorDatabase
			return fmt.Errorf("failed to update order status: %v", err)
		}
		// // 6. 创建用户商品记录
		// productUser := model2.Product{
		// 	Name:          product.Name,
		// 	CategoryID:    product.CategoryID,
		// 	Title:         product.Title,
		// 	Info:          product.Info,
		// 	ImgPath:       product.ImgPath,
		// 	Price:         product.Price,
		// 	DiscountPrice: product.DiscountPrice,
		// 	Num:           order.Num,
		// 	OnSale:        false,
		// 	BossID:        uId,
		// 	BossName:      user.UserName,
		// 	BossAvatar:    user.Avatar,
		// }

		// if err := productDao.CreateProduct(&productUser); err != nil {
		// 	code = e.ErrorDatabase
		// 	return fmt.Errorf("failed to create user product record: %v", err)
		// }

		// 事务成功
		return nil
	})
	// 	// 对钱进行解密。减去订单。再进行加密。
	// 	// 解密用户金额并进行扣减
	// 	userMoney, err := strconv.ParseFloat(util.Encrypt.AesDecoding(user.Money), 64)
	// 	if handleError(&code, err) != nil {
	// 		return err
	// 	}

	// 	if userMoney < totalMoney { // 金额不足
	// 		util.LogrusObj.WithFields(logrus.Fields{
	// 			"error":   err,
	// 			"context": "userMoney < totalMoney", // 描述了日志记录发生的服务
	// 		}).Error("金币不足")
	// 		return errors.New("金币不足")
	// 	}

	// 	user.Money = util.Encrypt.AesEncoding(fmt.Sprintf("%f", userMoney-totalMoney))

	// 	// 更新用户金额
	// 	if err := userDao.UpdateUserById(uId, user); handleError(&code, err) != nil {
	// 		return err
	// 	}

	// 	// 获取并更新Boss金额
	// 	boss, err := userDao.GetUserById(uint(service.BossID))
	// 	if handleError(&code, err) != nil {
	// 		return err
	// 	}

	// 	bossMoney, err := strconv.ParseFloat(util.Encrypt.AesDecoding(boss.Money), 64)
	// 	if handleError(&code, err) != nil {
	// 		return err
	// 	}

	// 	boss.Money = util.Encrypt.AesEncoding(fmt.Sprintf("%f", bossMoney+totalMoney))

	// 	// 更新Boss金额
	// 	if err := userDao.UpdateUserById(uint(service.BossID), boss); handleError(&code, err) != nil {
	// 		return err
	// 	}

	// 	// 获取并更新商品库存
	// 	productDao := dao2.NewProductDaoByDB(tx)
	// 	product, err := productDao.GetProductById(uint(service.ProductID))
	// 	if handleError(&code, err) != nil {
	// 		return err
	// 	}

	// 	product.Num -= order.Num
	// 	if err := productDao.UpdateProduct(uint(service.ProductID), product); handleError(&code, err) != nil {
	// 		return err
	// 	}

	// 	// 更新订单状态
	// 	order.Type = 2
	// 	if err := orderDao.UpdateOrderById(service.OrderId, order); handleError(&code, err) != nil {
	// 		return err
	// 	}

	// 	// 创建新的商品记录
	// 	productUser := model2.Product{
	// 		Name:          product.Name,
	// 		CategoryID:    product.CategoryID,
	// 		Title:         product.Title,
	// 		Info:          product.Info,
	// 		ImgPath:       product.ImgPath,
	// 		Price:         product.Price,
	// 		DiscountPrice: product.DiscountPrice,
	// 		Num:           order.Num,
	// 		OnSale:        false,
	// 		BossID:        uId,
	// 		BossName:      user.UserName,
	// 		BossAvatar:    user.Avatar,
	// 	}
	// 	if err := productDao.CreateProduct(&productUser); handleError(&code, err) != nil {
	// 		return err
	// 	}

	// 	return nil

	if err != nil {
		return errorhandler.HandleError(err, code, nil)
	}

	return serializer.Response{
		Status: code,
		Msg:    e.GetMsg(code),
		Data:   "支付成功，感谢您的使用！",
	}
}

// 处理用户支付和商家收款
func processPayment(user, boss *model2.User, amount float64, encryptor *util.Encryption, tx *gorm.DB) error {
	// 解密用户余额
	// 解密用户余额
	userMoneyStr := encryptor.AesDecoding(user.Money)
	if userMoneyStr == "Decryption failed" {
		return errors.New("用户余额解密失败")
	}
	userMoney, err := strconv.ParseFloat(userMoneyStr, 64)
	if err != nil {
		return fmt.Errorf("用户余额解析失败: %v", err)
	} else if userMoney < amount {
		return errors.New("用户余额不足")
	}

	// 更新用户余额
	newUserMoney := userMoney - amount
	user.Money = encryptor.AesEncoding(fmt.Sprintf("%f", newUserMoney))

	// 解密商家余额
	bossMoneyStr := encryptor.AesDecoding(boss.Money)
	if bossMoneyStr == "Decryption failed" {
		return errors.New("商家余额解密失败")
	}
	bossMoney, err := strconv.ParseFloat(bossMoneyStr, 64)
	if err != nil {
		return fmt.Errorf("商家余额解析失败: %v", err)
	}

	// 更新商家余额
	newBossMoney := bossMoney + amount
	boss.Money = encryptor.AesEncoding(fmt.Sprintf("%f", newBossMoney))

	// 将更新后的余额存入数据库
	if err := tx.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user balance: %v", err)
	}
	if err := tx.Save(boss).Error; err != nil {
		return fmt.Errorf("failed to update boss balance: %v", err)
	}
	return nil
}
