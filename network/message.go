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
	"github.com/dingqinghui/gas/api"
)

var msgCodec = new(BuiltinMsgCodec)

const MsgIdOffset = 2

type BuiltinMsgCodec struct{}

func (codec *BuiltinMsgCodec) Decode(buf []byte) *api.NetworkMessage {
	msg := new(api.NetworkMessage)
	msg.Id = binary.BigEndian.Uint16(buf[:MsgIdOffset])
	msg.Data = buf[MsgIdOffset:]
	return msg
}
func (codec *BuiltinMsgCodec) Encode(msg *api.NetworkMessage) []byte {
	buf := make([]byte, MsgIdOffset+len(msg.GetData()))
	binary.BigEndian.PutUint16(buf, msg.GetID())
	copy(buf[MsgIdOffset:], msg.GetData())
	return buf
}
