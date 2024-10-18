package mq

import (
	"context"
	"errors"
	"fmt"
	util "mall/pkg/utils"
	"mall/repository/cache"
	"mall/repository/db/dao"
	"mall/repository/db/model"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/exp/rand"
)

func HandleSecKillGoods(ctx context.Context, sk *model.SkillGood2MQ) error {
	//  1. 检查库存 (从 Redis 中获取库存)
	stockKey := "SK" + strconv.Itoa(int(sk.SkillGoodId))
	num, err := cache.RedisClient.HGet(stockKey, "num").Int()
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"skill_good_id": sk.SkillGoodId,
			"error":         err,
		}).Error("Failed to get stock from Redis")
		return err
	}

	if num <= 0 {
		util.LogrusObj.WithFields(logrus.Fields{
			"skill_good_id": sk.SkillGoodId,
		}).Error("Out of stock")
		return errors.New("out of stock")
	}

	// 2. 创建订单 (在数据库中记录订单信息)
	order := &model.Order{
		ProductID:   sk.ProductId,       // 商品ID
		UserID:      sk.UserId,          // 用户ID
		AddressID:   sk.AddressId,       // 地址ID
		SkillGoodID: sk.SkillGoodId,     // 秒杀商品ID
		Money:       sk.Money,           // 秒杀商品价格
		Num:         1,                  // 购买数量，假设是1，可以根据业务调整
		OrderNum:    generateOrderNum(), // 生成唯一订单号
		Type:        1,                  // 未支付状态
	}

	// 使用 OrderDao 创建订单
	if err := dao.NewOrderDao(ctx).CreateOrder(order); err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"skill_good_id": sk.SkillGoodId,
			"error":         err,
		}).Error("Failed to create order")
		return err
	}

	// 3. 更新库存 (减少库存数量)
	newNum := num - 1
	_, err = cache.RedisClient.HSet(stockKey, "num", newNum).Result()
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"skill_good_id": sk.SkillGoodId,
			"error":         err,
		}).Error("Failed to update stock in Redis")
		return err
	}

	// 记录库存更新后的日志
	util.LogrusObj.WithFields(logrus.Fields{
		"skill_good_id":    sk.SkillGoodId,
		"remaining_stock":  newNum,
		"created_order_id": order.ID,
	}).Info("Stock updated and order created successfully")

	return nil

}

// generateOrderNum 生成唯一的订单号
func generateOrderNum() string {
	// 模拟生成唯一订单号，这里可以改成更复杂的逻辑
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
		randomPart := rand.New(rand.NewSource(uint64(time.Now().UnixNano()))).Uint64()
		return fmt.Sprintf("%s%012d", dateStr, randomPart)
	}
	// 将订单号格式化为流水号，比如 "000000000001"
	counterStr := fmt.Sprintf("%012d", orderID)
	// 生成最终的订单号，比如 "202301010000000000001"
	orderNum := fmt.Sprintf("%s%s", dateStr, counterStr)
	return orderNum
}
