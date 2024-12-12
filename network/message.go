/**
 * @Author: dingQingHui
 * @Description:
 * @File: message
 * @Version: 1.0.0
 * @Date: 2024/12/6 15:52
 */

package network

import (
	"encoding/binary"
)

var msgCodec = new(BuiltinMsgCodec)

type BuiltinMessage struct {
	id   uint16
	data []byte
}

func (m *BuiltinMessage) GetID() uint16   { return m.id }
func (m *BuiltinMessage) GetData() []byte { return m.data }

const MsgIdOffset = 2

type BuiltinMsgCodec struct{}

func (codec *BuiltinMsgCodec) Decode(buf []byte) *BuiltinMessage {
	msg := new(BuiltinMessage)
	msg.id = binary.BigEndian.Uint16(buf[:MsgIdOffset])
	msg.data = buf[MsgIdOffset:]
	return msg
}
func (codec *BuiltinMsgCodec) Encode(msg *BuiltinMessage) []byte {
	buf := make([]byte, MsgIdOffset+len(msg.GetData()))
	binary.BigEndian.PutUint16(buf, msg.GetID())
	copy(buf[MsgIdOffset:], msg.GetData())
	return buf
}
