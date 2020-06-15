package client

type SyncHeader struct {
	Magic     uint16
	SessionId uint64 // 会话ID，默认为0
}

const (
	TypeNormal    byte = 0x00 // 正常数据包
	TypeHeartBeat      = 0x01 // 心跳数据包
	TypeOneway         = 0x02 // 单向数据包
)

type Header struct {
	Magic       byte   // 魔数
	Type        byte   // 包类型
	PayloadSize uint16 // 数据大小
}
