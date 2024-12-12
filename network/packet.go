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
	HandshakeAckPacket = NewPacket(PacketTypeHandshakeAck, nil)
)

func NewHandshakePacket(buf []byte) *api.BuiltinNetworkPacket {
	return NewPacket(PacketTypeHandshake, buf)
}

func NewDataPacket(data []byte) *api.BuiltinNetworkPacket {
	return &api.BuiltinNetworkPacket{Type: PacketTypeData, Data: data}
}

func NewPacket(tye api.NetPacketType, data []byte) *api.BuiltinNetworkPacket {
	return &api.BuiltinNetworkPacket{Type: tye, Data: data}
}

const (
	HeadLength    = 2 + 4
	MaxPacketSize = 64 * 1024
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

func (codec *BuiltinPacketCodec) Decode(reader gnet.Reader) []*api.BuiltinNetworkPacket {
	if reader == nil {
		return nil
	}
	var packets []*api.BuiltinNetworkPacket
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
		p := new(api.BuiltinNetworkPacket)
		p.Type = api.NetPacketType(typ)
		p.Data = buf[HeadLength:msgLen]
		packets = append(packets, p)
	}
	return packets
}
