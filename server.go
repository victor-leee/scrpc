package scrpc

import (
	"github.com/sirupsen/logrus"
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

func NewServer() Server {
	InitConnManager(func(cname string) (ConnPool, error) {
		return NewPool(WithInitSize(1), WithMaxSize(50), WithFactory(func() (net.Conn, error) {
			return net.Dial("unix", cname)
		}))
	})
	return &serverImpl{
		handlers: make(map[string]PluginHandler),
	}
}

func (s *serverImpl) RegisterHandler(name string, h PluginHandler) {
	s.handlers[name] = h
}

func (s *serverImpl) Start() error {
	concurrency := 1
	for i := 0; i < concurrency; i++ {
		if err := s.waitMsg(); err != nil {
			logrus.Errorf("[Start] start message connection failed: %v", err)
		}
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
