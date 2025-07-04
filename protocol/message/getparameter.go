package message

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	RegisterMessage(&getParameterMessage{})
}

type getParameterMessage struct {
}

func (h *getParameterMessage) Accept(message bson.M) bool {
	if db := getDB(message); db == "admin" {
		if _, ok := message["getParameter"]; ok {
			return true
		}
	}
	return false
}

func (h *getParameterMessage) Handle(message bson.M) bson.M {
	if db == nil {
		return bson.M{"ok": 0, "errmsg": "MongoDB server connection error"}
	}

	// 只保留 getParameter 相关字段，构造命令
	cmd := bson.D{}
	for k, v := range message {
		if k == "getParameter" || k == "featureCompatibilityVersion" {
			cmd = append(cmd, bson.E{Key: k, Value: v})
		}
	}

	var result bson.M
	err := db.Database("admin").RunCommand(context.TODO(), cmd).Decode(&result)
	if err != nil {
		return bson.M{"ok": 0, "errmsg": err.Error()}
	}

	return result
}
