package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/eleztian/toh"
)

func startServer() {
	s := toh.NewTcpServer(context.Background(), "tcp", nil)

	go func() {
		defer s.Close()
		for {
			conn, err := s.Accept()
			if err != nil {
				return
			}
			go func() {
				defer func() { _ = conn.Close() }()

				for {
					data := make([]byte, 1024)
					n, err := conn.Read(data)
					if err != nil {
						return
					}
					_, _ = conn.Write(data[:n])
				}
			}()
		}
	}()

	smux := http.NewServeMux()
	smux.Handle("/tcp", s)

	err := http.ListenAndServe("127.0.0.1:8080", smux)
	if err != nil {
		panic(err)
	}
}

func main() {
	go startServer()

	conn, err := toh.Dial(http.MethodGet, "http://192.168.111.62/kp/tcp")
	if err != nil {
		panic(err)
	}

	for {
		conn.Write([]byte("hello"))
		data := make([]byte, 1024)
		n, err := conn.Read(data)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(data[:n]))
		time.Sleep(time.Second)
	}

}
