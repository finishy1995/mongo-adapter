package message

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Message interface {
	Accept(bson.M) bool
	Handle(bson.M) bson.M
}

var (
	messages = []Message{}
	db       *mongo.Client
)

func RegisterMessage(msg Message) {
	messages = append(messages, msg)
}

func RegisterMongoDB(client *mongo.Client) {
	db = client
}

func RegisterMongoDBByURI(uri string) error {
	var err error
	db, err = mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	return err
}

func HandleMessage(message bson.M) bson.M {
	for _, msg := range messages {
		if msg.Accept(message) {
			return msg.Handle(message)
		}
	}
	return bson.M{"ok": 0, "errmsg": "command not supported"}
}
