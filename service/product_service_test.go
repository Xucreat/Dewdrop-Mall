package service

import (
	"context"
	"mall/conf"
	"mall/pkg/e"
	util "mall/pkg/utils"
	"mall/repository/db/dao"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestConcurrentListProducts(t *testing.T) {
	ctx := context.Background()

	// 初始化日志和数据库
	util.InitLog()
	conf.Init()
	dao.InitTestMySQL()

	productService := ProductService{
		CategoryID: 1,
		Page:       1,
		PageSize:   10,
	}

	// 并发数量
	numRequests := 100
	results := make(chan bool, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			response := productService.List(ctx)
			if response.Status == e.SUCCESS {
				results <- true
			} else {
				results <- false
			}
		}()
	}

	// 收集测试结果
	successCount := 0
	for i := 0; i < numRequests; i++ {
		if <-results {
			successCount++
		}
	}

	t.Logf("Successfully retrieved products: %d out of %d requests", successCount, numRequests)
	assert.Equal(t, numRequests, successCount)
}
