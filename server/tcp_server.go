// @Author: wanglanwei@baidu.com
// @Description: Socks5 Tcp Server
// @Reference: https://tools.ietf.org/html/rfc1928
package server

import (
	"fmt"
	"net"
	"sync"
)

// Tcp Server配置
type TcpConfig struct {
	MaxConnections     int    // 客户端最大连接数
	Host               string // 监听主机
	Port               int    // 监听端口
	LocalReadTimeout   int    // 本地读超时，单位：毫秒
	LocalWriteTimeout  int    // 本地写超时，单位：毫秒
	RemoteConnTimeout  int    // 远程连接超时，单位：毫秒
	RemoteReadTimeout  int    // 远程读超时，单位：毫秒
	RemoteWriteTimeout int    // 远程写超时，单位：毫秒
}

type TcpServer struct {
	host       string           // 主机
	port       int              // 端口
	config     *TcpConfig       // 服务器配置
	listener   *net.TCPListener // Listener
	statusLock sync.Mutex       // 状态变更的锁
	quit       bool             // 是否退出
	running    bool             // 运行中
	stopped    bool             // 已停止
}

func (tcpSrv *TcpServer) Proto() string {
	return ProtoTCP
}

func (tcpSrv *TcpServer) Address() string {
	return fmt.Sprintf("%s:%d", tcpSrv.host, tcpSrv.port)
}

func (tcpSrv *TcpServer) Config() *TcpConfig {
	return tcpSrv.config
}

func (tcpSrv *TcpServer) Run() error {
	if tcpSrv.running {
		return OperationAlreadyDone
	}
	ln, err := net.Listen("tcp", tcpSrv.Address())
	if err != nil {
		return err
	}
	tcpSrv.listener = ln.(*net.TCPListener)
	tcpSrv.running = true

	for !tcpSrv.quit {
		conn, err := tcpSrv.listener.Accept()
		if err != nil {
			continue
		}
		tcpSrv.handleConn(conn)
	}

	return nil
}

func (tcpSrv *TcpServer) handleConn(conn net.Conn) {
	fmt.Printf("accept an new connection <%s>\n", conn.RemoteAddr())
	go func() {
		tcpClient := newTcpClient(conn, tcpSrv)
		defer tcpClient.Close()
		if err := tcpClient.Serv(); err != nil {
			fmt.Printf("connection aborted, last err: %s\n", err)
		}
	}()
}

func (tcpSrv *TcpServer) Stop() error {
	tcpSrv.quit = true
	// 等待所有线程结束
	return nil
}

func newTcpServer(tcpConfig *TcpConfig) (*TcpServer, error) {
	tcpServer := &TcpServer{
		host:   tcpConfig.Host,
		port:   tcpConfig.Port,
		config: tcpConfig,
	}
	return tcpServer, nil
}
