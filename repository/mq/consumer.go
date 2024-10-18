package mq

/*增加消息确认机制（ACK）*/
import (
	"context"
	"encoding/json"
	util "mall/pkg/utils"
	"mall/repository/db/model"
)

func ReceiveSecKillGoodsFromMQ() {
	ch, err := RabbitMQ.Channel()
	if err != nil {
		util.LogrusObj.Error("Failed to open a channel")
		return
	}
	defer ch.Close()

	msgs, err := ch.Consume(
		"skill_goods",
		"",
		false, // 手动确认:消费者消费消息后，在消息成功处理后，才通过 Ack 方法手动确认。
		false, // 非排他
		false, // 非本地队列
		false, // 不等待其他消费者
		nil,   // 额外参数
	)
	if err != nil {
		util.LogrusObj.Error("Failed to register a consumer")
		return
	}

	for msg := range msgs {
		sk := model.SkillGood2MQ{}
		if err := json.Unmarshal(msg.Body, &sk); err != nil {
			util.LogrusObj.WithField("error", err).Error("Failed to unmarshall message")
			msg.Nack(false, true) // 消息处理失败,重新放回队列，等待重试
			continue
		}

		// 创建上下文对象
		ctx := context.Background()

		// 调用业务处理函数
		if err := HandleSecKillGoods(ctx, &sk); err != nil {
			util.LogrusObj.WithField("skill_good_id", sk.SkillGoodId).Error("Failed to process seckill goods")
			msg.Nack(false, true) // 处理失败, 重新放回队列
			continue
		}

		// 消费完成后手动发送ACK确认
		msg.Ack(false)
		util.LogrusObj.WithField("skill_good_id", sk.SkillGoodId).Info("Seckill goods processed successfully")
	}

}
