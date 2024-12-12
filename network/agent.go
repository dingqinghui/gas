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
	session api.ISession
	api.BuiltinActor
}

func (t *AgentActor) OnInit(ctx api.IActorContext) error {
	_ = t.BuiltinActor.OnInit(ctx)
	t.session = t.Ctx.InitParams().(api.ISession)
	return nil
}
func (t *AgentActor) Session() api.ISession {
	return t.session
}
