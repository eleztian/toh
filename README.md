# TOH 项目

TOH（Tcp Over HTTP）是一个通过HTTP进行TCP通信的Go语言库。
它允许客户端通过HTTP请求与服务器建立连接，并在两者之间传输数据。

## 安装

确保您已经安装了 Go 语言环境。然后，在您的终端中运行以下命令来获取此库：
```shell
go get github.com/eleztian/toh
```

## 使用方法

### 创建 TCP 服务器

创建一个 TCP 服务器，该服务器可以处理来自客户端的连接并回显收到的数据。

```go

func main() {
	s := toh.NewTcpServer(context.Background(), "tcp", nil)
	defer s.Close()

	smux := http.NewServeMux()
	smux.Handle("/tcp", s)

	go func() {
		err := http.ListenAndServe("0.0.0.0:8083", smux)
		if err != nil {
			panic(err)
		}
	}()

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
}
```


### 连接到 TCP 服务器

使用 `toh.Dial` 方法连接到远程服务器，并发送和接收数据。

```go
func func main() {
  conn, err := toh.Dial(http.MethodGet, "http://127.0.0.1/kp/tcp")
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
```


## 自定义选项

您可以使用自定义选项来自定义连接行为，例如设置TLS证书或连接ID。

## API 文档

### Dial

用于创建一个新的连接。

- **参数**
  - `method`: HTTP 请求方法，默认为 `GET`
  - `addr`: 目标地址
  - `ops...`: 可选的拨号选项

- **返回值**
  - `net.Conn`: 新的连接对象
  - `error`: 如果发生错误，则返回非空错误

### NewDialer

创建一个新的拨号器实例。

- **参数**
  - `method`: HTTP 请求方法，默认为 `GET`
  - `url`: 目标URL
  - `ops...`: 可选的拨号选项

- **返回值**
  - `*Dialer`: 拨号器实例
  - `error`: 如果发生错误，则返回非空错误

### OptionTlsCerts

加载TLS证书。

- **参数**
  - `certPath`: 证书路径
  - `keyPath`: 私钥路径

- **返回值**
  - `DialerOption`: TLS配置选项

### OptionConnID

设置连接ID。

- **参数**
  - `id`: 连接ID

- **返回值**
  - `DialerOption`: 连接ID选项

---

以上是 TOH 项目的 README 文件内容。希望这些信息能帮助您更好地理解和使用此库。
