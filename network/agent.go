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
	Session *api.Session
}

func (t *AgentActor) OnInit(ctx api.IActorContext) *api.Error {
	_ = t.BuiltinActor.OnInit(ctx)
	t.Session = ctx.InitParams().(*api.Session)
	return nil
}

// Push
// @Description: 外部发送调用
// @receiver t
// @param session
// @param data
// @return *api.Error
func (t *AgentActor) Push(session *api.Session, data []byte) *api.Error {
	if session.GetEntity() == nil {
		session, _ = SessionHub.Get(session.GetSid())
	}
	if session == nil {
		return api.ErrNetworkRespond
	}
	entity := session.GetEntity()
	if entity == nil {
		return api.ErrNetworkRespond
	}
	return entity.SendRawMessage(session.Mid, data)
}
