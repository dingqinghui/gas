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
	"github.com/duke-git/lancet/v2/xerror"
	"go.uber.org/zap"
	"time"
)

type IFsmState interface {
	Exec(packet api.INetPacket) error
	Next() IFsmState
	Name() string
}

type baseState struct {
	*Session
}

// closedState
// @Description: 初始状态
type closedState struct {
	*baseState
}

func (s *closedState) Exec(packet api.INetPacket) error {
	if s.Type() == api.NetListener && packet != nil {
		if packet.GetTyp() != PacketTypeHandshake {
			return xerror.New("wrong package type:%v ", packet.GetTyp())
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
	buf, err := s.opts.HandshakeAuth(s, packet.GetData())
	if err != nil {
		s.Log().Error("server handshake auth err",
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

func (w *waitHandshakeState) Exec(packet api.INetPacket) error {
	if packet == nil {
		return nil
	}
	if packet.GetTyp() != PacketTypeHandshake {
		return xerror.New("wrong package type:%v ", packet.GetTyp())
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

func (w *waitHandshakeAckState) Exec(packet api.INetPacket) error {
	if packet == nil {
		return nil
	}
	if packet.GetTyp() != PacketTypeHandshakeAck {
		return xerror.New("wrong package type:%v ", packet.GetTyp())
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

func (s *workingState) Exec(packet api.INetPacket) error {
	if packet == nil {
		return nil
	}
	if packet.GetTyp() != PacketTypeData && packet.GetTyp() != PacketTypeHeartbeat {
		return xerror.New("wrong package type:%v ", packet.GetTyp())
	}
	if packet.GetTyp() == PacketTypeHeartbeat {
		//s.heartBeat()
	}
	return s.processDataPack(packet)
}

func (s *workingState) processDataPack(packet api.INetPacket) error {
	pid := s.agentPid
	msg := msgCodec.Decode(packet.GetData())
	methodName := s.opts.Router.Get(msg.GetID())
	system := s.node.ActorSystem()
	process := system.Find(pid)
	if process == nil || process.Context() == nil {
		s.Log().Error("processDataPack err",
			zap.Uint64("sessionId", s.ID()),
			zap.String("pid", pid.String()),
			zap.Error(api.ErrNotLocalPid))
		return api.ErrNotLocalPid
	}

	router := process.Context().Router()
	if router == nil {
		s.Log().Error("processDataPack err",
			zap.Uint64("sessionId", s.ID()),
			zap.String("pid", pid.String()),
			zap.Error(api.ErrActorRouterIsNil))
		return api.ErrActorRouterIsNil
	}
	c2s, s2c, err := router.NewArgs(methodName)
	if err != nil {
		s.Log().Error("processDataPack err",
			zap.Uint64("sessionId", s.ID()),
			zap.String("pid", pid.String()),
			zap.String("methodName", methodName),
			zap.Error(err))
		return err
	}
	if err = s.opts.Serializer.Unmarshal(msg.GetData(), c2s); err != nil {
		s.Log().Error("processDataPack err",
			zap.Uint64("sessionId", s.ID()),
			zap.String("pid", pid.String()),
			zap.String("methodName", methodName),
			zap.Error(err))
		return err
	}
	if s2c == nil {
		if err = process.Send(nil, methodName, c2s); err != nil {
			s.Log().Error("processDataPack err",
				zap.Uint64("sessionId", s.ID()),
				zap.String("pid", pid.String()),
				zap.String("methodName", methodName),
				zap.Error(err))
			return err
		}
	} else {
		wait, callErr := process.Call(nil, methodName, time.Second*3, c2s, s2c)
		if callErr != nil {
			s.Log().Error("processDataPack err",
				zap.Uint64("sessionId", s.ID()),
				zap.String("pid", pid.String()),
				zap.String("methodName", methodName),
				zap.Error(callErr))
			return callErr
		}
		s.node.Workers().Submit(func() {
			if waitErr := wait.Wait(); waitErr != nil {
				s.Log().Error("processDataPack err",
					zap.Uint64("sessionId", s.ID()),
					zap.String("pid", pid.String()),
					zap.String("methodName", methodName),
					zap.Error(waitErr))
				return
			}
			if err = s.SendMessage(msg.GetID(), s2c); err != nil {
				s.Log().Error("processDataPack err",
					zap.Uint64("sessionId", s.ID()),
					zap.String("pid", pid.String()),
					zap.String("methodName", methodName),
					zap.Error(err))
				return
			}
		}, nil)
	}
	return nil
}

func (s *workingState) Next() IFsmState {
	return s
}

func (s *workingState) Name() string {
	return "workingState"
}
