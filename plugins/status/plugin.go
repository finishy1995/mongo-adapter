package status

import (
	"finishy1995/mongo-adapter/library/id"
	"finishy1995/mongo-adapter/library/log"
	"finishy1995/mongo-adapter/plugins"
	"finishy1995/mongo-adapter/protocol"
	"sync"
	"sync/atomic"
	"time"
)

func init() {
	plugins.RegisterPlugin(&StatusPlugin{})
}

type StatusPlugin struct {
	count     *int64 // 总请求数
	success   *int64 // 成功数
	totalTime *int64 // 累计耗时，单位纳秒
	minTime   *int64 // 最小耗时
	maxTime   *int64 // 最大耗时

	closer       chan bool
	storeSecond  uint32
	enableLogger bool

	contextMapMutex sync.Mutex
	contextMap      map[id.ID]int64
}

func (s *StatusPlugin) Name() string {
	return "status"
}

func (s *StatusPlugin) Requires() []string {
	return nil
}

func (s *StatusPlugin) Init() error {
	protocol.RegisterHook(protocol.HookStart, s.HookStart)
	protocol.RegisterHook(protocol.HookRespHandleAfter, s.HookRespHandleAfter)
	s.storeSecond = 600   // 数据存储 600 秒，TODO: 可配置
	s.enableLogger = true // TODO: 可配置
	s.closer = make(chan bool, 2)
	s.RefreshStatus()
	s.contextMap = make(map[id.ID]int64)

	go s.tpsAggregator()
	return nil
}

func (s *StatusPlugin) UnInit() error {
	protocol.UnRegisterHook(protocol.HookStart, s.HookStart)
	protocol.UnRegisterHook(protocol.HookRespHandleAfter, s.HookRespHandleAfter)
	s.closer <- true
	return nil
}

func (s *StatusPlugin) HookStart(ctx *protocol.HookContext) {
	atomic.AddInt64(s.count, 1)
	s.contextMapMutex.Lock()
	defer s.contextMapMutex.Unlock()
	s.contextMap[ctx.ID] = time.Now().UnixMicro()
}

func (s *StatusPlugin) HookRespHandleAfter(ctx *protocol.HookContext) {
	if ctx.ErrMsg != "" {
		return
	}
	atomic.AddInt64(s.success, 1)
	s.contextMapMutex.Lock()
	startTime, ok := s.contextMap[ctx.ID]
	if !ok {
		s.contextMapMutex.Unlock()
		return
	}
	delete(s.contextMap, ctx.ID)
	s.contextMapMutex.Unlock()

	costTime := time.Now().UnixMicro() - startTime
	atomic.AddInt64(s.totalTime, costTime)
	if s.minTime == nil || *s.minTime == 0 || costTime < *s.minTime {
		atomic.StoreInt64(s.minTime, costTime)
	}
	if s.maxTime == nil || costTime > *s.maxTime {
		atomic.StoreInt64(s.maxTime, costTime)
	}
}

func (s *StatusPlugin) RefreshStatus() {
	s.count = new(int64)
	s.success = new(int64)
	s.totalTime = new(int64)
	s.minTime = new(int64)
	s.maxTime = new(int64)
}

func (s *StatusPlugin) tpsAggregator() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var count uint32 = 0
	for {
		select {
		case <-ticker.C:
			count++
			if s.enableLogger {
				count := atomic.LoadInt64(s.count)
				success := atomic.LoadInt64(s.success)
				totalTime := atomic.LoadInt64(s.totalTime)
				minTime := atomic.LoadInt64(s.minTime)
				maxTime := atomic.LoadInt64(s.maxTime)

				var averageTime int64 = 0
				if count > 0 {
					averageTime = totalTime / count / 1000
				}
				log.Infof("[STATUS] count: %d, success: %d, successRatio: %.1f%%, averageTime: %d ms, minTime: %d ms, maxTime: %d ms", count, success, float64(success)/float64(count)*100, averageTime, minTime/1000, maxTime/1000)
			}
			if count > s.storeSecond {
				s.RefreshStatus()
				count = 0
			}
		case <-s.closer:
			return
		}
	}
}
