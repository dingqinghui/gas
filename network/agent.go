/**
 * @Author: dingQingHui
 * @Description:
 * @File: actor
 * @Version: 1.0.0
 * @Date: 2024/12/5 17:23
 */

package network

import (
	"github.com/dingqinghui/gas/api"
	"go.uber.org/zap"
)

type AgentActor struct {
	sessionPid *api.Pid
	api.BuiltinActor
}

func (t *AgentActor) OnInit(ctx api.IActorContext) error {
	_ = t.BuiltinActor.OnInit(ctx)
	t.sessionPid = t.Ctx.InitParams().(*api.Pid)
	t.Ctx.Info("builtin agent actor init", zap.Uint64("pid", t.sessionPid.GetUniqId()))
	return nil
}

func (t *AgentActor) Push(msgId uint16, s2c interface{}) error {
	return t.Ctx.Send(t.sessionPid, "Push", &AgentPushMessage{MsgId: msgId, S2c: s2c})
}

func (t *AgentActor) KillSession(_ api.ActorEmptyMessage) error {
	return t.Ctx.System().Kill(t.sessionPid)
}

func (t *AgentActor) OnSessionClose(_ api.ActorEmptyMessage) error {
	return nil
}
