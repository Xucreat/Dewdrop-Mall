package service
/*封装服务初始化逻辑，启动消息消费者*/
import (
	"log"
	"mall/repository/mq"
)

// InitServices 初始化所有服务
func InitServices() {
	// 初始化其他服务...

	// 启动 RabbitMQ 消费者
	go mq.ReceiveSecKillGoodsFromMQ() // 启动消息消费者
	log.Println("RabbitMQ 消费者已启动...")
}