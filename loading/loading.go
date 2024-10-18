package loading

import (
	"log"
	util "mall/pkg/utils"
	"mall/repository/cache"
	"mall/repository/db/dao"
	"mall/repository/mq"
)

func Loading() {
	// es.InitEs() // 如果需要接入ELK可以打开这个注释
	util.InitLog()
	dao.InitMySQL()
	cache.InitCache()
	mq.InitRabbitMQ() // 如果需要接入RabbitMQ可以打开这个注释

	// 启动 RabbitMQ 消费者，监听秒杀队列
	go mq.ReceiveSecKillGoodsFromMQ()
	log.Println("RabbitMQ 消费者已启动...")

	go scriptStarting()
}

func scriptStarting() {
	// 启动一些脚本
}
