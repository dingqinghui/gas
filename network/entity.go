/**
 * @Author: dingQingHui
 * @Description:
 * @File: Entity
 * @Version: 1.0.0
 * @Date: 2024/12/5 15:32
 */

package network

import (
	"errors"
	"github.com/dingqinghui/gas/api"
	xerror2 "github.com/dingqinghui/gas/extend/xerror"
	"github.com/dingqinghui/gas/zlog"
	"github.com/panjf2000/gnet/v2"
	"go.uber.org/zap"
	"sync/atomic"
)

var autoId atomic.Uint64

func newEntity(server api.INetServer, opts *Options, rawCon gnet.Conn) *Entity {
	entity := &Entity{
		id:     autoId.Add(1),
		server: server,
		rawCon: rawCon,
		node:   server.Node(),
		typ:    server.Typ(),
		opts:   opts,
	}

	entity.fsm = newClosedState(entity)

	if rawCon != nil && rawCon.RemoteAddr() != nil {
		entity.network = rawCon.RemoteAddr().Network()
		entity.remoteAddr = rawCon.RemoteAddr().String()
	}
	if rawCon != nil && rawCon.LocalAddr() != nil {
		entity.localAddr = rawCon.LocalAddr().String()
	}

	if entity.Type() == api.NetConnector {
		xerror2.Assert(entity.exec(nil))
	}

	zlog.Info("new entity",
		zap.Uint64("entityId", entity.ID()),
		zap.Int("typ", int(entity.Type())),
		zap.String("network", entity.Network()),
		zap.String("remoteAddr", entity.RemoteAddr()),
		zap.String("localAddr", entity.LocalAddr()))
	return entity
}

type Entity struct {
	api.BuiltinStopper
	id                uint64
	server            api.INetServer
	rawCon            gnet.Conn
	fsm               IFsmState
	agentPid          *api.Pid
	node              api.INode
	opts              *Options
	typ               api.NetEntityType
	network           string
	localAddr         string
	remoteAddr        string
	lastHeartBeatTime int64
	session           *api.Session
}

func (s *Entity) ID() uint64 {
	return s.id
}

func (s *Entity) Traffic(c gnet.Conn) error {
	packets := packCodec.Decode(c)
	for _, packet := range packets {
		if err := s.exec(packet); err != nil {
			return err
		}
	}
	return nil
}

func (s *Entity) exec(packet api.INetPacket) error {
	if err := s.fsm.Exec(packet); err != nil {
		zlog.Error("entity exec err",
			zap.Uint64("entityId", s.ID()), zap.Error(err))
		return err
	}
	fsm := s.fsm.Next()
	if fsm != s.fsm {
		s.fsm = fsm
		zlog.Info("entity fsm change state",
			zap.Uint64("entityId", s.ID()),
			zap.String("fsmName", fsm.Name()))
	}

	zlog.Debug("entity exec packet",
		zap.Uint64("entityId", s.ID()),
		zap.Any("packet", packet))
	return nil
}

func (s *Entity) spawnAgent() *api.Error {
	s.session = api.NewSession(s)
	pid, err := s.node.System().Spawn(s.opts.AgentProducer, s)
	if err != nil {
		zlog.Error("entity spawn agent err",
			zap.Uint64("entityId", s.ID()), zap.Error(err))
		return err
	}

	if pid == nil {
		zlog.Error("entity spawn agent err",
			zap.Uint64("entityId", s.ID()), zap.Error(api.ErrPidIsNil))
		return api.ErrPidIsNil
	}
	s.agentPid = pid
	zlog.Info("entity spawn agent",
		zap.Uint64("entityId", s.ID()), zap.String("agent", pid.String()))
	return nil
}

func (s *Entity) SendPacket(packet *api.NetworkPacket) *api.Error {
	buf := packCodec.Encode(packet)
	switch s.Network() {
	case "tcp":
		if err := s.rawCon.AsyncWrite(buf, nil); err != nil {
			zlog.Error("entity send packet err",
				zap.Uint64("entityId", s.ID()), zap.Error(err))
			return api.ErrGNetRaw
		}
	case "udp":
		_, err := s.rawCon.Write(buf)
		if err != nil {
			zlog.Error("entity send packet err",
				zap.Uint64("entityId", s.ID()), zap.Error(err))
			return api.ErrGNetRaw
		}
	}
	if packet.GetTyp() == PacketTypeKick {
		zlog.Info("entity send kick packet",
			zap.Uint64("entityId", s.ID()))
		return s.Close(errors.New("kick"))
	}
	zlog.Debug("entity send packet",
		zap.Uint64("entityId", s.ID()), zap.Any("packet", packet))
	return nil
}

func (s *Entity) SendMessage(message *api.NetworkMessage) *api.Error {
	if message == nil {
		return nil
	}
	data := msgCodec.Encode(message)
	return s.SendPacket(NewDataPacket(data))
}

func (s *Entity) Type() api.NetEntityType {
	return s.typ
}
func (s *Entity) Network() string {
	return s.network
}

func (s *Entity) Close(reason error) *api.Error {
	if err := s.BuiltinStopper.Stop(); err != nil {
		zlog.Error("entity close", zap.Uint64("id", s.ID()), zap.Error(err))
		return err
	}
	s.server.Unlink(s.rawCon)
	switch s.Network() {
	case "udp":
		return s.Closed(reason)
	case "tcp":
		if err := s.rawCon.Close(); err != nil {
			zlog.Error("entity close", zap.Uint64("id", s.ID()), zap.Error(err))
			return api.ErrGNetRaw
		}
	}
	return nil
}

func (s *Entity) Closed(err error) *api.Error {
	if s.agentPid != nil {
		if wrong := s.node.System().Send(nil, s.agentPid, "Closed", nil); wrong != nil {
			zlog.Error("entity closed", zap.Uint64("id", s.ID()), zap.Error(err))
			return wrong
		}
	}
	zlog.Info("entity closed", zap.Uint64("id", s.ID()), zap.Error(err))
	return nil
}
func (s *Entity) Node() api.INode {
	return s.node
}

func (s *Entity) LocalAddr() string {
	return s.localAddr
}
func (s *Entity) RemoteAddr() string {
	return s.remoteAddr
}

func (s *Entity) RawCon() gnet.Conn {
	return s.rawCon
}
func (s *Entity) GetAgent() *api.Pid {
	return s.agentPid
}

func (s *Entity) Session() *api.Session {
	return s.session
}
