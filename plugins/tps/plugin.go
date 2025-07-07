package tps

import (
	"finishy1995/mongo-adapter/library/log"
	"finishy1995/mongo-adapter/plugins"
	"finishy1995/mongo-adapter/protocol"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func init() {
	plugins.RegisterPlugin(&TpsPlugin{})
}

type TpsPlugin struct {
	counters     sync.Map // key: collection+"."+operation+"."+"timestamp" value: *int32
	closer       chan bool
	storeSecond  uint32
	enableLogger bool
}

func (t *TpsPlugin) Name() string {
	return "tps"
}

func (t *TpsPlugin) Requires() []string {
	return nil
}

func (t *TpsPlugin) Init() error {
	protocol.RegisterHook(protocol.HookRespHandleBefore, t.RespHandleBefore)
	t.storeSecond = 600   // 数据存储 600 秒，TODO: 可配置
	t.enableLogger = true // TODO: 可配置
	t.closer = make(chan bool, 2)
	go t.tpsAggregator()
	return nil
}

func (t *TpsPlugin) UnInit() error {
	protocol.UnRegisterHook(protocol.HookRespHandleBefore, t.RespHandleBefore)
	t.closer <- true
	return nil
}

func (t *TpsPlugin) RespHandleBefore(context *protocol.HookContext) {
	if context.ErrMsg != "" {
		return
	}
	operation, collection := protocol.GetOperationAndCollection(context.Request)
	if operation == "" || collection == "" {
		return
	}
	// 只统计增删改查
	if operation != "find" && operation != "insert" && operation != "update" && operation != "delete" {
		return
	}
	operationTime := protocol.GetOperationTime(context.Response)
	if operationTime == 0 {
		return
	}

	key := getKey(collection, operation, operationTime)
	v, _ := t.counters.LoadOrStore(key, new(int32))
	atomic.AddInt32(v.(*int32), 1)

	total := getKey("*.*", operation, operationTime)
	v, _ = t.counters.LoadOrStore(total, new(int32))
	atomic.AddInt32(v.(*int32), 1)
}

func (t *TpsPlugin) tpsAggregator() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			lastInterval := uint32(time.Now().Unix() - 1)
			if t.enableLogger {
				find := t.getValue(getKey("*.*", "find", lastInterval))
				insert := t.getValue(getKey("*.*", "insert", lastInterval))
				update := t.getValue(getKey("*.*", "update", lastInterval))
				delete := t.getValue(getKey("*.*", "delete", lastInterval))
				log.Infof("[TPS] timestamp: %d, FIND: %d, INSERT: %d, UPDATE: %d, DELETE: %d", lastInterval, find, insert, update, delete)
			}

			needDelete := lastInterval - t.storeSecond
			t.counters.Range(func(key, value interface{}) bool {
				// 解析 key 的 timestamp
				_, _, timestamp := parseKey(key.(string))
				if timestamp <= needDelete {
					t.counters.Delete(key)
				}
				return true
			})
		case <-t.closer:
			return
		}
	}
}

func getKey(collection string, operation string, operationTime uint32) string {
	return fmt.Sprintf("%s.%s.%d", collection, operation, operationTime)
}

func parseKey(key string) (collection string, operation string, timestamp uint32) {
	parts := strings.Split(key, ".")
	if len(parts) != 4 {
		return "", "", 0
	}
	database := parts[0]
	collectionName := parts[1]
	operation = parts[2]
	ts, err := strconv.ParseUint(parts[3], 10, 32)
	if err != nil {
		timestamp = 0
	} else {
		timestamp = uint32(ts)
	}
	collection = database + "." + collectionName
	return
}

func (t *TpsPlugin) getValue(key string) int32 {
	v, ok := t.counters.Load(key)
	if !ok {
		return 0
	}
	return *v.(*int32)
}
