package model

import (
	"github.com/jinzhu/gorm"
)

// Order 订单信息
type Order struct {
	gorm.Model
	UserID    uint    `gorm:"not null"`
	ProductID uint    `gorm:"not null"`
	BossID    uint    `gorm:"not null"`
	AddressID uint    `gorm:"not null"`
	Num       int     // 数量
	OrderNum  string  // 订单号
	Type      uint    // 1 未支付  2 已支付
	Money     float64 //

	SkillGoodID uint
}

// TableName 自定义表名
// 使用GORM对Order模型进行查询、创建、更新或删除操作时，GORM会自动使用TableName方法返回的字符串作为操作的表名
func (Order) TableName() string {
	return "user_order" // 指定模型对应的表名
}
