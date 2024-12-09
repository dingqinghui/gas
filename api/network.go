/**
 * @Author: dingQingHui
 * @Description:
 * @File: net
 * @Version: 1.0.0
 * @Date: 2024/11/25 19:04
 */

package api

import (
	"github.com/panjf2000/gnet/v2"
)

type NetServerType int

const (
	_ NetServerType = iota
	NetServiceListener
	NetServiceConnector
)

type NetPacketType byte

type INetServer interface {
	IModule
	gnet.EventHandler
	Type() NetServerType
}

type INetPacket interface {
	GetTyp() NetPacketType
	GetData() []byte
	String() string
}

type INetPackCodec interface {
	Encode(packet INetPacket) []byte
	Decode(reader gnet.Reader) []INetPacket
}

type INetRouter interface {
	Register(msgId uint16, methodName string)
	Get(msgId uint16) string
}
