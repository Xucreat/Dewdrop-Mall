package mq

import (
	"fmt"
	"time"

	"github.com/streadway/amqp"

	"mall/conf"
)

// RabbitMQ rabbitMQ链接单例
var RabbitMQ *amqp.Connection

// InitRabbitMQ 在中间件中初始化rabbitMQ链接
func InitRabbitMQ() {
	// 组装完整的 RabbitMQ 连接字符串
	pathRabbitMQ := fmt.Sprintf("%s://%s:%s@%s:%s/",
		conf.RabbitMQ,         // Protocol (例如 "amqp")
		conf.RabbitMQUser,     // 用户名
		conf.RabbitMQPassWord, // 密码
		conf.RabbitMQHost,     // 主机
		conf.RabbitMQPort,     // 端口
	)
	conn, err := amqp.Dial(pathRabbitMQ)
	if err != nil {
		panic(err)
	}
	RabbitMQ = conn
}

/* RabbitMQ 消息发送和接收的处理逻辑 */
// 发送延迟消息
func SendDelayedMessage(queue string, message string, delay time.Duration) error {
	ch, err := RabbitMQ.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"order_delay_exchange", "x-delayed-message", true, false, false, false,
		amqp.Table{"x-delayed-type": "direct"},
	)
	if err != nil {
		return err
	}

	// 声明队列并设置延迟
	err = ch.ExchangeDeclare(
		"order_delay_exchange", "x-delayed-message", true, false, false, false,
		amqp.Table{"x-delayed-type": "direct"},
	)
	if err != nil {
		return err
	}

	headers := amqp.Table{"x-delay": int32(delay.Milliseconds())}
	err = ch.Publish(
		"order_delay_exchange", queue, false, false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
			Headers:     headers,
		},
	)
	return err
}

// 消息接收
func ReceiveMessage(queue string, handler func(string)) {
	ch, err := RabbitMQ.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	msgs, err := ch.Consume(queue, "", true, false, false, false, nil)
	if err != nil {
		panic(err)
	}

	go func() {
		for d := range msgs {
			handler(string(d.Body))
		}
	}()
}
