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

type NetConnectionType int

const (
	_ NetConnectionType = iota
	NetListener
	NetConnector
)

type NetPacketType byte

type INetServer interface {
	IModule
	Link(session ISession, c gnet.Conn)
	Ref(c gnet.Conn) ISession
	Unlink(c gnet.Conn)
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

type ISession interface {
	ID() uint64
	Type() NetConnectionType
	Network() string
	LocalAddr() string
	RemoteAddr() string
	Traffic(c gnet.Conn) error
	SendPacket(packet *BuiltinNetworkPacket) error
	SendMessage(msgId uint16, s2c interface{}) error
	Close(reason error) error
	Closed(err error) error
	Node() INode
	RawCon() gnet.Conn
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
