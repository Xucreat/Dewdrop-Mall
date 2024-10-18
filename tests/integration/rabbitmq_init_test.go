package integration

/*验证 RabbitMQ 连接初始化是否正常*/
import (
	"mall/repository/mq"
	"testing"
)

// 测试 RabbitMQ 初始化
func TestInitRabbitMQ(t *testing.T) {
	defer func() {
		if mq.RabbitMQ != nil {
			mq.RabbitMQ.Close()
		}
	}()

	// 初始化 RabbitMQ
	mq.InitRabbitMQ()

	// 验证 RabbitMQ 连接是否成功建立
	if mq.RabbitMQ == nil {
		t.Fatal("Failed to initialize RabbitMQ connection.")
	}

	t.Log("RabbitMQ connection initialized successfully.")
}
