/**
 * @Author: dingQingHui
 * @Description:
 * @File: FSMState
 * @Version: 1.0.0
 * @Date: 2024/11/13 17:21
 */

package network

import (
	"errors"
	"github.com/dingqinghui/gas/api"
	"github.com/duke-git/lancet/v2/xerror"
	"time"
)

type IFsmState interface {
	Exec(packet api.INetPacket) error
	Next() IFsmState
}

type baseState struct {
	*SessionActor
}

// closedState
// @Description: 初始状态
type closedState struct {
	*baseState
}

func (s *closedState) Exec(packet api.INetPacket) error {

	if s.service.Type() == api.NetServiceListener && packet != nil {
		if packet.GetTyp() != PacketTypeHandshake {
			return xerror.New("not handshake packet %v", packet.GetTyp())
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

func (s *closedState) serverHandshake(packet api.INetPacket) error {
	// handshake auth
	buf, err := s.opts.HandshakeAuth(s.Ctx, packet.GetData())
	if err != nil {
		return err
	}
	// ack client handshake
	return s.SendPacket(NewHandshakePacket(buf))
}

func (s *closedState) Next() IFsmState {
	if s.service.Type() == api.NetServiceListener {
		return &waitHandshakeAckState{baseState: s.baseState}
	} else {
		return &waitHandshakeState{baseState: s.baseState}
	}
}

// waitHandshakeState
// @Description: 客戶端等待回复握手结果
type waitHandshakeState struct {
	*baseState
}

func (w *waitHandshakeState) Exec(packet api.INetPacket) error {
	if packet == nil {
		return nil
	}
	if packet.GetTyp() != PacketTypeHandshake {
		return xerror.New("not handshake packet %v", packet.GetTyp())
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

// waitHandshakeAckState
// @Description: 服务器等待握手ACK
type waitHandshakeAckState struct {
	*baseState
}

func (w *waitHandshakeAckState) Exec(packet api.INetPacket) error {
	if packet == nil {
		return nil
	}
	if packet.GetTyp() != PacketTypeHandshakeAck {
		return xerror.New("not handshake packet %v", packet.GetTyp())
	}
	return w.spawnAgent()
}

func (w *waitHandshakeAckState) Next() IFsmState {
	return &workingState{baseState: w.baseState}
}

// workingState
// @Description: 开始接受数据包
type workingState struct {
	*baseState
}

func (s *workingState) Exec(packet api.INetPacket) error {
	if packet == nil {
		return nil
	}
	if packet.GetTyp() != PacketTypeData && packet.GetTyp() != PacketTypeHeartbeat {
		return xerror.New("not data or heartbeat packet %v", packet.GetTyp())
	}
	pid := s.agentPid
	msg := msgCodec.Decode(packet.GetData())
	methodName := s.opts.Router.Get(msg.GetID())
	process := s.Ctx.System().Find(pid)
	if process == nil || process.Context() == nil {
		return errors.New("not agent pid")
	}

	router := process.Context().Router()
	if router == nil {
		return api.ErrActorRouterIsNil
	}
	c2s, s2c, err := router.NewArgs(methodName)
	if err != nil {
		return err
	}
	if err = s.opts.Serializer.Unmarshal(msg.GetData(), c2s); err != nil {
		return err
	}
	if s2c == nil {
		if err = s.Ctx.System().Send(nil, pid, methodName, c2s); err != nil {
			return err
		}
	} else {
		if err = s.Ctx.System().Call(nil, pid, methodName, time.Second*1, c2s, s2c); err != nil {
			return err
		}
		return s.push(msg.GetID(), s2c)
	}
	return nil
}

func (s *workingState) Next() IFsmState {
	return s
}
