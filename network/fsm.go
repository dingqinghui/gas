/**
 * @Author: dingQingHui
 * @Description:
 * @File: FSMState
 * @Version: 1.0.0
 * @Date: 2024/11/13 17:21
 */

package network

import (
	"github.com/dingqinghui/gas/api"
	"github.com/dingqinghui/gas/zlog"
	"go.uber.org/zap"
)

type IFsmState interface {
	Exec(packet api.INetPacket) *api.Error
	Next() IFsmState
	Name() string
}

func newClosedState(entity *Entity) *closedState {
	return &closedState{&baseState{
		Entity: entity,
	}}
}

type baseState struct {
	*Entity
}

// closedState
// @Description: 初始状态
type closedState struct {
	*baseState
}

func (s *closedState) Exec(packet api.INetPacket) *api.Error {
	if s.Type() == api.NetListener && packet != nil {
		if packet.GetTyp() != PacketTypeHandshake {
			zlog.Error("net entity fsm exec", zap.Uint64("id", s.ID()),
				zap.Int("typ", int(packet.GetTyp())), zap.Error(api.ErrPacketType))
			return api.ErrPacketType
		}
		return s.serverHandshake(packet)
	} else {
		_packet := NewHandshakePacket(s.opts.HandshakeBody)
		if err := s.SendPacket(_packet); err != nil {
			return err
		}
	}
	return nil
}

func (s *closedState) serverHandshake(packet api.INetPacket) *api.Error {
	// handshake auth
	buf, err := s.opts.HandshakeAuth(s, packet.GetData())
	if !api.IsOk(err) {
		zlog.Error("server handshake auth err",
			zap.Uint64("sessionId", s.ID()), zap.Error(err))
		return err
	}
	// ack client handshake
	return s.SendPacket(NewHandshakePacket(buf))
}

func (s *closedState) Next() IFsmState {
	if s.Type() == api.NetListener {
		return &waitHandshakeAckState{baseState: s.baseState}
	} else {
		return &waitHandshakeState{baseState: s.baseState}
	}
}

func (s *closedState) Name() string {
	return "closedState"
}

// waitHandshakeState
// @Description: 客戶端等待回复握手结果
type waitHandshakeState struct {
	*baseState
}

func (w *waitHandshakeState) Exec(packet api.INetPacket) *api.Error {
	if packet == nil {
		return nil
	}
	if packet.GetTyp() != PacketTypeHandshake {
		zlog.Error("net entity fsm exec", zap.Uint64("id", w.ID()),
			zap.Int("typ", int(packet.GetTyp())), zap.Error(api.ErrPacketType))
		return api.ErrPacketType
	}
	// send handshake ack
	if err := w.SendPacket(HandshakeAckPacket); err != nil {
		return err
	}

	return w.spawnAgent()
}
func (w *waitHandshakeState) Next() IFsmState {
	return &workingState{baseState: w.baseState}
}

func (w *waitHandshakeState) Name() string {
	return "waitHandshakeState"
}

// waitHandshakeAckState
// @Description: 服务器等待握手ACK
type waitHandshakeAckState struct {
	*baseState
}

func (w *waitHandshakeAckState) Exec(packet api.INetPacket) *api.Error {
	if packet == nil {
		return nil
	}
	if packet.GetTyp() != PacketTypeHandshakeAck {
		zlog.Error("net entity fsm exec", zap.Uint64("id", w.ID()),
			zap.Int("typ", int(packet.GetTyp())), zap.Error(api.ErrPacketType))
		return api.ErrPacketType
	}
	return w.spawnAgent()
}

func (w *waitHandshakeAckState) Next() IFsmState {
	//w.heartBeat()
	//w.addHeartBeatTimer()
	return &workingState{baseState: w.baseState}
}

func (w *waitHandshakeAckState) Name() string {
	return "waitHandshakeAckState"
}

// workingState
// @Description: 开始接受数据包
type workingState struct {
	*baseState
}

func (s *workingState) Exec(packet api.INetPacket) *api.Error {
	if packet == nil {
		return nil
	}
	if packet.GetTyp() != PacketTypeData && packet.GetTyp() != PacketTypeHeartbeat {
		zlog.Error("net entity fsm exec", zap.Uint64("id", s.ID()),
			zap.Int("typ", int(packet.GetTyp())), zap.Error(api.ErrPacketType))
		return api.ErrPacketType
	}
	if packet.GetTyp() == PacketTypeHeartbeat {
		//s.heartBeat()
	}
	return s.processDataPack(packet)
}
func (s *workingState) processDataPack(packet api.INetPacket) *api.Error {
	msg := msgCodec.Decode(packet.GetData())
	router := s.opts.RouterHandler
	if router == nil {
		zlog.Error("net entity fsm exec", zap.Uint64("id", s.ID()),
			zap.Int("typ", int(packet.GetTyp())), zap.Error(api.ErrNetworkRoute))
		return api.ErrNetworkRoute
	}
	return router(s.Session(), msg)
}

func (s *workingState) Next() IFsmState {
	return s
}

func (s *workingState) Name() string {
	return "workingState"
}
