package toh

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"
)

type DialerOption func(req *http.Request, tlsConfig *tls.Config)

func OptionTlsCerts(certPath string, keyPath string) DialerOption {
	cliCert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		panic(err)
	}

	return OptionTlsCertsWithConfig([]tls.Certificate{cliCert})
}

func OptionTlsCertsWithConfig(certs []tls.Certificate) DialerOption {
	return func(req *http.Request, tlsConfig *tls.Config) {
		if tlsConfig == nil {
			tlsConfig = &tls.Config{
				Certificates:       certs,
				InsecureSkipVerify: true,
			}
		} else {
			tlsConfig.InsecureSkipVerify = true
			if tlsConfig.Certificates != nil {
				tlsConfig.Certificates = append(tlsConfig.Certificates, certs...)
			} else {
				tlsConfig.Certificates = certs
			}
		}
	}
}

func OptionConnID(id string) DialerOption {
	return func(req *http.Request, tlsConfig *tls.Config) {
		req.Header.Set(HeaderID, id)
	}
}

type Dialer struct {
	req       *http.Request
	tlsConfig *tls.Config
}

func NewDialer(method string, url string, ops ...DialerOption) (*Dialer, error) {
	if method == "" {
		method = http.MethodGet
	}

	var tlsConfig = new(tls.Config)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	for _, op := range ops {
		op(req, tlsConfig)
	}

	return &Dialer{
		req:       req,
		tlsConfig: tlsConfig,
	}, nil
}

func (d *Dialer) DialContext(ctx context.Context, ops ...DialerOption) (net.Conn, error) {
	req := d.req.Clone(ctx)
	for _, op := range ops {
		op(req, d.tlsConfig)
	}

	return dial(ctx, req, d.tlsConfig)
}

func Dial(method string, addr string, ops ...DialerOption) (net.Conn, error) {
	if method == "" {
		method = http.MethodGet
	}

	var tlsConfig = new(tls.Config)
	req, err := http.NewRequest(method, addr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Connection", "upgrade")
	req.Header.Set("Upgrade", FullProtocol)

	for _, op := range ops {
		op(req, tlsConfig)
	}

	return dial(context.Background(), req, tlsConfig)
}

func dial(ctx context.Context, req *http.Request, tlsConfig *tls.Config) (conn net.Conn, err error) {
	var addr = req.Host

	ctxTo, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	if req.URL.Scheme == "https" {
		dialer := tls.Dialer{}
		if len(tlsConfig.Certificates) != 0 {
			dialer.Config = tlsConfig
		}
		if !strings.Contains(addr, ":") {
			addr = addr + ":443"
		}
		conn, err = dialer.DialContext(ctxTo, "tcp", addr)
	} else {
		var d net.Dialer
		if !strings.Contains(addr, ":") {
			addr = addr + ":80"
		}
		conn, err = d.DialContext(ctxTo, "tcp", addr)
	}
	if err != nil {
		return nil, err
	}

	err = req.WithContext(ctx).Write(conn)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	rsp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	if rsp.StatusCode != http.StatusSwitchingProtocols {
		_ = conn.Close()
		return nil, errors.New(rsp.Status)
	}

	return &Conn{
		Conn:   conn,
		URL:    *req.URL,
		Header: rsp.Header,
	}, nil
}
