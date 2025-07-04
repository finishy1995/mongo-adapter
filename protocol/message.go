package protocol

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

var reservedKeys = map[string]struct{}{
	"lsid": {},
}

func messageHandle(message bson.M) bson.M {
	if db == nil {
		return bson.M{"ok": 0, "errmsg": "MongoDB server connection error"}
	}
	var result bson.M

	if _, ok := message["ismaster"]; ok {
		err := db.Database("admin").RunCommand(context.TODO(), bson.M{"hello": 1}).Decode(&result)
		if err != nil {
			return bson.M{"ok": 0, "errmsg": err.Error()}
		}
		result["hosts"] = []string{"127.0.0.1:27017"}
		result["primary"] = "127.0.0.1:27017"
		result["me"] = "127.0.0.1:27017"
		return result
	}

	cmd := bson.D{}
	for k, v := range message {
		// 跳过 $ 开头的保留字段
		if len(k) > 0 && k[0] == '$' {
			continue
		}
		// 跳过在保留字段集合中的字段
		if _, found := reservedKeys[k]; found {
			continue
		}
		cmd = append(cmd, bson.E{Key: k, Value: v})
	}

	err := db.Database(getDB(message)).RunCommand(context.TODO(), cmd).Decode(&result)
	if err != nil {
		return bson.M{"ok": 0, "errmsg": err.Error()}
	}
	return result
}

func getDB(message bson.M) string {
	dbName := "admin"
	if v, ok := message["$db"].(string); ok && v != "" {
		dbName = v
	}
	return dbName
}
