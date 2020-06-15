// @Author: wanglanwei@baidu.com
// @Description: Socks5 Tcp Server
// @Reference: https://tools.ietf.org/html/rfc1928
package server

import (
	// "errors"
	"fmt"
)

const (
	MaxConnections            = 10000     // 硬连接上限
	DefaultPort               = 8000      // 默认端口号
	DefaultHost               = "0.0.0.0" // 默认主机
	DefaultLocalReadTimeout   = 2000      // 默认本地读超时2000ms
	DefaultLocalWriteTimeout  = 2000      // 默认本地写超时2000ms
	DefaultRemoteConnTimeout  = 2000      // 默认远程主机连接超时2000ms
	DefaultRemoteReadTimeout  = 2000      // 默认远程主机读超时2000ms
	DefaultRemoteWriteTimeout = 2000      // 默认远程主机写超时2000ms
)

// Tcp Server配置
type TcpConfig struct {
	MaxConnections     int    `json:"max_connections"`      // 客户端最大连接数
	Host               string `json:"host"`                 // 监听主机
	Port               int    `json:"port"`                 // 监听端口
	LocalReadTimeout   int    `json:"local_read_timeout"`   // 本地读超时，单位：毫秒
	LocalWriteTimeout  int    `json:"local_write_timeout"`  // 本地写超时，单位：毫秒
	RemoteConnTimeout  int    `json:"remote_conn_timeout"`  // 远程连接超时，单位：毫秒
	RemoteReadTimeout  int    `json:"remote_read_timeout"`  // 远程读超时，单位：毫秒
	RemoteWriteTimeout int    `json:"remote_write_timeout"` // 远程写超时，单位：毫秒
}

func (tcpConfig *TcpConfig) Check() error {
	if tcpConfig.MaxConnections < 0 {
		tcpConfig.MaxConnections = 0
	}
	if tcpConfig.Host == "" {
		tcpConfig.Host = DefaultHost
	}
	if tcpConfig.Port <= 0 {
		tcpConfig.Port = DefaultPort
	}
	if tcpConfig.Port > 65535 {
		return fmt.Errorf("tcp server port <%d> is invalid", tcpConfig.Port)
	}
	if tcpConfig.LocalReadTimeout < 0 {
		tcpConfig.LocalReadTimeout = DefaultLocalReadTimeout
	}
	if tcpConfig.LocalWriteTimeout < 0 {
		tcpConfig.LocalWriteTimeout = DefaultLocalWriteTimeout
	}
	if tcpConfig.RemoteConnTimeout < 0 {
		tcpConfig.RemoteConnTimeout = DefaultRemoteConnTimeout
	}
	if tcpConfig.RemoteReadTimeout < 0 {
		tcpConfig.RemoteReadTimeout = DefaultRemoteReadTimeout
	}
	if tcpConfig.RemoteWriteTimeout < 0 {
		tcpConfig.RemoteWriteTimeout = DefaultRemoteWriteTimeout
	}
	return nil
}
