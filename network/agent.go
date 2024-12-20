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
)

type AgentActor struct {
	api.BuiltinActor
	api.INetEntity
}

func (t *AgentActor) OnInit(ctx api.IActorContext) *api.Error {
	_ = t.BuiltinActor.OnInit(ctx)
	t.INetEntity = ctx.InitParams().(api.INetEntity)
	return nil
}

func (t *AgentActor) Response(session *api.Session, s2c interface{}) *api.Error {
	return t.PushMid(session.Msg.GetID(), s2c)
}

func (t *AgentActor) PushMid(mid uint16, s2c interface{}) *api.Error {
	data, err := t.Ctx.System().Serializer().Marshal(s2c)
	if err != nil {
		return api.ErrMarshal
	}
	message := api.NewNetworkMessage(mid, data)
	return t.Push(message)
}

func (t *AgentActor) Push(message *api.NetworkMessage) *api.Error {
	return t.SendMessage(message)
}
