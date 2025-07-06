package protocol

import (
	"finishy1995/mongo-adapter/library/log"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
	"go.mongodb.org/mongo-driver/bson"
)

type HookType int8
type HookContext struct {
	ID       string
	Type     HookType
	ErrMsg   string
	Request  bson.D
	Response bson.M
}
type HookFunc func(*HookContext)

const (
	HookStart            HookType = 0
	HookReqHandleBefore  HookType = 1 // 原始请求经过加工前
	HookReqHandleAfter   HookType = 2 // 请求经过加工后，但未发送给 MongoDB Server 前
	HookRespHandleBefore HookType = 3 // MongoDB Server 响应经过加工前
	HookRespHandleAfter  HookType = 4 // 响应经过加工，并已发送给客户端后
	HookEnd              HookType = 127

	MaxHookType            = 6
	MaxHookChannelLifetime = 1 * time.Hour // TODO：可配置
)

var (
	hookManager = map[HookType][]HookFunc{}
	routinePool *ants.Pool

	hookChanMu     sync.RWMutex
	hookChannelMap = map[string]chan *HookContext{}
)

func init() {
	var err error
	routinePool, err = ants.NewPool(1000) // TODO: 1000 并发数应该可配置
	if err != nil {
		panic(err)
	}
}

// RegisterHook 非线程安全
func RegisterHook(hookType HookType, hookFunc HookFunc) {
	if hookManager[hookType] == nil {
		hookManager[hookType] = []HookFunc{}
	}
	hookManager[hookType] = append(hookManager[hookType], hookFunc)
}

func fireHook(context *HookContext) {
	log.Debugf("fireHook. context: %+v", context)
	if hookManager[context.Type] == nil && context.Type != HookEnd {
		return
	}
	copyContext := deepCopyContext(context)

	hookChanMu.Lock()
	channel, getChannelOk := hookChannelMap[copyContext.ID]
	if !getChannelOk {
		channel = make(chan *HookContext, MaxHookType)
		hookChannelMap[copyContext.ID] = channel

		err := routinePool.Submit(func() {
			ctx := &HookContext{
				ID:   copyContext.ID,
				Type: HookStart,
			}

			defer func() {
				if r := recover(); r != nil {
					log.Errorf("hook routine panic: %v", r)
				}

				if c, ok := hookChannelMap[ctx.ID]; ok {
					close(c)
					hookChanMu.Lock()
					delete(hookChannelMap, ctx.ID)
					hookChanMu.Unlock()
				}
			}()

			hookChanMu.RLock()
			c, ok := hookChannelMap[ctx.ID]
			hookChanMu.RUnlock()

			if !ok {
				log.Errorf("hook channel not found, id: %v", ctx.ID)
				return
			}
			for {
				select {
				case hookContext := <-c:
					if hookContext.Type == HookEnd {
						return
					}
					if hookContext.Type <= ctx.Type {
						log.Errorf("hook type error, now: %v, hook: %v", ctx.Type, hookContext.Type)
						return
					}

					ctx.Type = hookContext.Type
					if hookContext.ErrMsg != "" {
						ctx.ErrMsg = hookContext.ErrMsg
					}
					if hookContext.Request != nil {
						ctx.Request = hookContext.Request
					}
					if hookContext.Response != nil {
						ctx.Response = hookContext.Response
					}

					for _, hookFunc := range hookManager[ctx.Type] {
						safeRunHookFunc(hookFunc, ctx)
					}
				case <-time.After(MaxHookChannelLifetime):
					log.Errorf("hook routine timeout exit: %v", ctx.ID)
					return
				}
			}
		})
		if err != nil {
			log.Errorf("hook routine submit error: %v", err)
			close(channel)
			delete(hookChannelMap, copyContext.ID)
			hookChanMu.Unlock()
			return
		}
	}
	hookChanMu.Unlock()
	channel <- copyContext
}

func safeRunHookFunc(hookFunc HookFunc, context *HookContext) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("hookFunc panic: %v", r)
		}
	}()
	hookFunc(context)
}

func deepCopyContext(context *HookContext) *HookContext {
	return &HookContext{
		ID:       context.ID,
		Type:     context.Type,
		ErrMsg:   context.ErrMsg,
		Request:  deepCopyBsonD(context.Request),
		Response: deepCopyBsonM(context.Response),
	}
}
