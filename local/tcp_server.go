// @Author: wanglanwei@baidu.com
// @Description: Socks5 Tcp Server
// @Reference: https://tools.ietf.org/html/rfc1928
package local

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
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
	quit       chan struct{}    // Server退出
	status     int              // 服务器状态
	statistics *Statistics      // 服务器运行信息
}

func (tcpSrv *TcpServer) Proto() string {
	return ProtoTCP
}

func (tcpSrv *TcpServer) Address() string {
	return fmt.Sprintf("%s:%d", tcpSrv.host, tcpSrv.port)
}

func (tcpSrv *TcpServer) Statistics() *Statistics {
	return tcpSrv.statistics
}

func (tcpSrv *TcpServer) Done() <-chan struct{} {
	return tcpSrv.quit
}

func (tcpSrv *TcpServer) Run() error {
	tcpSrv.statusLock.Lock()
	// 运行中，直接返回
	if tcpSrv.status == SrvStatusRunning {
		tcpSrv.statusLock.Unlock()
		return ServerAlreadyStarted
	}
	// 启动服务器
	ln, err := net.Listen(ProtoTCP, tcpSrv.Address())
	if err != nil {
		tcpSrv.statusLock.Unlock()
		return err
	}
	fmt.Printf("server <%s> started!\n", tcpSrv.Address())
	tcpSrv.listener = ln.(*net.TCPListener)
	tcpSrv.status = SrvStatusRunning
	tcpSrv.quit = make(chan struct{})
	tcpSrv.statusLock.Unlock()

	// 启动成功，则记录启动时间
	tcpSrv.statistics.StartTime = time.Now()
	for tcpSrv.status == SrvStatusRunning {
		err := tcpSrv.listener.SetDeadline(time.Now().Add(time.Second))
		if err != nil {
			fmt.Printf("accept timeout failed, err=%+v\n", err)
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
		// 达到连接上限，关闭连接
		// 达到连接上限，仍需要Accept调用，从操作系统全连接队列中取出连接
		// 然后关闭连接，防止连接堆积
		if tcpSrv.statistics.ConnectionCount+1 >= int64(tcpSrv.config.MaxConnections) {
			conn.Close()
			continue
		}
		// 连接成功，更新连接数
		atomic.AddInt64(&tcpSrv.statistics.ConnectionCount, 1)
		// 处理连接
		tcpSrv.handleConn(conn)
	}

	// 等待所有连接，处理完毕退出
	close(tcpSrv.quit)

	return nil
}

func (tcpSrv *TcpServer) handleConn(conn net.Conn) {
	go func() {
		tcpClient := newTcpClient(conn, tcpSrv, tcpSrv.config)
		defer tcpClient.Close()
		if err := tcpClient.Serv(); err != nil {
			fmt.Printf("connection aborted, last err: %s\n", err)
		}
		atomic.AddInt64(&tcpSrv.statistics.ConnectionCount, -1)
	}()
}

func (tcpSrv *TcpServer) Stop() error {
	tcpSrv.statusLock.Lock()
	defer tcpSrv.statusLock.Unlock()

	if tcpSrv.status == SrvStatusStopping || tcpSrv.status == SrvStatusStopped {
		return ServerAlreadyStopped
	}

	tcpSrv.status = SrvStatusStopping
	<-tcpSrv.quit // 等待Server退出
	tcpSrv.status = SrvStatusStopped

	return nil
}

func newTcpServer(tcpConfig *TcpConfig) (*TcpServer, error) {
	tcpServer := &TcpServer{
		host:       tcpConfig.Host,
		port:       tcpConfig.Port,
		config:     tcpConfig,
		statistics: &Statistics{},
	}
	return tcpServer, nil
}
