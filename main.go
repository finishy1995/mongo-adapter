package main

import (
	"finishy1995/mongo-adapter/library/log"
	"finishy1995/mongo-adapter/network"
	"finishy1995/mongo-adapter/protocol"
)

func main() {
	log.SetLevel(log.DEBUG)
	protocol.RegisterMongoDBByURI("mongodb+srv://david:19950521sjtu@test.phc4r.mongodb.net/?retryWrites=true&w=majority&appName=test")
	network.NewServerAndMustStart("127.0.0.1:27017", protocol.NewServer())
}
