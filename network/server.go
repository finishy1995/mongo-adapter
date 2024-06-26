package network

import (
	"encoding/binary"
	"github.com/finishy1995/mongo-adapter/library/log"
	"github.com/finishy1995/mongo-adapter/processor"
	"github.com/finishy1995/mongo-adapter/types"
	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet/v2"
)

type Server struct {
	gnet.BuiltinEventEngine
	addr string
}

func NewServerAndMustStart(addr string) *Server {
	server := new(Server)
	err := gnet.Run(server, "tcp://"+addr, gnet.WithMulticore(true), gnet.WithLogger(log.GetLogger()))
	if err != nil {
		panic(err)
	}
	return server
}

func (s *Server) OnBoot(_ gnet.Engine) (action gnet.Action) {
	log.Infof("MongoDB adapter is listening on %s\n", s.addr)
	return gnet.None
}

func (s *Server) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	log.Debugf("New connection opened: %s", c.RemoteAddr().String())
	return
}

func (s *Server) OnClose(c gnet.Conn, _ error) (action gnet.Action) {
	log.Debugf("Connection closed: %s", c.RemoteAddr().String())
	return
}

func (s *Server) OnTraffic(c gnet.Conn) (action gnet.Action) {
	s.handleMessage(c)
	return
}

func (s *Server) handleMessage(conn gnet.Conn) {
	header := types.MsgHeader{}
	// 把 buf 处理为 header
	if err := binary.Read(conn, binary.LittleEndian, &header); err != nil {
		log.Errorf("Read header failed: %s", err.Error())
		return
	}

	buf := make([]byte, header.MessageLength-16)
	if _, err := conn.Read(buf); err != nil {
		log.Errorf("Error reading body: %s", err.Error())
		return
	}

	log.Debugf("Received header: %+v, message: %s", header, string(buf))

	err := ants.Submit(func() {
		processor.HandleMessage(&header, buf, conn)
	})

	if err != nil {
		log.Errorf("Failed to submit task: %s", err.Error())
	}
}
