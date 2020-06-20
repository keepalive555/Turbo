package local

import ()

type Server interface {
	Proto() string           // TCP或UDP
	Address() string         // 服务器监听地址
	Run() error              // 运行服务器
	Stop() error             // 停止服务器
	Statistics() *Statistics // 获取服务器统计信息（故障排查，运维信息统计）

	// IsRunning() bool         // 是否运行
	// IsStopped() bool         // 是否停止
}

func NewServer(proto string, args ...interface{}) (Server, error) {
	switch proto {
	case ProtoTCP:
		if len(args) < 1 {
			return newTcpServer(nil)
		}
		if tcpConfig, ok := args[0].(*TcpConfig); ok {
			return newTcpServer(tcpConfig)
		}
		return nil, BadArguments
	case ProtoUDP:
		panic(NotImplementedYet)
	}
	panic(UnknownProto)
}
