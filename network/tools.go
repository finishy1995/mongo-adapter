package network

import (
	"bytes"
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
