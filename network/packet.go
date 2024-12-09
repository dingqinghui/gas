/**
 * @Author: dingQingHui
 * @Description:
 * @File: pack
 * @Version: 1.0.0
 * @Date: 2024/11/12 10:33
 */

package network

import (
	"encoding/binary"
	"fmt"
	"github.com/dingqinghui/gas/api"
	"github.com/panjf2000/gnet/v2"
)

const (
	_                      api.NetPacketType = iota
	PacketTypeHandshake                      = 0x01
	PacketTypeHandshakeAck                   = 0x02
	PacketTypeHeartbeat                      = 0x03
	PacketTypeData                           = 0x04
	PacketTypeKick                           = 0x05 // disconnect message from server
)

var (
	KickPacket         = NewPacket(PacketTypeKick, nil)
	HandshakeAckPacket = NewPacket(PacketTypeHandshakeAck, nil)
)

func NewHandshakePacket(buf []byte) *BuiltinNetworkPacket {
	return NewPacket(PacketTypeHandshake, buf)
}

func NewDataPacket(data []byte) *BuiltinNetworkPacket {
	return &BuiltinNetworkPacket{Type: PacketTypeData, Data: data}
}

func NewPacket(tye api.NetPacketType, data []byte) *BuiltinNetworkPacket {
	return &BuiltinNetworkPacket{Type: tye, Data: data}
}

type BuiltinNetworkPacket struct {
	Type api.NetPacketType
	Data []byte
}

func (p *BuiltinNetworkPacket) GetTyp() api.NetPacketType {
	return p.Type
}

func (p *BuiltinNetworkPacket) GetData() []byte {
	return p.Data
}

func (p *BuiltinNetworkPacket) String() string {
	return fmt.Sprintf("NetPacketType: %d,  Data: %s", p.Type, string(p.Data))
}

const (
	HeadLength    = 2 + 4
	MaxPacketSize = 64 * 1024
	msgDataOffset = 4
)

var packCodec = &BuiltinPacketCodec{}

type BuiltinPacketCodec struct{}

func (codec *BuiltinPacketCodec) Encode(packet api.INetPacket) []byte {
	msgLen := HeadLength + len(packet.GetData())
	buf := make([]byte, msgLen)
	binary.BigEndian.PutUint16(buf[:2], uint16(packet.GetTyp()))
	binary.BigEndian.PutUint32(buf[2:HeadLength], uint32(len(packet.GetData())))
	copy(buf[HeadLength:msgLen], packet.GetData())
	return buf
}

func (codec *BuiltinPacketCodec) Decode(reader gnet.Reader) []*BuiltinNetworkPacket {
	if reader == nil {
		return nil
	}
	var packets []*BuiltinNetworkPacket
	for {
		buf, _ := reader.Peek(HeadLength)
		if len(buf) < HeadLength {
			break
		}
		typ := binary.BigEndian.Uint16(buf[:2])
		bodyLen := binary.BigEndian.Uint32(buf[2:HeadLength])
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
		p := new(BuiltinNetworkPacket)
		p.Type = api.NetPacketType(typ)
		p.Data = buf[HeadLength:msgLen]
		packets = append(packets, p)
	}
	return packets
}
