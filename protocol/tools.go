package protocol

import (
	"bytes"
	"encoding/binary"
	"finishy1995/mongo-adapter/library/log"
	"fmt"
	"io"
	"math/rand"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// readCString reads a C-style string from the provided buffer
func readCString(buf *bytes.Buffer) (string, error) {
	str, err := buf.ReadString(0)
	if err != nil {
		return "", err
	}
	// Remove the null terminator
	return str[:len(str)-1], nil
}

// 生成一个随机字符串
func getRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// 从 io.Reader 读取一个 BSON 文档完整 []byte
func readBSONBytes(reader io.Reader) ([]byte, error) {
	// 先读 4 字节长度
	var length int32
	err := binary.Read(reader, binary.LittleEndian, &length)
	if err != nil {
		return nil, fmt.Errorf("read bson length: %w", err)
	}
	if length < 5 || length > 16*1024*1024 { // 简单校验
		return nil, fmt.Errorf("unrealistic bson length: %d", length)
	}

	// 已经读了4字节，还需要再读length-4字节
	buf := make([]byte, length)
	binary.LittleEndian.PutUint32(buf[:4], uint32(length))

	// 读剩下的length-4字节
	if _, err := io.ReadFull(reader, buf[4:]); err != nil {
		return nil, fmt.Errorf("read bson body: %w", err)
	}
	return buf, nil
}

// 深拷贝 bson.D
func deepCopyBsonD(src bson.D) bson.D {
	if src == nil {
		return nil
	}
	dst := make(bson.D, len(src))
	for i, elem := range src {
		dst[i] = bson.E{
			Key:   elem.Key,
			Value: deepCopyValue(elem.Value),
		}
	}
	return dst
}

// 深拷贝 bson.M
func deepCopyBsonM(src bson.M) bson.M {
	if src == nil {
		return nil
	}
	dst := make(bson.M, len(src))
	for k, v := range src {
		dst[k] = deepCopyValue(v)
	}
	return dst
}

// 深度拷贝任意值
func deepCopyValue(v interface{}) interface{} {
	switch val := v.(type) {
	case bson.D:
		return deepCopyBsonD(val)
	case bson.M:
		return deepCopyBsonM(val)
	case []interface{}:
		copied := make([]interface{}, len(val))
		for i, item := range val {
			copied[i] = deepCopyValue(item)
		}
		return copied
	case map[string]interface{}:
		copied := make(map[string]interface{}, len(val))
		for k, item := range val {
			copied[k] = deepCopyValue(item)
		}
		return copied
	default:
		return val // 基本类型直接返回
	}
}

// GetOperationAndCollection
// 输入: request: [{Key:find Value:testcoll} {Key:$db Value:testdb} {Key:filter Value:[{Key:test_key Value:1}]} {Key:limit Value:1} {Key:projection Value:[]}]
// 输出: operation: find, collection: testdb.testcoll
func GetOperationAndCollection(request bson.D) (operation string, collection string) {
	var dbName, collName string

	// 定义可能的操作键
	operationKeys := map[string]bool{
		"find":          true,
		"insert":        true,
		"update":        true,
		"delete":        true,
		"aggregate":     true,
		"count":         true,
		"distinct":      true,
		"mapReduce":     true,
		"create":        true,
		"drop":          true,
		"createIndexes": true,
		"dropIndexes":   true,
	}

	for _, elem := range request {
		// 提取数据库名称
		if elem.Key == "$db" {
			if db, ok := elem.Value.(string); ok {
				dbName = db
			}
		} else if operationKeys[elem.Key] && operation == "" {
			// 提取操作名称和集合名称（只取第一个匹配的操作键）
			operation = elem.Key
			if coll, ok := elem.Value.(string); ok {
				collName = coll
			}
		}
	}

	// 组合数据库和集合名称
	if dbName != "" && collName != "" {
		collection = dbName + "." + collName
	}

	if operation == "" {
		log.Warnf("get operation and collection failed, request: %+v", request)
	}

	return operation, collection
}

func GetOperationTime(response bson.M) uint32 {
	if response == nil {
		return 0
	}
	if operationTime, ok := response["operationTime"]; ok {
		if t, ok := operationTime.(primitive.Timestamp); ok {
			return t.T
		}
	}
	return 0
}
