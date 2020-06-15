package client

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"
)

const (
	StatusDisconnected int = 0x00 // 未连接
	StatusConnected        = 0x01 // 已连接
)

var (
	sessionIdSeed int64 = 0xf0000000 // 会话种子
)

// Turbo客户端
type TurboClient struct {
	conn      *net.TCPConn  // Tcp连接
	sessionId int64         // 会话ID
	address   string        // 远程主机
	status    int           // 状态
	config    *ClientConfig // 配置
}

// Turbo客户端配置
type ClientConfig struct {
	Host        string
	Port        int
	ConnTimeout int
}

func newSessionId() int64 {
	return atomic.AddInt64(&sessionIdSeed, 1)
}

func NewTurboClient(config *ClientConfig) *TurboClient {
	client := &TurboClient{
		config:    config,
		address:   fmt.Sprintf("%s:%d", config.Host, config.Port),
		sessionId: newSessionId(),
	}
	return client
}

func (cli *TurboClient) Connect() error {
	timeout := time.Duration(cli.config.ConnTimeout) * time.Millisecond
	conn, err := net.DialTimeout("tcp", cli.address, timeout)
	if err != nil {
		return err
	}
	cli.conn = conn.(*net.TCPConn)
	// 与Server端协商
	// cli.status = StatusConnected
	return nil
}
