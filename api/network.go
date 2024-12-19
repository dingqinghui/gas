/**
 * @Author: dingQingHui
 * @Description:
 * @File: net
 * @Version: 1.0.0
 * @Date: 2024/11/25 19:04
 */

package api

import (
	"fmt"
	"github.com/panjf2000/gnet/v2"
)

type NetEntityType int

const (
	_ NetEntityType = iota
	NetListener
	NetConnector
)

type NetPacketType byte

type INetServer interface {
	IModule
	Link(session INetEntity, c gnet.Conn)
	Ref(c gnet.Conn) INetEntity
	Unlink(c gnet.Conn)
	Typ() NetEntityType
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

type INetRouters interface {
	Add(msgId uint16, router INetRouter)
	Get(msgId uint16) INetRouter
}

type INetRouter interface {
	GetNodeType() string
	GetActorId() uint64
	GetMethod() string
}

type INetEntity interface {
	ID() uint64
	Type() NetEntityType
	Network() string
	LocalAddr() string
	RemoteAddr() string
	Traffic(c gnet.Conn) error
	SendPacket(packet *BuiltinNetworkPacket) *Error
	SendMessage(msgId uint16, s2c interface{}) *Error
	SendRawMessage(msgId uint16, data []byte) *Error
	Close(reason error) *Error
	Closed(err error) *Error
	Node() INode
	RawCon() gnet.Conn
	GetAgent() *Pid
}

type BuiltinNetworkPacket struct {
	Type NetPacketType
	Data []byte
}

func (p *BuiltinNetworkPacket) GetTyp() NetPacketType {
	return p.Type
}
func (p *BuiltinNetworkPacket) GetData() []byte {
	return p.Data
}
func (p *BuiltinNetworkPacket) String() string {
	return fmt.Sprintf("NetPacketType: %d,  Data: %s", p.Type, string(p.Data))
}
