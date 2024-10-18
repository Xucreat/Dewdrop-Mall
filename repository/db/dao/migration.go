package dao

import (
	"fmt"
	util "mall/pkg/utils"
	model2 "mall/repository/db/model"
	"os"

	"github.com/sirupsen/logrus"
)

// Migration 执行数据迁移
func Migration() {
	//自动迁移模式
	err := _db.Set("gorm:table_options", "charset=utf8mb4").
		AutoMigrate(&model2.User{},
			&model2.Product{},
			&model2.Carousel{},
			&model2.Category{},
			&model2.FavoriteStore{},
			&model2.ProductImg{},
			&model2.Order{},
			&model2.Cart{},
			&model2.Admin{},
			&model2.Address{},
			&model2.Notice{},
			&model2.SkillGoods{})
	// 增加详细错误日志输出
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error": err,
		}).Error("register table fail")
		fmt.Println("register table fail:", err) // 打印具体错误信息
		os.Exit(1)                               // 退出程序，错误码设为1
	}
	fmt.Println("register table success")
	util.LogrusObj.Info("register table success")
}
