package network

import (
	"encoding/binary"
	"errors"
	"io"
	"time"

	"finishy1995/mongo-adapter/library/log"
	"finishy1995/mongo-adapter/protocol"

	"github.com/panjf2000/gnet/v2"
	"github.com/panjf2000/gnet/v2/pkg/logging"
)

type Server struct {
	eng            gnet.Engine
	address        string
	protocolServer *protocol.Server
}

func NewServerAndMustStart(address string, protocol *protocol.Server) {
	s := &Server{
		address:        address,
		protocolServer: protocol,
	}
	if err := gnet.Run(s, "tcp://"+address, gnet.WithLogLevel(logging.ErrorLevel)); err != nil {
		panic(err)
	}
}

func (s *Server) OnBoot(eng gnet.Engine) (action gnet.Action) {
	s.eng = eng
	log.Infof("MongoDB adapter is listening on %s with gnet", s.address)
	return
}

func (s *Server) OnShutdown(eng gnet.Engine) {
	log.Infof("MongoDB adapter is shutting down on %s with gnet", s.address)
}

func (s *Server) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	log.Debugf("New client connection established, addr: %s", c.RemoteAddr())
	return
}

func (s *Server) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	if err != nil {
		// 忽略EOF错误，视为正常断开连接
		if !errors.Is(err, io.EOF) {
			log.Errorf("Client disconnected with error: %v, addr: %s", err, c.RemoteAddr())
			return
		}
	}
	log.Debugf("Client disconnected, addr: %s", c.RemoteAddr())
	return
}

func (s *Server) OnTraffic(c gnet.Conn) (action gnet.Action) {
	for {
		// MongoDB 协议头固定4 + 4 + 4 + 4 = 16字节
		header, _ := c.Peek(16)
		if len(header) < 16 {
			break // 半包
		}
		// 取前4字节为包长（含头），小端序
		length := int(binary.LittleEndian.Uint32(header[:4]))
		if length < 16 {
			// 协议错误，直接断开
			return gnet.Close
		}
		// 判断整包是否到齐
		full, _ := c.Peek(length)
		if len(full) < length {
			break // 半包
		}
		// 交给 protocolServer 处理
		if !s.protocolServer.OnMessage(c, full[:length]) {
			return gnet.Close
		}
		// 消费掉已处理数据
		c.Discard(length)
	}
	return
}

func (s *Server) OnTick() (delay time.Duration, action gnet.Action) {
	return 0, gnet.None
}
