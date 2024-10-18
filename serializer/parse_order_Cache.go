package serializer

import (
	"encoding/json"
	"fmt"
	model2 "mall/repository/db/model"
)

// ParseOrderFromCache 从缓存中解析订单信息
func ParseOrderFromCache(orderData string) (*model2.Order, error) {
	var order model2.Order
	err := json.Unmarshal([]byte(orderData), &order)
	if err != nil {
		return nil, fmt.Errorf("failed to parse order from cache: %v", err)
	}
	return &order, nil
}
