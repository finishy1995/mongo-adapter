package protocol

import (
	"context"
	"finishy1995/mongo-adapter/library/log"

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

func messageHandle(message bson.D, hookContext *HookContext) bson.M {
	if db == nil {
		return bson.M{"ok": 0, "errmsg": "MongoDB server connection error"}
	}
	var result bson.M
	hookContext.Type = HookReqHandleAfter
	defer func() {
		hookContext.Response = result
		hookContext.Type = HookRespHandleBefore
		fireHook(hookContext)
	}()

	// MongoDB 4.0 前，使用 MONGODB-CR 认证
	if _, ok := getValueFromD(message, "getnonce"); ok {
		hookContext.Request = bson.D{{Key: "getnonce", Value: 1}}
		fireHook(hookContext)
		log.Warnf("getnonce command is not supported after MongoDB 4.0, message: %+v", message)
		result = bson.M{
			"nonce": getRandomString(16),
			"ok":    1,
		}
		return result
	}

	cmd := bson.D{}
	// 正确查找 ismaster 字段
	if _, ok := getValueFromD(message, "ismaster"); ok {
		cmd = bson.D{{Key: "ismaster", Value: 1}}
		fireHook(hookContext)
		var response bson.M
		err := db.Database("admin").RunCommand(context.TODO(), cmd).Decode(&response)
		if err != nil {
			log.Warnf("ismaster command failed: %v, message: %+v", err, message)
			result = bson.M{"ok": 0, "errmsg": err.Error()}
			return result
		}
		// 构造更简洁的响应
		result = bson.M{
			"ismaster":       true,
			"maxWireVersion": response["maxWireVersion"],
			"minWireVersion": response["minWireVersion"],
			"ok":             1,
			"hosts":          []string{exposeAddr},
			"primary":        exposeAddr,
			"me":             exposeAddr,
		}
		if v, ok := response["logicalSessionTimeoutMinutes"]; ok {
			response["logicalSessionTimeoutMinutes"] = v
		}
		return result
	}

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

	fireHook(hookContext)
	err := db.Database(getDBFromD(message)).RunCommand(context.TODO(), cmd).Decode(&result)
	if err != nil {
		// TODO: 使用更好的错误处理和记录方式，例如筛选出权限不足的错误、参数错误等
		//
		// 1. (AuthenticationFailed) Authentication failed.
		//
		// 2. Skynet testmongodb.lua test_find_and_remove db.testcoll:ensureIndex({test_key = 1}, {test_key2 = -1}, {unique = true, name = "test_index"})
		//   RunCommand failed: (InvalidIndexSpecificationOption) Error in specification { test_key2: -1, name: "test_key_1", key: { test_key: 1 } } :: caused by :: The field 'test_key2' is not valid for an index specification. Specification: { test_key2: -1, name: "test_key_1", key: { test_key: 1 } }, message: [{Key:createIndexes Value:testcoll} {Key:$db Value:admin} {Key:indexes Value:[[{Key:test_key2 Value:-1} {Key:name Value:test_key_1} {Key:key Value:[{Key:test_key Value:1}]}]]}]
		//   Seems like a skynet bug, any MongoDB version will return the same error.
		log.Warnf("RunCommand failed: %v, message: %+v", err, message)
		return bson.M{"ok": 0, "errmsg": err.Error()}
	}
	if ok, _ := result["ok"].(float64); ok != 1 {
		log.Warnf("RunCommand MongoDB server error, errmsg: %v, message: %+v", result["errmsg"], message)
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
