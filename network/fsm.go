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
	"github.com/dingqinghui/gas/network/message"
	"github.com/dingqinghui/gas/network/packet"
	"github.com/dingqinghui/gas/zlog"
	"go.uber.org/zap"
)

type IFsmState interface {
	Exec(packet *packet.NetworkPacket) *api.Error
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

func (s *closedState) Exec(pkt *packet.NetworkPacket) *api.Error {
	if s.Type() == api.NetListener && pkt != nil {
		if pkt.GetTyp() != packet.HandshakeType {
			zlog.Error("net entity fsm exec", zap.Uint64("id", s.ID()),
				zap.Int("typ", int(pkt.GetTyp())), zap.Error(api.ErrPacketType))
			return api.ErrPacketType
		}
		return s.serverHandshake(pkt)
	} else {
		if err := s.SendRaw(packet.HandshakeType, s.opts.HandshakeBody); err != nil {
			return err
		}
	}
	return nil
}

func (s *closedState) serverHandshake(pkt *packet.NetworkPacket) *api.Error {
	// handshake auth
	buf, err := s.opts.HandshakeAuth(s, pkt.GetData())
	if !api.IsOk(err) {
		zlog.Error("server handshake auth err",
			zap.Uint64("sessionId", s.ID()), zap.Error(err))
		return err
	}
	// ack client handshake
	return s.SendRaw(packet.HandshakeType, buf)
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

func (w *waitHandshakeState) Exec(pkt *packet.NetworkPacket) *api.Error {
	if pkt == nil {
		return nil
	}
	if pkt.GetTyp() != packet.HandshakeType {
		zlog.Error("net entity fsm exec", zap.Uint64("id", w.ID()),
			zap.Int("typ", int(pkt.GetTyp())), zap.Error(api.ErrPacketType))
		return api.ErrPacketType
	}
	// send handshake ack
	if err := w.SendRaw(packet.HandshakeAckType, nil); err != nil {
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

func (w *waitHandshakeAckState) Exec(pkt *packet.NetworkPacket) *api.Error {
	if pkt == nil {
		return nil
	}
	if pkt.GetTyp() != packet.HandshakeAckType {
		zlog.Error("net entity fsm exec", zap.Uint64("id", w.ID()),
			zap.Int("typ", int(pkt.GetTyp())), zap.Error(api.ErrPacketType))
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

func (s *workingState) Exec(pkt *packet.NetworkPacket) *api.Error {
	if pkt == nil {
		return nil
	}
	if pkt.GetTyp() != packet.DataType && pkt.GetTyp() != packet.HeartbeatType {
		zlog.Error("net entity fsm exec", zap.Uint64("id", s.ID()),
			zap.Int("typ", int(pkt.GetTyp())), zap.Error(api.ErrPacketType))
		return api.ErrPacketType
	}
	if pkt.GetTyp() == packet.HeartbeatType {
		//s.heartBeat()
	}
	return s.processDataPack(pkt)
}
func (s *workingState) processDataPack(packet *packet.NetworkPacket) *api.Error {
	msg := message.Decode(packet.GetData())

	session := s.session.Dup()
	session.Message = msg

	routeFunc := s.opts.RouterHandler
	if routeFunc == nil {
		return api.ErrNetworkRoute
	}
	to, method, err := routeFunc(s.Session(), msg)
	if err != nil {
		return err
	}

	actMessage := &api.Message{
		Typ:        api.ActorNetMessage,
		MethodName: method,
		Session:    session,
	}
	return api.GetNode().System().PostMessage(to, actMessage)
}

func (s *workingState) Next() IFsmState {
	return s
}

func (s *workingState) Name() string {
	return "workingState"
}
