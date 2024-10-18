/*专用的消费者程序，监听超时订单的处理逻辑*/
package main

import (
	"context"
	"fmt"
	"mall/repository/db/dao"
	"mall/repository/mq"

	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("Starting Order Consumer...")
	mq.ReceiveMessage("order_timeout_queue", processExpiredOrder)
}

func processExpiredOrder(message string) {
	var orderNum uint64
	_, err := fmt.Sscanf(message, "Order:%d", &orderNum)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"msg":   message,
		}).Error("Failed to parse order number from message")
		return
	}

	ctx := context.Background()
	orderDao := dao.NewOrderDao(ctx)
	order, err := orderDao.GetOrderByOrderNum(orderNum)
	if err != nil || order.Type != 1 {
		return
	}

	order.Type = 3 // 取消订单
	err = orderDao.UpdateOrderById(order.ID, order)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":    err,
			"orderNum": orderNum,
		}).Error("Failed to update order to cancelled state")
	}
}
