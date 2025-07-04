package message

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	RegisterMessage(&helloMessage{})
}

type helloMessage struct {
}

func (h *helloMessage) Accept(message bson.M) bool {
	if _, ok := message["ismaster"]; ok {
		return true
	}
	return false
}

func (h *helloMessage) Handle(message bson.M) bson.M {
	if db == nil {
		return bson.M{"ok": 0, "errmsg": "MongoDB server connection error"}
	}

	var result bson.M
	err := db.Database("admin").RunCommand(context.TODO(), bson.M{"hello": 1}).Decode(&result)
	if err != nil {
		return bson.M{"ok": 0, "errmsg": err.Error()}
	}

	return result
}
