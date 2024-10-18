package main

import (
	"log"
	"mall/conf"
	"mall/loading"
	"mall/routes"
)

func main() {
	// Ek1+Ep1==Ek2+Ep2
	conf.Init()
	// 加载所有服务，包括数据库、Redis、RabbitMQ等
	loading.Loading()

	// 启动 HTTP 服务器
	r := routes.NewRouter()
	log.Println("HTTP server starting on port", conf.HttpPort)
	_ = r.Run(conf.HttpPort)
}
