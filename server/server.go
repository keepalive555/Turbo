package server

import ()

type Server interface {
	Proto() string   // TCP或UDP
	Address() string // 服务器监听地址
	Run() error      // 运行服务器
	Stop() error     // 停止服务器
}

func NewServer(proto string, args ...interface{}) (Server, error) {
	switch proto {
	case ProtoTCP:
		if len(args) < 1 {
			return nil, BadArguments
		}
		if tcpConfig, ok := args[0].(*TcpConfig); ok {
			return newTcpServer(tcpConfig)
		}
		return nil, BadArguments
	case ProtoUDP:
	}
	panic(UnknownProto)
}
