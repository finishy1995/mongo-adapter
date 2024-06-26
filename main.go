package main

import (
	"finishy1995/mongo-adapter/library/log"
	"finishy1995/mongo-adapter/network"
)

func main() {
	log.SetLevel(log.DEBUG)
	network.NewServerAndMustStart("127.0.0.1:27017")
}
