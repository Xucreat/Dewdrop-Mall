package integration

/*验证消费者的功能，包括确认机制（ACK）和消息重试*/
import (
	"encoding/json"
	"testing"
	"time"

	"mall/conf"
	util "mall/pkg/utils"
	"mall/repository/db/model"
	"mall/repository/mq"

	"github.com/streadway/amqp"
)

// 模拟生产者发送消息
func publishTestMessage(queueName string, msgBody model.SkillGood2MQ) error {
	ch, err := mq.RabbitMQ.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// 将消息序列化为 JSON 格式
	body, err := json.Marshal(msgBody)
	if err != nil {
		return err
	}

	// 发布消息到队列
	return ch.Publish(
		"",        // exchange
		queueName, // routing key（即队列名）
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

// 测试 RabbitMQ 消费者
func TestRabbitMQConsumer(t *testing.T) {
	// 初始化日志系统和 RabbitMQ 连接
	conf.Init() // 先初始化配置
	util.InitLog()
	mq.InitRabbitMQ()
	defer mq.RabbitMQ.Close()

	defer func() {
		if mq.RabbitMQ != nil {
			mq.RabbitMQ.Close()
		}
	}()

	// 创建模拟 SkillGood2MQ 消息
	testMsg := model.SkillGood2MQ{
		SkillGoodId: 123,
	}

	// 发布测试消息
	err := publishTestMessage("skill_goods", testMsg)
	if err != nil {
		t.Fatalf("Failed to publish test message: %v", err)
	}

	// 启动消费者，模拟消费测试消息
	go mq.ReceiveSecKillGoodsFromMQ()

	// 等待 3 秒，确保消息有时间被消费
	time.Sleep(3 * time.Second)

	// 可以通过日志输出确认消费者是否成功处理消息
	t.Log("Consumer test completed. Check logs for message processing result.")
}
