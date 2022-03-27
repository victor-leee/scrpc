package scrpc

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

type ConnPoolFactory func(cname string) (ConnPool, error)

//go:generate mockgen -destination ../mock/conn_pool/mock.go -source ./conn_pool.go
type ConnPool interface {
	Get() (net.Conn, error)
	Put(conn net.Conn) error
	Close() error
}

type pool struct {
	opts       *options
	connChan   chan net.Conn
	ticketChan chan struct{}
}

type options struct {
	factory  func() (net.Conn, error)
	initConn int
	maxConn  int
}

type PoolOpt func(pool *pool)

func NewPool(opts ...PoolOpt) (ConnPool, error) {
	p := &pool{
		opts: &options{},
	}
	for _, opt := range opts {
		opt(p)
	}
	if err := p.init(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *pool) Get() (net.Conn, error) {
	select {
	case conn := <-p.connChan:
		return conn, nil
	default:
		// try to create one, if failed block wait the channel
		if !p.requestTicket() {
			return <-p.connChan, nil
		}

		return p.opts.factory()
	}
}

func (p *pool) Put(conn net.Conn) error {
	// before we put back the connection to the pool, we should check its status
	if p.isBrokenConn(conn) {
		// for each broken connection we allow one more creation
		p.createTicket()
		return errors.New("connection is broken")
	}
	p.connChan <- conn

	return nil
}

func (p *pool) isBrokenConn(conn net.Conn) (broken bool) {
	defer func() {
		// set read deadline "never"
		if err := conn.SetReadDeadline(time.Time{}); err != nil {
			broken = true
		}
	}()

	if err := conn.SetReadDeadline(time.Now()); err != nil {
		logrus.Errorf("set deadline failed, err:%v", err)
		return true
	}
	b := make([]byte, 1)
	if _, err := conn.Read(b); err != nil {
		if p.isReadTimeoutErr(err) {
			return false
		}
		logrus.Errorf("found a broken connection: %v", err)
	}

	return true
}

func (p *pool) isReadTimeoutErr(err error) bool {
	if netErr, ok := err.(*net.OpError); ok {
		return netErr.Timeout()
	}

	return false
}

func (p *pool) Close() error {
	// TODO impl
	panic("todo")
}

func (p *pool) init() error {
	if p.opts.initConn > p.opts.maxConn {
		return fmt.Errorf("initConn shouldn't exceed maxConn, actual is %d > %d", p.opts.initConn, p.opts.maxConn)
	}
	if p.opts.factory == nil {
		return errors.New("must provide factory method to generate new connection")
	}
	if p.opts.maxConn <= 0 {
		return errors.New("a pool should be allowed to contain resources, otherwise it's unnecessary to create one")
	}

	p.ticketChan = make(chan struct{}, p.opts.maxConn)
	for i := 0; i < p.opts.maxConn; i++ {
		p.createTicket()
	}

	p.connChan = make(chan net.Conn, p.opts.maxConn)
	for i := 0; i < p.opts.initConn; i++ {
		if !p.requestTicket() {
			return errors.New("request ticket to create connection failed")
		}
		conn, err := p.opts.factory()
		if err != nil {
			// failed ? no worry, retry later
			p.createTicket()
			continue
		}
		p.connChan <- conn
	}

	return nil
}

func (p *pool) requestTicket() (success bool) {
	select {
	case <-p.ticketChan:
		return true
	default:
		return false
	}
}

func (p *pool) createTicket() {
	select {
	case p.ticketChan <- struct{}{}:
	default:
	}
}

func WithFactory(f func() (net.Conn, error)) PoolOpt {
	return func(pool *pool) {
		pool.opts.factory = f
	}
}

func WithInitSize(s int) PoolOpt {
	return func(pool *pool) {
		pool.opts.initConn = s
	}
}

func WithMaxSize(s int) PoolOpt {
	return func(pool *pool) {
		pool.opts.maxConn = s
	}
}
