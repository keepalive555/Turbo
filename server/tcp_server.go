// @Author: wanglanwei@baidu.com
// @Description: Socks5 Tcp Server
// @Reference: https://tools.ietf.org/html/rfc1928
package server

import (
	"fmt"
	"net"
	"sync"
	"time"
)

const (
	SrvStatusStopped  int = 0x00 // 服务器已停止
	SrvStatusRunning      = 0x01 // 服务器正在运行
	SrvStatusStopping     = 0x03 // 服务器正在停止
)

type TcpServer struct {
	host       string           // 主机
	port       int              // 端口
	config     *TcpConfig       // 服务器配置
	listener   *net.TCPListener // Listener
	statusLock sync.Mutex       // 状态变更的锁
	done       chan struct{}    // Server退出
	status     int              // 服务器状态
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
	tcpSrv.statusLock.Lock()
	if tcpSrv.status == SrvStatusRunning {
		tcpSrv.statusLock.Unlock()
		return ServerAlreadyStarted
	}
	ln, err := net.Listen(ProtoTCP, tcpSrv.Address())
	if err != nil {
		tcpSrv.statusLock.Unlock()
		return err
	}
	fmt.Printf("server <%s> startted!!!\n", tcpSrv.Address())
	tcpSrv.listener = ln.(*net.TCPListener)
	tcpSrv.status = SrvStatusRunning
	tcpSrv.done = make(chan struct{})
	tcpSrv.statusLock.Unlock()

	for tcpSrv.status == SrvStatusRunning {
		err := tcpSrv.listener.SetDeadline(time.Now().Add(time.Second))
		if err != nil {
			fmt.Printf("[fatal] accept timeout failed, err=%+v\n", err)
			continue
		}
		conn, err := tcpSrv.listener.Accept()
		if err != nil {
			op, ok := err.(*net.OpError)
			if ok && op.Timeout() {
				// Accept超时
				continue
			}
			if ok && op.Temporary() {
				// 被信号唤醒，处理信号
				continue
			}
			fmt.Printf("unknown error <%+v>\n", err)
		}
		// 处理连接
		tcpSrv.handleConn(conn)
	}
	close(tcpSrv.done)

	return nil
}

func (tcpSrv *TcpServer) handleConn(conn net.Conn) {
	go func() {
		tcpClient := newTcpClient(conn, tcpSrv)
		defer tcpClient.Close()
		if err := tcpClient.Serv(); err != nil {
			fmt.Printf("connection aborted, last err: %s\n", err)
		}
	}()
}

func (tcpSrv *TcpServer) Stop() error {
	tcpSrv.statusLock.Lock()
	defer tcpSrv.statusLock.Unlock()

	if tcpSrv.status == SrvStatusStopping || tcpSrv.status == SrvStatusStopped {
		return ServerAlreadyStopped
	}

	tcpSrv.status = SrvStatusStopping
	<-tcpSrv.done // 等待Server退出
	tcpSrv.status = SrvStatusStopped

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
