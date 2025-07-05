package protocol

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

var (
	reservedKeys = map[string]struct{}{
		"lsid": {},
	}
	exposeAddr = "127.0.0.1:27017"
)

func RegisterExposeAddress(addr string) {
	exposeAddr = addr
}

// 帮助函数，在 bson.D 中查找 key
func getValueFromD(doc bson.D, key string) (interface{}, bool) {
	for _, e := range doc {
		if e.Key == key {
			return e.Value, true
		}
	}
	return nil, false
}

func messageHandle(message bson.D) bson.M {
	if db == nil {
		return bson.M{"ok": 0, "errmsg": "MongoDB server connection error"}
	}
	var result bson.M

	// MongoDB 4.0 前，使用 MONGODB-CR 认证
	if _, ok := getValueFromD(message, "getnonce"); ok {
		result = bson.M{
			"nonce": getRandomString(16),
			"ok":    1,
		}
		return result
	}

	// 正确查找 ismaster 字段
	if _, ok := getValueFromD(message, "ismaster"); ok {
		err := db.Database("admin").RunCommand(context.TODO(), bson.M{"ismaster": 1}).Decode(&result)
		if err != nil {
			return bson.M{"ok": 0, "errmsg": err.Error()}
		}
		// 构造更简洁的响应
		response := bson.M{
			"ismaster":       true,
			"maxWireVersion": result["maxWireVersion"],
			"minWireVersion": result["minWireVersion"],
			"ok":             1,
			"hosts":          []string{exposeAddr},
			"primary":        exposeAddr,
			"me":             exposeAddr,
		}
		if v, ok := result["logicalSessionTimeoutMinutes"]; ok {
			response["logicalSessionTimeoutMinutes"] = v
		}
		return response
	}

	cmd := bson.D{}
	for _, e := range message {
		k := e.Key
		v := e.Value
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

	err := db.Database(getDBFromD(message)).RunCommand(context.TODO(), cmd).Decode(&result)
	if err != nil {
		return bson.M{"ok": 0, "errmsg": err.Error()}
	}
	return result
}

// 这里改成 bson.D 版
func getDBFromD(message bson.D) string {
	dbName := "admin"
	for _, e := range message {
		if e.Key == "$db" {
			if v, ok := e.Value.(string); ok && v != "" {
				dbName = v
			}
			break
		}
	}
	return dbName
}
