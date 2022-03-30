package scrpc

import (
	"context"
	"github.com/victor-leee/scrpc/github.com/victor-leee/scrpc"
	"google.golang.org/protobuf/proto"
	"io"
)

type RequestContext struct {
	Ctx           context.Context
	Req           proto.Message
	ReqService    string
	ReqMethod     string
	SenderService string
	Resp          proto.Message
}

type Client interface {
	UnaryRPCRequest(reqCtx *RequestContext) error
}

type clientImpl struct {
}

func NewClient() Client {
	InitConnManager(func(cname string) (ConnPool, error) {
		localTransportCfg := GetConfig().LocalTransportConfig
		return NewPool(WithInitSize(localTransportCfg.PoolCfg.InitSize), WithMaxSize(localTransportCfg.PoolCfg.MaxSize), WithFactory(func() (*Conn, error) {
			return Dial(localTransportCfg.Protocol, cname, WithType(ConnTypeSideCar2Local))
		}))
	})
	return &clientImpl{}
}

func (c *clientImpl) UnaryRPCRequest(ctx *RequestContext) error {
	rpcReq := FromProtoMessage(ctx.Req, &scrpc.Header{
		ReceiverServiceName: ctx.ReqService,
		ReceiverMethodName:  ctx.ReqMethod,
		SenderServiceName:   ctx.SenderService,
		MessageType:         scrpc.Header_SIDE_CAR_PROXY,
		TraceId:             "todo",                  // TODO
		Extra:               make(map[string]string), // TODO
	})
	outErr := GlobalConnManager().Func(GetConfig().LocalTransportConfig.Path, func(conn *Conn) error {
		if _, writeErr := rpcReq.Write(conn); writeErr != nil {
			return writeErr
		}
		resp, respErr := FromReader(conn, blockRead)
		if respErr != nil {
			return respErr
		}
		if unmarshalErr := proto.Unmarshal(resp.Body, ctx.Resp); unmarshalErr != nil {
			return unmarshalErr
		}

		return nil
	})
	if outErr != nil {
		return outErr
	}

	return nil
}

func blockRead(reader io.Reader, size uint64) ([]byte, error) {
	b := make([]byte, size)
	already := 0
	inc := 0
	var err error
	for uint64(already) < size {
		if inc, err = reader.Read(b[already:]); err != nil {
			return nil, err
		}
		already += inc
	}

	return b, nil
}
