/**
 * @Author: dingQingHui
 * @Description:
 * @File: pack
 * @Version: 1.0.0
 * @Date: 2024/11/12 10:33
 */

package packet

import (
	"encoding/binary"
	"fmt"
	"github.com/panjf2000/gnet/v2"
)

type Type = byte

const (
	// packet type
	_                Type = iota
	HandshakeType         = 0x01
	HandshakeAckType      = 0x02
	HeartbeatType         = 0x03
	DataType              = 0x04
	KickType              = 0x05 // disconnect message from server

	HeadLength    = 1 + 2
	MaxPacketSize = 64 * 1024
)

type NetworkPacket struct {
	Type Type
	Len  uint16
	Data []byte
}

func (p *NetworkPacket) GetTyp() Type {
	return p.Type
}
func (p *NetworkPacket) GetData() []byte {
	return p.Data
}
func (p *NetworkPacket) String() string {
	return fmt.Sprintf("NetPacketType: %d,  Data: %s", p.Type, string(p.Data))
}

// Encode
// @Description:
// @param packet
// @return []byte
// -<type>-|--------<length>--------|-<data>-
// --------|------------------------|--------
// 1 byte packet type, 2 bytes packet data length(big end), and data segment
func Encode(typ Type, data []byte) []byte {
	msgLen := HeadLength + len(data)
	buf := make([]byte, msgLen)
	buf[0] = typ
	binary.BigEndian.PutUint16(buf[1:HeadLength], uint16(len(data)))
	copy(buf[HeadLength:msgLen], data)
	return buf
}

func Decode(reader gnet.Reader) []*NetworkPacket {
	if reader == nil {
		return nil
	}
	var packets []*NetworkPacket
	for {
		buf, _ := reader.Peek(HeadLength)
		if len(buf) < HeadLength {
			break
		}
		typ := buf[0]
		bodyLen := binary.BigEndian.Uint16(buf[1:HeadLength])
		msgLen := HeadLength + int(bodyLen)
		if msgLen >= MaxPacketSize {
			break
		}
		if reader.InboundBuffered() < msgLen {
			break
		}
		buf, err := reader.Peek(msgLen)
		if buf == nil || err != nil {
			return nil
		}
		_, _ = reader.Discard(msgLen)
		p := new(NetworkPacket)
		p.Type = typ
		p.Data = buf[HeadLength:msgLen]
		p.Len = bodyLen
		packets = append(packets, p)
	}
	return packets
}
