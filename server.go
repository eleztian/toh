package toh

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/google/uuid"
)

const (
	ProtocolVersion = "1.0"
	Protocol        = "toh"
	FullProtocol    = Protocol + "/" + ProtocolVersion
	HeaderID        = "X-Toh-ID"
	HeaderToh       = "Sec-Toh"
)

type TcpServer struct {
	*Listener
}

func NewTcpServer(ctx context.Context, name string, addr net.Addr) *TcpServer {
	res := &TcpServer{
		Listener: newListener(ctx, name, addr),
	}

	return res
}

func (t *TcpServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	conn, err := getConnFromHttp(request, writer)
	if err != nil {
		log.Printf("failed to get conn from http, path=%s, err=%v", request.URL.Path, err)
		return
	}

	err = t.Listener.AddConn(conn)
	if err != nil {
		log.Printf("failed to add conn to server, path=%s, err=%v", request.URL.Path, err)
		return
	}
	return
}

func getConnFromHttp(request *http.Request, writer http.ResponseWriter) (net.Conn, error) {
	if request.Header.Get(HeaderID) == "" {
		request.Header.Set(HeaderID, uuid.NewString())
	}

	writer.Header().Set(HeaderID, request.Header.Get(HeaderID))
	writer.Header().Set(HeaderToh, ProtocolVersion)
	writer.WriteHeader(http.StatusSwitchingProtocols)

	conn, buf, err := writer.(http.Hijacker).Hijack()
	if err != nil {
		return nil, err
	}
	if tcp, ok := conn.(*net.TCPConn); ok {
		_ = tcp.SetKeepAlive(true)
	}
	buffered := buf.Reader.Buffered()
	if buffered != 0 {
		data := make([]byte, buffered)
		_, err = io.ReadFull(buf.Reader, data)
		if err != nil {
			_ = conn.Close()
			return nil, err
		}
		conn = newBuffedConn(conn, data)
	}

	return &Conn{
		Conn:   conn,
		Header: request.Header,
		URL:    *request.URL,
	}, nil
}
