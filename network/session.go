/**
 * @Author: dingQingHui
 * @Description:
 * @File: sessionactor
 * @Version: 1.0.0
 * @Date: 2024/12/5 15:32
 */

package network

import (
	"github.com/dingqinghui/gas/api"
	"github.com/panjf2000/gnet/v2"
)

type SessionActor struct {
	api.BuiltinActor
	opts     *Options
	rawCon   gnet.Conn
	service  api.INetServer
	fsm      IFsmState
	agentPid *api.Pid
	node     api.INode
}

func (t *SessionActor) OnInit(ctx api.IActorContext) error {
	_ = t.BuiltinActor.OnInit(ctx)
	t.fsm = &closedState{&baseState{
		SessionActor: t,
	}}
	return nil
}

func (t *SessionActor) Open(_ api.ActorEmptyMessage) error {
	if t.service.Type() == api.NetServiceConnector {
		if err := t.exec(nil); err != nil {
			return err
		}
	}
	return nil
}

func (t *SessionActor) Traffic(msg *GNetTrafficMessage) error {
	for _, packet := range msg.Packets {
		if err := t.exec(packet); err != nil {
			return err
		}
	}
	return nil
}

func (t *SessionActor) exec(packet api.INetPacket) error {
	if err := t.fsm.Exec(packet); err != nil {
		return err
	}
	t.fsm = t.fsm.Next()
	return nil
}

func (t *SessionActor) spawnAgent() error {
	pid, err := t.Ctx.System().Spawn(t.opts.AgentProducer, t.Ctx.Self())
	if err != nil {
		return nil
	}
	if pid == nil {
		return api.ErrPidIsNil
	}
	t.agentPid = pid
	return nil
}

func (t *SessionActor) SendPacket(packet *BuiltinNetworkPacket) error {
	buf := packCodec.Encode(packet)
	if err := t.rawCon.AsyncWrite(buf, nil); err != nil {
		return err
	}
	return nil
}

func (t *SessionActor) Push(pushMsg *AgentPushMessage) error {
	return t.push(pushMsg.MsgId, pushMsg.S2c)
}

func (t *SessionActor) push(msgId uint16, s2c interface{}) error {
	buf, err := t.opts.Serializer.Marshal(s2c)
	if err != nil {
		return err
	}
	respond := &BuiltinMessage{
		id:   msgId,
		data: buf,
	}
	data := msgCodec.Encode(respond)
	return t.SendPacket(NewDataPacket(data))
}

func (t *SessionActor) OnStop() error {
	return t.stop()
}

func (t *SessionActor) stop() error {
	if err := t.rawCon.Close(); err != nil {
		return err
	}
	if t.agentPid != nil {
		return t.Ctx.Send(t.agentPid, "OnSessionClose", nil)
	}
	return nil
}
