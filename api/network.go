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

const (
	_ NetEntityType = iota
	NetListener
	NetConnector
)

type (
	NetEntityType int

	NetPacketType byte

	NetRouterFunc func(session *Session, msg *NetworkMessage) (*Pid, string, *Error)

	NetworkPacket struct {
		Type NetPacketType
		Data []byte
	}

	INetServer interface {
		IModule
		Link(session INetEntity, c gnet.Conn)
		Ref(c gnet.Conn) INetEntity
		Unlink(c gnet.Conn)
		Typ() NetEntityType
	}

	INetPacket interface {
		GetTyp() NetPacketType
		GetData() []byte
		String() string
	}

	INetPackCodec interface {
		Encode(packet INetPacket) []byte
		Decode(reader gnet.Reader) []INetPacket
	}

	INetRouter interface {
		GetService() string
		GetActorId() uint64
		GetMethod() string
	}

	INetEntity interface {
		ID() uint64
		Type() NetEntityType
		Network() string
		LocalAddr() string
		RemoteAddr() string
		Traffic(c gnet.Conn) error
		SendPacket(packet *NetworkPacket) *Error
		SendMessage(message *NetworkMessage) *Error
		Close(reason error) *Error
		Closed(err error) *Error
		Node() INode
		RawCon() gnet.Conn
		GetAgent() *Pid
		Session() *Session
	}

	NetworkMessage struct {
		Id   uint16
		Data []byte
	}

	Session struct {
		Agent  *Pid
		Mid    uint16
		entity INetEntity
	}
)

func (p *NetworkPacket) GetTyp() NetPacketType {
	return p.Type
}
func (p *NetworkPacket) GetData() []byte {
	return p.Data
}
func (p *NetworkPacket) String() string {
	return fmt.Sprintf("NetPacketType: %d,  Data: %s", p.Type, string(p.Data))
}

func NewNetworkMessage(Id uint16, Data []byte) *NetworkMessage {
	return &NetworkMessage{
		Id:   Id,
		Data: Data,
	}
}

func (m *NetworkMessage) GetID() uint16   { return m.Id }
func (m *NetworkMessage) GetData() []byte { return m.Data }

func NewSession(entity INetEntity) *Session {
	ses := new(Session)
	ses.Agent = entity.GetAgent()
	ses.entity = entity
	return ses
}
func (s *Session) GetEntity() INetEntity {
	return s.entity
}
