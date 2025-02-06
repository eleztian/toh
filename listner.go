package toh

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
)

var ErrClosed = errors.New("use of closed network connection")

type Listener struct {
	name   string
	ctx    context.Context
	cancel func()
	conn   chan net.Conn
	addr   net.Addr
	closed *atomic.Bool
}

func newListener(ctx context.Context, name string, addr net.Addr) *Listener {
	ctxSub, ctxCancel := context.WithCancel(ctx)

	return &Listener{
		name: name,
		addr: addr,

		ctx:    ctxSub,
		cancel: ctxCancel,

		conn:   make(chan net.Conn),
		closed: &atomic.Bool{},
	}
}

func (h *Listener) Name() string {
	return h.name
}

func (h *Listener) AddConn(conn net.Conn) error {
	if h.closed.Load() {
		_ = conn.Close()
		return ErrClosed
	}

	select {
	case h.conn <- conn:
		return nil
	case <-h.ctx.Done():
		_ = conn.Close()
		return ErrClosed
	}
}

func (h *Listener) Accept() (net.Conn, error) {
	select {
	case <-h.ctx.Done():
		return nil, ErrClosed
	case conn := <-h.conn:
		if conn == nil {
			return nil, ErrClosed
		}
		return conn, nil
	}
}

func (h *Listener) Close() error {
	if h.closed.Load() {
		return nil
	}
	h.closed.Store(true)

	h.cancel()
	close(h.conn)

	return nil
}

func (h *Listener) Addr() net.Addr {
	return h.addr
}
