// Package id 提供了生成数字和字符串 ID 的方法，同时支持 JSON 对 ID 的加解密
//
// Example Usage
//
//	import "nbserver/common/id"
//
//	func main() {
//	  id := id.GenerateID()
//	}
//
// Benchmark History
//
//	goos: windows
//	goarch: amd64
//	pkg: nbserver/common/id
//	cpu: Intel(R) Core(TM) i7-10700K CPU @ 3.80GHz
//	BenchmarkIDGeneration
//	BenchmarkIDGeneration-16        89568280                13.29 ns/op
package id
