package model

import "github.com/jinzhu/gorm"

// 购物车模型
type Cart struct {
	gorm.Model
	UserID    uint
	ProductID uint `gorm:"not null"`
	BossID    uint
	Num       int
	MaxNum    int
	Check     bool
}
