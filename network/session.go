/**
 * @Author: dingQingHui
 * @Description:
 * @File: Session
 * @Version: 1.0.0
 * @Date: 2024/12/5 15:32
 */

package network

import (
	"errors"
	"github.com/dingqinghui/gas/api"
	xerror2 "github.com/dingqinghui/gas/extend/xerror"
	"github.com/duke-git/lancet/v2/xerror"
	"github.com/panjf2000/gnet/v2"
	"go.uber.org/zap"
	"sync/atomic"
)

var autoId atomic.Uint64

func newSession(server api.INetServer, meta *Meta, rawCon gnet.Conn) *Session {
	network := rawCon.RemoteAddr().Network()
	session := &Session{
		id:         autoId.Add(1),
		server:     server,
		Meta:       meta,
		rawCon:     rawCon,
		network:    network,
		remoteAddr: rawCon.RemoteAddr().String(),
		localAddr:  rawCon.LocalAddr().String(),
	}
	session.fsm = &closedState{&baseState{
		Session: session,
	}}
	if session.Type() == api.NetConnector {
		xerror2.Assert(session.exec(nil))
	}
	server.Link(session, rawCon)
	session.Log().Info("new session",
		zap.Uint64("sessionId", session.ID()),
		zap.Int("typ", int(session.Type())),
		zap.String("network", session.Network()),
		zap.String("remoteAddr", session.RemoteAddr()),
		zap.String("localAddr", session.LocalAddr()))
	return session
}

type Session struct {
	api.BuiltinStopper
	*Meta
	id                uint64
	server            api.INetServer
	rawCon            gnet.Conn
	fsm               IFsmState
	agentPid          *api.Pid
	network           string
	localAddr         string
	remoteAddr        string
	lastHeartBeatTime int64
}

func (s *Session) ID() uint64 {
	return s.id
}

func (s *Session) Log() api.IZLogger {
	return s.node.Log()
}

func (s *Session) Traffic(c gnet.Conn) error {
	packets := packCodec.Decode(c)
	for _, packet := range packets {
		if err := s.exec(packet); err != nil {
			return err
		}
	}
	return nil
}

func (s *Session) exec(packet api.INetPacket) error {
	if err := s.fsm.Exec(packet); err != nil {
		s.Log().Error("session exec err",
			zap.Uint64("sessionId", s.ID()), zap.Error(err))
		return err
	}
	fsm := s.fsm.Next()
	if fsm != s.fsm {
		s.fsm = fsm
		s.Log().Info("session fsm change state",
			zap.Uint64("sessionId", s.ID()), zap.String("fsmName", fsm.Name()))
	}
	return nil
}

func (s *Session) spawnAgent() error {
	pid, err := s.node.ActorSystem().Spawn(s.opts.AgentProducer, s)
	if err != nil {
		s.Log().Error("session spawn agent err",
			zap.Uint64("sessionId", s.ID()), zap.Error(err))
		return nil
	}
	if pid == nil {
		s.Log().Error("session spawn agent err",
			zap.Uint64("sessionId", s.ID()), zap.Error(api.ErrPidIsNil))
		return api.ErrPidIsNil
	}
	s.agentPid = pid
	s.Log().Info("session spawn agent",
		zap.Uint64("sessionId", s.ID()), zap.String("agent", pid.String()))
	return nil
}

func (s *Session) SendPacket(packet *api.BuiltinNetworkPacket) error {
	buf := packCodec.Encode(packet)
	switch s.Network() {
	case "tcp":
		if err := s.rawCon.AsyncWrite(buf, nil); err != nil {
			s.Log().Error("session send packet err",
				zap.Uint64("sessionId", s.ID()), zap.Error(err))
			return err
		}
	case "udp":
		_, err := s.rawCon.Write(buf)
		if err != nil {
			s.Log().Error("session send packet err",
				zap.Uint64("sessionId", s.ID()), zap.Error(err))
		}
	default:
		return xerror.New("unknown network:v", s.Network())
	}
	if packet.GetTyp() == PacketTypeKick {
		s.Log().Info("session send kick packet",
			zap.Uint64("sessionId", s.ID()))
		return s.Close(errors.New("kick"))
	}
	s.Log().Debug("session send packet",
		zap.Uint64("sessionId", s.ID()), zap.Any("packet", packet))
	return nil
}

func (s *Session) SendMessage(msgId uint16, s2c interface{}) error {
	buf, err := s.opts.Serializer.Marshal(s2c)
	if err != nil {
		s.Log().Error("session send message err",
			zap.Uint64("sessionId", s.ID()), zap.Error(err))
		return err
	}
	respond := &BuiltinMessage{
		id:   msgId,
		data: buf,
	}
	data := msgCodec.Encode(respond)
	return s.SendPacket(NewDataPacket(data))
}

func (s *Session) Type() api.NetConnectionType {
	return s.typ
}
func (s *Session) Network() string {
	return s.network
}

func (s *Session) Close(reason error) error {
	if err := s.BuiltinStopper.Stop(); err != nil {
		return err
	}
	s.server.Unlink(s.rawCon)
	switch s.Network() {
	case "udp":
		return s.Closed(reason)
	case "tcp":
		return s.rawCon.Close()
	}
	return nil
}

func (s *Session) Closed(err error) error {
	if s.agentPid != nil {
		return s.node.Send(nil, s.agentPid, "Closed", nil)
	}
	s.node.Log().Info("session close", zap.Uint64("id", s.ID()), zap.Error(err))
	return nil
}
func (s *Session) Node() api.INode {
	return s.node
}

func (s *Session) LocalAddr() string {
	return s.localAddr
}
func (s *Session) RemoteAddr() string {
	return s.remoteAddr
}

func (s *Session) RawCon() gnet.Conn {
	return s.rawCon
}
