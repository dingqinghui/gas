/**
 * @Author: dingQingHui
 * @Description:
 * @File: message
 * @Version: 1.0.0
 * @Date: 2024/12/6 15:52
 */

package message

import (
	"encoding/binary"
)

const HeadLen = 8

func Encode(msg *Message) []byte {
	buf := make([]byte, HeadLen+len(msg.Data))
	offset := 0
	binary.BigEndian.PutUint16(buf[offset:], msg.ID)
	offset += 2
	binary.BigEndian.PutUint16(buf[offset:], msg.Error)
	offset += 2
	binary.BigEndian.PutUint32(buf[offset:], msg.Index)
	offset += 4
	copy(buf[offset:], msg.Data)
	return buf
}

func Decode(buf []byte) *Message {
	msg := New()
	offset := 0
	msg.ID = binary.BigEndian.Uint16(buf[offset : offset+2])
	offset += 2
	msg.Error = binary.BigEndian.Uint16(buf[offset : offset+2])
	offset += 2
	msg.Index = binary.BigEndian.Uint32(buf[offset : offset+4])
	offset += 4
	msg.Data = buf[offset:]
	return msg
}

func New() *Message {
	return &Message{
		Head: new(Head),
		Data: nil,
	}
}

func NewWithData(data []byte) *Message {
	return &Message{
		Head: new(Head),
		Data: data,
	}
}

func NewErr(code uint16) *Message {
	m := New()
	m.Error = code
	return m
}

type Message struct {
	*Head
	Data []byte
}

func (m *Message) Copy(old *Message) {
	m.Index = old.Index
	m.ID = old.ID
}

type Head struct {
	ID    uint16 // 命令
	Error uint16 // 错误码
	Index uint32 // 序号
}
