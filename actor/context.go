/**
 * @Author: dingQingHui
 * @Description:
 * @File: actor
 * @Version: 1.0.0
 * @Date: 2023/12/8 14:19
 */

package actor

import (
	"github.com/dingqinghui/gas/api"
	"github.com/dingqinghui/gas/extend/reflectx"
	"go.uber.org/zap"
)

type baseActorContext struct {
	api.IZLogger
	actor      api.IActor
	process    api.IProcess
	system     api.IActorSystem
	mbm        *api.ActorMessage
	router     api.IActorRouter
	pid        *api.Pid
	initParams interface{}
}

var _ api.IActorContext = &baseActorContext{}
var _ api.IActorMessageInvoker = &baseActorContext{}

func NewBaseActorContext() *baseActorContext {
	return new(baseActorContext)
}

func (a *baseActorContext) InvokerMessage(msg interface{}) *api.Error {
	a.mbm = msg.(*api.ActorMessage)
	if err := a.invokerMessage(a.mbm); err != nil {
		a.Error("actor invoker message err",
			zap.String("name", reflectx.TypeFullName(a.Actor())),
			zap.String("method", a.mbm.MethodName),
			zap.Error(err))
		return err
	}
	return nil
}

func (a *baseActorContext) invokerMessage(msg *api.ActorMessage) *api.Error {
	if msg.MethodName == InitFuncName {
		return a.Actor().OnInit(a)
	}
	switch msg.Typ {
	case api.ActorInnerMessage:
		return a.invokerInnerMessage(msg)
	case api.ActorNetMessage:
		return a.invokerNetMessage(msg)
	}
	return nil
}

func (a *baseActorContext) invokerNetMessage(msg *api.ActorMessage) *api.Error {
	if a.router == nil {
		return api.ErrActorRouterIsNil
	}
	md := a.router.Get(msg.MethodName)
	if md == nil {
		return api.ErrActorNotMethod
	}
	msg.Session.Mid = msg.Mid
	method := &networkMethod{md}
	return method.call(a, msg)
}

func (a *baseActorContext) invokerInnerMessage(msg *api.ActorMessage) *api.Error {
	if a.router == nil {
		return api.ErrActorRouterIsNil
	}
	md := a.router.Get(msg.MethodName)
	if md == nil {
		return api.ErrActorNotMethod
	}
	method := &innerMethod{md}
	rsq := method.call(a, msg)
	if !api.IsOk(rsq.Err) {
		return rsq.Err
	}
	msg.Respond(rsq)
	return nil
}

func (a *baseActorContext) Message() *api.ActorMessage {
	return a.mbm
}

func (a *baseActorContext) Process() api.IProcess {
	return a.process
}

func (a *baseActorContext) System() api.IActorSystem {
	return a.system
}

func (a *baseActorContext) Actor() api.IActor {
	return a.actor
}

func (a *baseActorContext) Self() *api.Pid {
	return a.pid
}

func (a *baseActorContext) InitParams() interface{} {
	return a.initParams
}

func (a *baseActorContext) Router() api.IActorRouter {
	return a.router
}

func (a *baseActorContext) RegisterName(name string) *api.Error {
	return a.System().RegisterName(name, a.Self())
}

func (a *baseActorContext) UnregisterName(name string) (*api.Pid, *api.Error) {
	return a.System().UnregisterName(name)
}

func (a *baseActorContext) Send(to *api.Pid, funcName string, request interface{}) *api.Error {
	return a.System().Send(a.Self(), to, funcName, request)
}

func (a *baseActorContext) Call(to *api.Pid, funcName string, request, reply interface{}) *api.Error {
	return a.System().Call(a.Self(), to, funcName, request, reply)
}

func (a *baseActorContext) Response(session *api.Session, s2c interface{}) *api.Error {
	return a.Push(session, session.Mid, s2c)
}

func (a *baseActorContext) Push(session *api.Session, mid uint16, s2c interface{}) *api.Error {
	if a.Message().Typ != api.ActorNetMessage {
		return api.ErrNetworkRespond
	}
	data, err := a.System().Serializer().Marshal(s2c)
	if err != nil {
		return api.ErrMarshal
	}
	if a.System().IsLocalPid(session.Agent) {
		entity := session.GetEntity()
		if entity == nil {
			return api.ErrNetworkRespond
		}
		return entity.SendRawMessage(mid, data)
	} else {
		message := api.BuildNetMessage(session, "Push", mid, data)
		message.To = session.Agent
		message.From = a.Self()
		return a.System().PostMessage(session.Agent, message)
	}
}
