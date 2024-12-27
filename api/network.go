/**
 * @Author: dingQingHui
 * @Description:
 * @File: net
 * @Version: 1.0.0
 * @Date: 2024/11/25 19:04
 */

package api

import (
	"github.com/dingqinghui/gas/network/message"
	"github.com/dingqinghui/gas/network/packet"
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

	INetServer interface {
		IModule
		Link(session INetEntity, c gnet.Conn)
		Ref(c gnet.Conn) INetEntity
		Unlink(c gnet.Conn)
		Typ() NetEntityType
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
		SendRaw(typ packet.Type, data []byte) *Error
		SendMessage(msg *message.Message) *Error
		//Response(session *Session, payload interface{}) *Error
		//ResponseErr(session *Session, err *Error) *Error
		//Push(mid uint16, payload interface{}) *Error
		Close(reason *Error) *Error
		Closed(err error) *Error
		RawCon() gnet.Conn
		GetAgent() *Pid
		Session() *Session
		Kick(reason *Error) *Error
	}

	NetworkMessage struct {
		Id      uint16
		ErrCode uint16
		Data    []byte
	}
)

func (m *NetworkMessage) GetID() uint16   { return m.Id }
func (m *NetworkMessage) GetData() []byte { return m.Data }
