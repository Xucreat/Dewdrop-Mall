package service

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"mall/pkg/e"
	"mall/pkg/errorhandler"
	util "mall/pkg/utils"
	"mall/repository/cache"
	dao2 "mall/repository/db/dao"
	model2 "mall/repository/db/model"
	"mall/serializer"

	"mall/repository/mq"

	"github.com/sirupsen/logrus"
)

const OrderTimeKey = "OrderTime"

type OrderService struct {
	ProductID uint `form:"product_id" json:"product_id"`
	Num       uint `form:"num" json:"num"`
	AddressID uint `form:"address_id" json:"address_id"`
	Money     int  `form:"money" json:"money"`
	BossID    uint `form:"boss_id" json:"boss_id"`
	UserID    uint `form:"user_id" json:"user_id"`
	OrderNum  uint `form:"order_num" json:"order_num"`
	Type      int  `form:"type" json:"type"`
	model2.BasePage
}

// func (service *OrderService) Create(ctx context.Context, id uint) serializer.Response {
// 	code := e.SUCCESS

// 	order := &model2.Order{
// 		UserID:    id,
// 		ProductID: service.ProductID,
// 		BossID:    service.BossID,
// 		Num:       int(service.Num),
// 		Money:     float64(service.Money),
// 		Type:      1,
// 	}
// 	addressDao := dao2.NewAddressDao(ctx)
// 	address, err := addressDao.GetAddressByAid(service.AddressID)
// 	if err != nil {
// 		util.LogrusObj.WithFields(logrus.Fields{
// 			"error":   err,
// 			"context": "OrderCreate", // 描述了日志记录发生的服务
// 		}).Error("Failed to get address by Id")
// 		code = e.ErrorDatabase
// 		return serializer.Response{
// 			Status: code,
// 			Msg:    e.GetMsg(code),
// 			Error:  err.Error(),
// 		}
// 	}

// 	order.AddressID = address.ID
// 	number := fmt.Sprintf("%09v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000000))
// 	productNum := strconv.Itoa(int(service.ProductID))
// 	userNum := strconv.Itoa(int(id))
// 	number = number + productNum + userNum
// 	// 将字符串转换为无符号整数（uint64)
// 	// base int: 字符串 s 中数字的基数（进制）
// 	// bitSize int: 指定结果必须能无溢出地存放在多少位宽的整数中
// 	orderNum, _ := strconv.ParseUint(number, 10, 64)
// 	order.OrderNum = orderNum

// 	orderDao := dao2.NewOrderDao(ctx)
// 	err = orderDao.CreateOrder(order)
// 	if err != nil {
// 		util.LogrusObj.WithFields(logrus.Fields{
// 			"error":   err,
// 			"context": "OrderCreate", // 描述了日志记录发生的服务
// 		}).Error("Failed to create order")
// 		code = e.ErrorDatabase
// 		return serializer.Response{
// 			Status: code,
// 			Msg:    e.GetMsg(code),
// 			Error:  err.Error(),
// 		}
// 	}

// 	// 订单号存入Redis中，设置过期时间
// 	data := redis.Z{
// 		Score:  float64(time.Now().Unix()) + 15*time.Minute.Seconds(),
// 		Member: orderNum,
// 	}
// 	cache.RedisClient.ZAdd(OrderTimeKey, data)
// 	return serializer.Response{
// 		Status: code,
// 		Msg:    e.GetMsg(code),
// 	}
// }

func (service *OrderService) Create(ctx context.Context, id uint) serializer.Response {
	code := e.SUCCESS
	product, err := dao2.NewProductDao(ctx).GetProductById(service.ProductID)
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":     err,
			"context":   "OrderCreate",
			"productID": service.ProductID,
		}).Error("Failed to get product by Id")
		code = e.ErrorDatabase
		return errorhandler.HandleError(err, code, nil)
	}
	// 计算订单总金额
	// 计算折扣后总金额
	discountPrice, _ := strconv.ParseFloat(product.DiscountPrice, 64)
	totalMoney := discountPrice * float64(service.Num)

	order := &model2.Order{
		UserID:    id,
		ProductID: service.ProductID,
		BossID:    product.BossID,
		Num:       int(service.Num),
		Money:     totalMoney,
		Type:      1, // 初始状态，表示订单已创建
	}

	// 获取用户地址信息
	addressDao := dao2.NewAddressDao(ctx)
	address, err := addressDao.GetAddressByAid(service.AddressID)
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":   err,
			"context": "OrderCreate",
			"userID":  id,
		}).Error("Failed to get address by Id")
		code = e.ErrorDatabase
		return errorhandler.HandleError(err, code, nil)
	}
	order.AddressID = address.ID

	// 生成订单号
	order.OrderNum = generateUniqueOrderNum()

	// 将订单存储到数据库
	orderDao := dao2.NewOrderDao(ctx)
	err = orderDao.CreateOrder(order)
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":   err,
			"context": "OrderCreate",
			"userID":  id,
		}).Error("Failed to create order")
		code = e.ErrorDatabase
		return errorhandler.HandleError(err, code, nil)

	}

	// 缓存订单信息到 Redis，有效期为 15 分钟
	cacheKey := fmt.Sprintf("order:%s", order.OrderNum)
	err = cache.RedisClient.Set(cacheKey, order.OrderNum, 15*time.Minute).Err()
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error": err,
			"order": order.OrderNum,
		}).Error("Failed to cache order to Redis")
	}

	// 将订单信息发送到消息队列，设置 15 分钟过期时间
	err = mq.SendDelayedMessage("order_timeout_queue", fmt.Sprintf("Order:%s", order.OrderNum), 15*time.Minute)
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error": err,
			"order": order.OrderNum,
		}).Error("Failed to send order timeout message")
	}

	return serializer.Response{
		Status: code,
		Msg:    e.GetMsg(code),
		Data:   serializer.BuildOrder(order, product, address),
	}
}

// func generateUniqueOrderNum() uint64 {
// 	// 使用 Redis 自增确保唯一性
// 	orderID, err := cache.RedisClient.Incr("order:counter").Result()
// 	if err != nil {
// 		// 处理错误，记录日志
// 		logrus.Error("Failed to generate unique order number:", err)
// 		return uint64(rand.New(rand.NewSource(time.Now().UnixNano())).Uint64()) // fallback
// 	}
// 	return uint64(orderID)
// }

func generateUniqueOrderNum() string {
	// 获取当前时间
	currentTime := time.Now()
	// 格式化日期，比如 "20230101"
	dateStr := currentTime.Format("20060102")
	// 使用 Redis 自增确保唯一性
	orderID, err := cache.RedisClient.Incr("order:counter").Result()
	if err != nil {
		// 处理错误，记录日志
		logrus.Error("Failed to generate unique order number:", err)
		// fallback: 使用随机数生成订单号
		randomPart := rand.New(rand.NewSource(time.Now().UnixNano())).Uint64()
		return fmt.Sprintf("%s%012d", dateStr, randomPart)
	}
	// 将订单号格式化为流水号，比如 "000000000001"
	counterStr := fmt.Sprintf("%012d", orderID)
	// 生成最终的订单号，比如 "202301010000000000001"
	orderNum := fmt.Sprintf("%s%s", dateStr, counterStr)
	return orderNum
}

func (service *OrderService) List(ctx context.Context, uId uint) serializer.Response {
	var orders []*model2.Order
	var total int64
	code := e.SUCCESS
	if service.PageSize == 0 {
		service.PageSize = 5
	}

	orderDao := dao2.NewOrderDao(ctx)
	condition := make(map[string]interface{})
	condition["user_id"] = uId

	if service.Type == 0 {
		condition["type"] = 0
	} else {
		condition["type"] = service.Type
	}
	orders, total, err := orderDao.ListOrderByCondition(condition, service.BasePage)
	if err != nil {
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}
	}

	return serializer.BuildListResponse(serializer.BuildOrders(ctx, orders), uint(total))
}

func (service *OrderService) Show(ctx context.Context, uId string) serializer.Response {
	code := e.SUCCESS

	orderId, _ := strconv.Atoi(uId)
	orderDao := dao2.NewOrderDao(ctx)
	order, _ := orderDao.GetOrderById(uint(orderId))

	addressDao := dao2.NewAddressDao(ctx)
	address, err := addressDao.GetAddressByAid(order.AddressID)
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":   err,
			"context": "OrderShow", // 描述了日志记录发生的服务
		}).Error("Failed to get address by address Id")
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}
	}

	productDao := dao2.NewProductDao(ctx)
	product, err := productDao.GetProductById(order.ProductID)
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":   err,
			"context": "OrderShow", // 描述了日志记录发生的服务
		}).Error("Failed to get product by Id")
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}
	}

	return serializer.Response{
		Status: code,
		Msg:    e.GetMsg(code),
		Data:   serializer.BuildOrder(order, product, address),
	}
}

func (service *OrderService) Delete(ctx context.Context, oId string) serializer.Response {
	code := e.SUCCESS

	orderDao := dao2.NewOrderDao(ctx)
	orderId, _ := strconv.Atoi(oId)
	err := orderDao.DeleteOrderById(uint(orderId))
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":   err,
			"context": "OrderDelete", // 描述了日志记录发生的服务
		}).Error("Failed to delete order by Id")
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
