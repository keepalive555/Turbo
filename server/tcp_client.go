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
	"time"
)

const (
	MaxMethodCount = 0xff // Socks5协议最多支持255种认证方法
	IPv4Size       = 0x04 // IPv4地址4字节
	IPv6Size       = 0x10 // IPv6地址16字节
	PortSize       = 0x02 // 端口号2字节
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
	tcpSrv        *TcpServer   // Tcp服务器
	conn          *net.TCPConn // tcp连接
	err           error        // 客户端lastErr
	status        int          // 客户端状态
	remoteAddress string       // 远程地址
	remoteConn    net.Conn     // 远程连接
}

// 析构函数
func finalizer(obj *TcpClient) {
	obj.Close()
}

// 新建Tcp客户端
func newTcpClient(conn net.Conn, tcpSrv *TcpServer) *TcpClient {
	client := &TcpClient{
		conn:   conn.(*net.TCPConn),
		tcpSrv: tcpSrv,
	}
	client.init()
	runtime.SetFinalizer(client, finalizer)
	return client
}

// 初始化
func (tcpClient *TcpClient) init() {
	// 关闭连接Nagle算法，防止Nagle+Delayed ACK出现小数据Write-Write-Read的问题
	tcpClient.conn.SetNoDelay(true)
	// 开启TCP保活机制：代理无法干预业务层数据，无法使用业务数据做保活
}

// 记录最后的Error
func (tcpClient *TcpClient) LastErr() error {
	return tcpClient.err
}

// 关闭连接
func (tcpClient *TcpClient) Close() error {
	if tcpClient != nil && tcpClient.conn != nil {
		return tcpClient.conn.Close()
	}
	return nil
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
			return tcpClient.err
		}

		if err != nil {
			tcpClient.err = err
			break
		}
	}
	return err
}

func (tcpClient *TcpClient) handleFullDuplexComm() (err error) {
	// 全双工通讯
	tcpClient.status = StatusFullDuplexComm
	done := make(chan struct{}, 1)
	go func() {
		_, err := io.Copy(tcpClient.remoteConn, tcpClient.conn)
		if err != nil && err != io.EOF {
			tcpClient.err = err
		}
		done <- struct{}{}
	}()
	_, err = io.Copy(tcpClient.conn, tcpClient.remoteConn)
	if err != nil {
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
		tiny := make([]byte, 1, 1)
		n, err := io.ReadFull(tcpClient.conn, tiny) // 1字节长度
		if err != nil {
			return err
		}
		if n != 1 {
			return BadRequest
		}
		data := make([]byte, int(tiny[0])+2, int(tiny[0])+2) // length字节域名，2字节端口
		n, err = io.ReadFull(tcpClient.conn, data)
		if n != (int(tiny[0]) + 2) {
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
	conn, err := net.DialTimeout("tcp", tcpClient.remoteAddress, timeout*time.Millisecond)
	if err != nil {
		return err
	}
	// 响应客户端
	_, err = tcpClient.conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		return err
	}
	tcpClient.remoteConn = conn
	tcpClient.status = StatusRemoteHostConnected
	return nil
}

func (tcpClient *TcpClient) handleClose() (err error) {
	if tcpClient.conn != nil {
		err = tcpClient.conn.Close()
	}
	if tcpClient.remoteConn != nil {
		err = tcpClient.remoteConn.Close()
	}
	tcpClient.status = StatusClosed
	return err
}

// SOCKS5握手协议
func (tcpClient *TcpClient) handleCSTalk() error {
	syncHeader := &TcpSyncHeader{}
	err := binary.Read(tcpClient.conn, binary.LittleEndian, syncHeader)
	if err != nil {
		return err
	}
	if syncHeader.Version != MagicVersion {
		return BadRequest
	}
	methods := make([]byte, MaxMethodCount, MaxMethodCount)
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
		// TODO: @wanglanwei 认证方法先写死，后续扩展
		if methods[i] == NoAuthorization {
			method = NoAuthorization
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
