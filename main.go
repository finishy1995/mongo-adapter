package main

import (
	"flag"
	"fmt"
	"os"

	"finishy1995/mongo-adapter/library/log"
	"finishy1995/mongo-adapter/network"
	"finishy1995/mongo-adapter/protocol"
)

func main() {
	// 定义命令行参数
	logLevel := flag.String("loglevel", "INFO", "log level: DEBUG, INFO, WARN, ERROR")
	uri := flag.String("uri", "", "MongoDB URI (e.g., mongodb://user:pass@host/db)")
	listenAddr := flag.String("listen", "0.0.0.0:27017", "listen address (e.g., 0.0.0.0:27017)")
	// 对外服务暴露的地址，即外面的应用链接这个服务，访问哪个地址
	exposeAddr := flag.String("expose", "127.0.0.1:27017", "expose address (e.g., 1.1.1.1:27017)")

	flag.Parse()

	if *uri == "" {
		fmt.Fprintf(os.Stderr, "Error: -uri parameter is required\n")
		flag.Usage()
		os.Exit(1)
	}

	// 设置日志等级
	log.SetLevelWithString(*logLevel)
	// 注册 MongoDB
	protocol.RegisterMongoDBByURI(*uri)
	// 注册对外服务暴露的地址
	protocol.RegisterExposeAddress(*exposeAddr)
	// 启动服务
	network.NewServerAndMustStart(*listenAddr, protocol.NewServer())
}
