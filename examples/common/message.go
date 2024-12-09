/**
 * @Author: dingQingHui
 * @Description:
 * @File: common
 * @Version: 1.0.0
 * @Date: 2024/12/2 11:05
 */

package common

import "github.com/dingqinghui/gas/extend/serializer"

type ClientMessage struct {
	Name    string
	Content string
}

type RpcRoomJoin struct {
	UserId int32
}

type HandshakeMessage struct {
	Version     string
	EncryptSeed uint64
	ServerTime  int64
}

func MarshalHandshakeMessage(m *HandshakeMessage) []byte {
	if data, err := serializer.Json.Marshal(m); err != nil {
		return nil
	} else {
		return data
	}
}

func UnmarshalHandshakeMessage(data []byte) *HandshakeMessage {
	m := new(HandshakeMessage)
	if err := serializer.Json.Unmarshal(data, m); err != nil {
		return nil
	} else {
		return m
	}
}
