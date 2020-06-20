package local

import (
	"errors"
	"time"
)

const (
	ProtoTCP = "tcp"
	ProtoUDP = "udp"
)

// 连接信息
type Connections struct {
	TotalConnectionCount int64
	ConnectionCount      int64
}

// 统计信息
type Statistics struct {
	StartTime time.Time // 启动时间
	Connections
}

// 异常
var (
	BadArguments         = errors.New("bad arguments")
	UnknownProto         = errors.New("unknown proto")
	NotImplementedYet    = errors.New("not implemented yet")
	BadRequest           = errors.New("bad request")
	ServerAlreadyStarted = errors.New("server already started")
	ServerAlreadyStopped = errors.New("server already stopped")
)

const (
	MagicVersion byte = 0x05 // Socks5协议魔数
)

const (
	NoAuthorization     byte = 0x00 // 无需授权
	GSSAPI                   = 0x01 // GSSAPI
	Password                 = 0x02 // 帐号名，密码
	NoAcceptableMethods      = 0xff // 无可接受方法
	// 0x03 - 0x7f IANA保留
	// 0x80 - 0xfe 用户保留方法
)

const (
	CmdConnect      byte = 0x01 // Connect命令
	CmdBind              = 0x02 // Bind命令
	CmdUdpAssociate      = 0x03 // UDP ASSOCIATE命令
)

const (
	AddrFamilyIPv4   byte = 0x01 // IPv4地址
	AddrFamilyDomain      = 0x03 // Domain
	AddrFamilyIPv6        = 0x04 // IPv6地址
)

type TcpSyncHeader struct {
	Version         byte // 协议版本，固定为：0x05
	NumberOfMethods byte // 方法数组
}

type TcpReplySyncHeader struct {
	Version byte // 协议版本，固定为：0x05
	Method  byte // 服务端选择的版本
}

type TcpCmdHeader struct {
	Version     byte // 协议版本，固定为：0x05
	Cmd         byte // 命令
	Reserved    byte // 保留，取值0x00
	AddressType byte // 地址类型
}

type TcpReplyCmdHeader struct {
	Version byte // 协议版本，固定为：0x05
}
