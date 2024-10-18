package cache

import (
	"sync"
)

type Cache struct {
	data sync.Map
}

// NewCache 创建新的缓存实例
func NewCache() *Cache {
	return &Cache{}
}

// Get 从缓存中获取解密后的金额
func (c *Cache) Get(userID uint) (float64, bool) {
	value, ok := c.data.Load(userID)
	if ok {
		return value.(float64), true
	}
	return 0, false
}

// Set 将解密后的金额存入缓存
func (c *Cache) Set(userID uint, amount float64) {
	c.data.Store(userID, amount)
}

// Delete 从缓存中删除加密后的金额
func (c *Cache) Delete(userID uint) {
	c.data.Delete(userID)
}
