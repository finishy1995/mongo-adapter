package network

import (
	"finishy1995/mongo-adapter/library/log"
	"finishy1995/mongo-adapter/protocol"
	"time"

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
	if err := gnet.Run(s, "tcp://"+address, gnet.WithLogLevel(logging.InfoLevel)); err != nil {
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
		log.Errorf("Client disconnected with error: %v", err)
	} else {
		log.Debugf("Client disconnected, addr: %s", c.RemoteAddr())
	}
	return
}

func (s *Server) OnTraffic(c gnet.Conn) (action gnet.Action) {
	buf, _ := c.Next(-1)
	if !s.protocolServer.OnMessage(c, buf) {
		return gnet.Close
	}

	return
}

func (s *Server) OnTick() (delay time.Duration, action gnet.Action) {
	return 0, gnet.None
}
