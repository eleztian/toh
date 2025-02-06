package toh

import (
	"net"
	"net/http"
	"net/url"
)

type Stream interface {
	net.Conn
	ID() string
}

type buffedConn struct {
	buffered []byte
	net.Conn
}

func newBuffedConn(conn net.Conn, buffed []byte) net.Conn {
	return &buffedConn{
		buffered: buffed,
		Conn:     conn,
	}
}

func (c *buffedConn) Read(b []byte) (n int, err error) {
	if len(c.buffered) != 0 {
		if len(b) < len(c.buffered) {
			n = copy(b, c.buffered[:len(b)])
		} else {
			n = copy(b, c.buffered)
		}
		if n != 0 {
			c.buffered = c.buffered[n:]
			return
		}
	}

	return c.Conn.Read(b)
}

type Conn struct {
	net.Conn
	URL    url.URL
	Header http.Header
}

func (c *Conn) ID() string {
	return c.Header.Get("X-Connection-ID")
}

func (c *Conn) GetHeader(key string) string {
	if c.Header != nil {
		return c.Header.Get(key)
	}
	return ""
}

func (c *Conn) GetPath() string {
	return c.URL.Path
}
