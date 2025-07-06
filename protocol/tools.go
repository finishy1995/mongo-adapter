package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"

	"go.mongodb.org/mongo-driver/bson"
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
