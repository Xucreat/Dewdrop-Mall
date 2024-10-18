package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"mime/multipart"
	"strconv"
	"time"

	"mall/pkg/e"
	util "mall/pkg/utils"
	"mall/repository/cache"
	"mall/repository/db/dao"
	model2 "mall/repository/db/model"
	"mall/repository/mq"
	"mall/serializer"

	xlsx "github.com/360EntSecGroup-Skylar/excelize"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type SkillGoodsImport struct {
}

// 限购一个
type SkillGoodsService struct {
	SkillGoodsId uint   `json:"skill_goods_id" form:"skill_goods_id"`
	ProductId    uint   `json:"product_id" form:"product_id"`
	BossId       uint   `json:"boss_id" form:"boss_id"`
	AddressId    uint   `json:"address_id" form:"address_id"`
	Key          string `json:"key" form:"key"`
}

func (service *SkillGoodsImport) Import(ctx context.Context, file multipart.File) serializer.Response {
	xlFile, err := xlsx.OpenReader(file)
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":   err,
			"context": "Import", // 描述了日志记录发生的上下文或服务
		}).Error("Failed to open reader")
	}
	code := e.SUCCESS
	rows := xlFile.GetRows("Sheet1")
	length := len(rows[1:])
	skillGoods := make([]*model2.SkillGoods, length)
	for index, colCell := range rows {
		if index == 0 {
			continue
		}
		pId, _ := strconv.Atoi(colCell[0])
		bId, _ := strconv.Atoi(colCell[1])
		num, _ := strconv.Atoi(colCell[3])
		money, _ := strconv.ParseFloat(colCell[4], 64)
		skillGood := &model2.SkillGoods{
			ProductId: uint(pId),
			BossId:    uint(bId),
			Title:     colCell[2],
			Money:     money,
			Num:       num,
		}
		skillGoods[index-1] = skillGood
	}
	err = dao.NewSkillGoodsDao(ctx).CreateByList(skillGoods)
	if err != nil {
		code = e.ERROR
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Data:   "上传失败",
		}
	}
	return serializer.Response{
		Status: code,
		Msg:    e.GetMsg(code),
	}
}

// 直接放到这里，初始化秒杀商品信息，将mysql的信息存入redis中
func (service *SkillGoodsService) InitSkillGoods(ctx context.Context) error {
	skillGoods, _ := dao.NewSkillGoodsDao(ctx).ListSkillGoods()
	r := cache.RedisClient
	// 加载到redis
	// for i := range skillGoods {
	// 	fmt.Println(*skillGoods[i])
	// 	r.HSet("SK"+strconv.Itoa(int(skillGoods[i].Id)), "num", skillGoods[i].Num)
	// 	r.HSet("SK"+strconv.Itoa(int(skillGoods[i].Id)), "money", skillGoods[i].Money)
	// }
	for i := range skillGoods {
		r.HMSet("SK"+strconv.Itoa(int(skillGoods[i].Id)), map[string]interface{}{
			"num":   skillGoods[i].Num,
			"money": skillGoods[i].Money,
		})
		util.LogrusObj.WithFields(logrus.Fields{
			"skill_goods_id": skillGoods[i].Id,
			"num":            skillGoods[i].Num,
			"money":          skillGoods[i].Money,
		}).Info("Loaded skill goods into Redis")
	}
	return nil
}

func (service *SkillGoodsService) SkillGoods(ctx context.Context, uId uint) serializer.Response {
	mo, err := cache.RedisClient.HGet("SK"+strconv.Itoa(int(service.SkillGoodsId)), "money").Float64()
	if err != nil || mo == 0 {
		util.LogrusObj.WithFields(logrus.Fields{
			"skill_goods_id": service.SkillGoodsId,
			"error":          err,
		}).Error("Failed to get money from Redis, reloading from DB")

		// 从数据库重新加载数据
		skillGood, dbErr := dao.NewSkillGoodsDao(ctx).GetSkillGoodById(service.SkillGoodsId)
		if dbErr != nil {
			util.LogrusObj.WithFields(logrus.Fields{
				"skill_goods_id": service.SkillGoodsId,
				"error":          dbErr,
			}).Error("Failed to reload skill goods from DB")
			return serializer.Response{
				Status: e.ERROR,
				Msg:    "商品不存在",
			}
		} else {
			// 如果从数据库成功加载商品信息，则更新 mo 变量
			mo = skillGood.Money
		}
	}
	sk := &model2.SkillGood2MQ{
		ProductId:   service.ProductId,
		BossId:      service.BossId,
		UserId:      uId,
		AddressId:   service.AddressId,
		Key:         service.Key,
		Money:       mo,
		SkillGoodId: service.SkillGoodsId,
	}
	err = RedissonSecKillGoods(sk)
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":   err,
			"user_id": uId,
		}).Error("Seckill operation failed")
		return serializer.Response{
			Status: e.ERROR,
			Msg:    "秒杀失败，请稍后重试",
		}
	}
	return serializer.Response{}
}

// 加锁
func RedissonSecKillGoods(sk *model2.SkillGood2MQ) error {
	p := strconv.Itoa(int(sk.ProductId))
	uuid := getUuid(p)

	// 使用 SetNX 设置锁，设置 3 秒的超时时间
	lockSuccess, err := cache.RedisClient.SetNX(p, uuid, time.Second*3).Result()
	if err != nil || !lockSuccess {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":      err,
			"product_id": sk.ProductId,
		}).Error("Failed to aquire lock")
		return errors.New("get lock fail")
	}

	// 调用处理逻辑，确保锁定期间执行任务
	defer func() {
		// 使用 Lua 脚本来确保原子操作，比较 UUID 并删除锁
		luaScript := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
		`
		_, err = cache.RedisClient.Eval(luaScript, []string{p}, uuid).Result()
		if err != nil {
			util.LogrusObj.WithFields(logrus.Fields{
				"error":      err,
				"product_id": sk.ProductId,
			}).Error("Failed to release lock")
		} else {
			util.LogrusObj.WithFields(logrus.Fields{
				"product_id": sk.ProductId,
			}).Error("Successfully released lock")
		}
	}()

	util.LogrusObj.WithFields(logrus.Fields{
		"product_id": sk.ProductId,
	}).Info("Successfully acquired lock")

	// 处理业务逻辑
	err = SendSecKillGoodsToMQ(sk)
	if err != nil {
		return err
	}

	return nil
}

// 传送到MQ
func SendSecKillGoodsToMQ(sk *model2.SkillGood2MQ) error {
	// 获取 RabbitMQ 通道
	ch, err := mq.RabbitMQ.Channel()
	if err != nil {
		return fmt.Errorf("failed to get RabbitMQ channel: %v", err)
	}
	defer ch.Close() // 确保通道在使用完后关闭

	// 声明队列
	q, err := ch.QueueDeclare(
		"skill_goods", // 队列名称
		true,          // 持久化队列
		false,         // 非自动删除队列
		false,         // 非排他队列
		false,         // 非自动删除队列
		nil,           // 额外属性
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %v", err)
	}
	// 序列化秒杀商品数据
	body, err := json.Marshal(sk)
	if err != nil {
		return fmt.Errorf("failed to marshal skill goods data: %v", err)
	}

	// 发布消息到队列
	err = ch.Publish(
		"",     // 交换机名称，空字符串代表默认交换机
		q.Name, // 队列名称
		true,   // mandatory，确保消息能被路由到队列
		false,  // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,    // 持久化消息
			ContentType:  "application/json", // 持久化消息
			Body:         body,               // 消息体
		})
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	// 记录消息发送成功日志
	util.LogrusObj.WithFields(logrus.Fields{
		"queue":   q.Name,
		"message": string(body),
	}).Info("Successfully sent message to RabbitMQ")
	return nil
}

func getUuid(gid string) string {
	codeLen := 8
	// 1. 定义原始字符串
	rawStr := "jkwangagDGFHGSERKILMJHSNOPQR546413890_"
	// 2. 定义一个buf，并且将buf交给bytes往buf中写数据
	buf := make([]byte, 0, codeLen)
	b := bytes.NewBuffer(buf)
	// 随机从中获取
	rand.New(rand.NewSource(time.Now().UnixNano()))
	for rawStrLen := len(rawStr); codeLen > 0; codeLen-- {
		randNum := rand.Intn(rawStrLen)
		b.WriteByte(rawStr[randNum])
	}
	return b.String() + gid
}

// 取消订单的操作,redis的商品回退
