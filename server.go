package scrpc

import (
	"github.com/victor-leee/scrpc/github.com/victor-leee/scrpc"
	"google.golang.org/protobuf/proto"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type PluginHandler func(b []byte) (proto.Message, error)

type Server interface {
	RegisterHandler(name string, h PluginHandler)
	Start() error
	WaitTermination()
}

type serverImpl struct {
	handlers map[string]PluginHandler
}

func (s *serverImpl) RegisterHandler(name string, h PluginHandler) {
	if s.handlers == nil {
		s.handlers = make(map[string]PluginHandler)
	}
	s.handlers[name] = h
}

func (s *serverImpl) Start() error {
	concurrency := 50
	for i := 0; i < concurrency; i++ {
		go func() {
			if err := s.waitMsg(); err != nil {
				// TODO LOG HERE
			}
		}()
	}

	return nil
}

func (s *serverImpl) waitMsg() error {
	outErr := GlobalConnManager().Func("/tmp/sc.sock", func(conn net.Conn) error {
		// the connection is used to receive requests
		_, buildErr := FromBody([]byte{}, &scrpc.Header{
			MessageType: scrpc.Header_SET_USAGE,
		}).Write(conn)
		if buildErr != nil {
			return buildErr
		}

		for {
			msg, readErr := FromReader(conn, blockRead)
			if readErr != nil {
				return readErr
			}
			h := s.handlers[msg.Header.ReceiverMethodName]
			resp, err := h(msg.Body)
			if err != nil {
				// TODO LOG HERE
				continue
			}
			_, writeErr := FromProtoMessage(resp, nil).Write(conn)
			if writeErr != nil {
				// TODO LOG HERE
				continue
			}
		}

	})
	if outErr != nil {
		return outErr
	}

	return nil
}

func (s *serverImpl) WaitTermination() {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
