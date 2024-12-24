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
	*api.Session
}

func (t *AgentActor) OnInit(ctx api.IActorContext) *api.Error {
	_ = t.BuiltinActor.OnInit(ctx)
	t.INetEntity = ctx.InitParams().(api.INetEntity)
	t.Session = t.INetEntity.Session()
	return nil
}

func (t *AgentActor) Push(data []byte) *api.Error {
	return t.SendPacket(NewDataPacket(data))
}
