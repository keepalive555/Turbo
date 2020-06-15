// @Author: wanglanwei@baidu.com
// @Description: Socks5 客户端
// @Reference: https://tools.ietf.org/html/rfc1928
package server

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	MaxMethodSize = 0xff // Socks5协议最多支持255种认证方法
	IPv4Size      = 0x04 // IPv4地址4字节
	IPv6Size      = 0x10 // IPv6地址16字节
	PortSize      = 0x02 // 端口号2字节
)

// 状态
const (
	StatusTcpHandshake        int = 0x00 // Tcp握手成功（默认状态）
	StatusCSTalked                = 0x01 // C/S端认证方式已协商
	StatusRemoteHostConnected     = 0x02 // 已连接远程主机
	StatusFullDuplexComm          = 0x03 // 全双工通讯
	StatusClosing                 = 0x10 // 连接正在关闭
	StatusClosed                  = 0xff // 连接已关闭
)

type TcpClient struct {
	tcpSrv        *TcpServer     // tcp服务器
	conn          *net.TCPConn   // tcp连接
	err           unsafe.Pointer // 客户端lastErr
	authMethod    byte           // 认证方法
	status        int            // 客户端状态
	remoteAddress string         // 远程地址
	remoteConn    net.Conn       // 远程连接
}

// 新建Tcp客户端
func newTcpClient(conn net.Conn, tcpSrv *TcpServer) *TcpClient {
	client := &TcpClient{
		conn:       conn.(*net.TCPConn),
		tcpSrv:     tcpSrv,
		err:        nil,
		authMethod: NoAuthorization,
	}
	runtime.SetFinalizer(client, finalizer)
	// 关闭tcp连接Nagle算法，防止Nagle+Delayed ACK出现小数据Write-Write-Read的问题
	tcpClient.conn.SetNoDelay(true)
	// 开启tcp keepalive功能
	return client
}

// 获取客户端错误
func (tcpClient *TcpClient) LastErr() error {
	p := atomic.LoadPointer(&tcpClient.err)
	if p == nil {
		return nil
	}
	return *(*error)(p)
}

// 设置客户端错误
func (tcpClient *TcpClient) setErr(i interface{}) {
	atomic.StorePointer(&tcpClient.err, unsafe.Pointer(&i))
}

// 关闭连接
func (tcpClient *TcpClient) Close() (err error) {
	if tcpClient != nil && tcpClient.conn != nil {
		err = tcpClient.conn.Close()
	}
	if tcpClient != nil && tcpClient.remoteConn != nil {
		err = tcpClient.remoteConn.Close()
	}
	return err
}

// 状态处理
func (tcpClient *TcpClient) Serv() (err error) {
	for {
		switch tcpClient.status {
		case StatusTcpHandshake:
			err = tcpClient.handleCSTalk()
		case StatusCSTalked:
			err = tcpClient.handleRemoteConn()
		case StatusRemoteHostConnected:
			err = tcpClient.handleFullDuplexComm()
		case StatusClosing:
			err = tcpClient.handleClose()
		case StatusClosed:
			return tcpClient.LastErr()
		}

		if err != nil {
			tcpClient.setErr(err)
			tcpClient.Close()
			break
		}
	}
	return err
}

func (tcpClient *TcpClient) handleFullDuplexComm() (err error) {
	// 双向通讯
	// TODO: @wanglanwei 移除io.copy实现方式，增加channel处理
	tcpClient.status = StatusFullDuplexComm
	done := make(chan struct{}, 1)
	go func() {
		_, err := io.Copy(tcpClient.remoteConn, tcpClient.conn)
		if err != nil && err != io.EOF {
			tcpClient.setErr(err)
		}
		done <- struct{}{}
	}()
	_, err = io.Copy(tcpClient.conn, tcpClient.remoteConn)
	if err != nil && err != io.EOF {
		return err
	}
	<-done
	tcpClient.status = StatusClosing
	return nil
}

func (tcpClient *TcpClient) handleRemoteConn() error {
	header := &TcpCmdHeader{}
	err := binary.Read(tcpClient.conn, binary.LittleEndian, header)
	if err != nil {
		return err
	}
	if header.Version != MagicVersion {
		return BadRequest
	}

	var host string
	var port int
	switch header.AddressType {
	case AddrFamilyIPv4:
		// 解析IPv4数据
		// 数据格式：|--ipv4--|--port--|
		data := make([]byte, IPv4Size+PortSize, IPv4Size+PortSize) // 4字节IP，2字节网络字节序端口
		n, err := io.ReadFull(tcpClient.conn, data)
		if err != nil {
			return err
		}
		if n != (IPv4Size + PortSize) {
			return BadRequest
		}
		host = net.IPv4(data[0], data[1], data[2], data[3]).String()
		port = int(data[n-2])<<8 | int(data[n-1])
	case AddrFamilyDomain:
		// 解析域名
		// 数据格式：|--len--|--domain--|--port--|
		domainLength := make([]byte, 1, 1)
		n, err := io.ReadFull(tcpClient.conn, domainLength) // 1字节长度
		if err != nil {
			return err
		}
		if n != 1 {
			return BadRequest
		}
		data := make([]byte, int(domainLength[0])+2, int(domainLength[0])+2)
		n, err = io.ReadFull(tcpClient.conn, data)
		if n != (int(domainLength[0]) + 2) {
			return BadRequest
		}
		host = string(data[:n-2])
		port = int(data[n-2])<<8 | int(data[n-1])
	case AddrFamilyIPv6:
		// 解析IPv6数据
		// 数据格式：|--ipv6--|--port--|
		data := make([]byte, IPv6Size+PortSize, IPv6Size+PortSize)
		n, err := io.ReadFull(tcpClient.conn, data) // 16字节IPv6，2字节网络字节序端口
		if err != nil {
			return err
		}
		if n != (IPv6Size + PortSize) {
			return BadRequest
		}
		host = net.IP{
			data[0], data[1], data[2], data[3],
			data[4], data[5], data[6], data[7],
			data[8], data[9], data[10], data[11],
			data[12], data[13], data[14], data[15],
		}.String()
		port = int(data[n-2])<<8 | int(data[n-1])
	default:
		return BadRequest
	}
	tcpClient.remoteAddress = fmt.Sprintf("%s:%d", host, port)
	fmt.Printf("remote host: %s\n", tcpClient.remoteAddress)

	// 连接目标主机
	timeout := time.Duration(tcpClient.tcpSrv.Config().RemoteConnTimeout)
	conn, err := net.DialTimeout(ProtoTCP, tcpClient.remoteAddress, timeout*time.Millisecond)
	if err != nil {
		return err
	}
	// 响应客户端
	// TODO: @wanglanwei 优化细节
	_, err = tcpClient.conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		return err
	}
	tcpClient.remoteConn = conn
	tcpClient.status = StatusRemoteHostConnected
	return nil
}

func (tcpClient *TcpClient) handleClose() (err error) {
	err = tcpClient.Close()
	tcpClient.status = StatusClosed
	return err
}

// Socks5协议握手
func (tcpClient *TcpClient) handleCSTalk() error {
	syncHeader := &TcpSyncHeader{}
	err := binary.Read(tcpClient.conn, binary.LittleEndian, syncHeader)
	if err != nil {
		return err
	}
	if syncHeader.Version != MagicVersion {
		return BadRequest
	}
	methods := make([]byte, MaxMethodSize, MaxMethodSize)
	n, err := tcpClient.conn.Read(methods)
	if err != nil {
		return err
	}
	if n != int(syncHeader.NumberOfMethods) {
		return BadRequest
	}
	// 检测客户端同步的认证方法，Server端是否支持
	var method byte = NoAcceptableMethods
	for i := 0; i < n; i++ {
		if methods[i] == tcpClient.AuthMethod {
			method = tcpClient.AuthMethod
			break
		}
	}

	replyHeader := &TcpReplySyncHeader{
		Version: MagicVersion,
		Method:  method,
	}
	// 响应Socks5客户端
	err = binary.Write(tcpClient.conn, binary.LittleEndian, replyHeader)
	if err != nil {
		return err
	}
	if method != NoAcceptableMethods {
		// 修改状态
		tcpClient.status = StatusCSTalked
		fmt.Printf("client/server talk succeed, peer address: %s\n", tcpClient.conn.RemoteAddr())
		return nil
	}
	return BadRequest
}

// 析构函数，析构时释放本地与远程连接
func finalizer(obj *TcpClient) {
	obj.Close()
}
