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
	"github.com/dingqinghui/gas/network/message"
)

type AgentActor struct {
	api.BuiltinActor
	api.INetEntity
	*api.Session
}

func (t *AgentActor) OnInit(ctx api.IActorContext) *api.Error {
	_ = t.BuiltinActor.OnInit(ctx)
	t.INetEntity = ctx.InitParams().(api.INetEntity)
	t.Session = t.INetEntity.Session()
	return nil
}

func (t *AgentActor) Push(message *message.Message) *api.Error {
	return t.SendMessage(message)
}

func (t *AgentActor) Close(err *api.Error) *api.Error {
	return t.INetEntity.Close(err)
}

func (t *AgentActor) Closed() *api.Error {
	return nil
}
