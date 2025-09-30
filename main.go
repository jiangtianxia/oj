package main

import (
	"oj/router"
)

func main() {
	r := router.Router()
	r.Run("127.0.0.1:8081") // 监听并在 0.0.0.0:8080 上启动服务
}


