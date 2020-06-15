// @Author: wanglanwei@baidu.com
// @Description: Socks5 Tcp Server
// @Reference: https://tools.ietf.org/html/rfc1928
package server

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

func (tcpConn *TcpConfig) Validate() bool {
	return true
}
