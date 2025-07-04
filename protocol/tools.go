package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
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
