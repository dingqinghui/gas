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
	"github.com/dingqinghui/gas/zlog"
	"go.uber.org/zap"
)

type baseActorContext struct {
	name       string
	actor      api.IActor
	process    api.IProcess
	system     api.IActorSystem
	mbm        *api.Message
	router     api.IActorRouter
	pid        *api.Pid
	initParams interface{}
	groups     map[string]struct{}
}

var _ api.IActorContext = &baseActorContext{}
var _ api.IActorMessageInvoker = &baseActorContext{}

func NewBaseActorContext() *baseActorContext {
	ctx := new(baseActorContext)
	ctx.groups = make(map[string]struct{})
	return ctx
}

func (a *baseActorContext) Name() string {
	return a.name
}

func (a *baseActorContext) InvokerMessage(msg interface{}) *api.Error {
	a.mbm = msg.(*api.Message)
	if err := a.invokerMessage(a.mbm); err != nil {
		zlog.Error("actor处理消息失败",
			zap.String("name", reflectx.TypeFullName(a.Actor())),
			zap.String("method", a.mbm.Method),
			zap.Error(err))
		return err
	}
	return nil
}

func (a *baseActorContext) invokerMessage(msg *api.Message) *api.Error {

	switch msg.Method {
	case api.InitFuncName:
		return a.Actor().OnInit(a)
	case api.StopFuncName:
		return a.OnStop()
	}

	switch msg.Typ {
	case api.MessageEnumInner:
		return a.invokerInnerMessage(msg)
	case api.MessageEnumNetwork:
		return a.invokerNetMessage(msg)
	}
	return nil
}

func (a *baseActorContext) invokerNetMessage(msg *api.Message) *api.Error {
	if a.router == nil {
		return api.ErrActorRouterIsNil
	}
	md := a.router.Get(msg.Method)
	if md == nil {
		return api.ErrActorNotMethod
	}
	msg.Session.SetContext(a)
	method := &networkMethod{md}
	return method.call(a, msg)
}

func (a *baseActorContext) invokerInnerMessage(msg *api.Message) *api.Error {
	if a.router == nil {
		return api.ErrActorRouterIsNil
	}
	md := a.router.Get(msg.Method)
	if md == nil {
		return api.ErrActorNotMethod
	}
	method := &innerMethod{md}
	rsq := method.call(a, msg)
	if !api.IsOk(rsq.Err) {
		return rsq.Err
	}
	return msg.Respond(rsq)
}

func (a *baseActorContext) Message() *api.Message {
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
	a.name = name
	return a.System().RegisterName(name, a.Self())
}

func (a *baseActorContext) UnregisterName(name string) (*api.Pid, *api.Error) {
	a.name = ""
	return a.System().UnregisterName(name)
}

func (a *baseActorContext) Send(to *api.Pid, funcName string, request interface{}) *api.Error {
	return a.System().Send(a.Self(), to, funcName, request)
}

func (a *baseActorContext) Call(to *api.Pid, funcName string, request, reply interface{}) *api.Error {
	return a.System().Call(a.Self(), to, funcName, request, reply)
}

//func (a *baseActorContext) Response(session *api.Session, s2c interface{}) *api.Error {
//	return a.Push(session, session.Mid, s2c)
//}
//
//func (a *baseActorContext) Push(session *api.Session, mid uint16, s2c interface{}) *api.Error {
//	if a.Message().Typ != api.ActorNetMessage {
//		return api.ErrNetworkRespond
//	}
//	data, err := a.System().Serializer().Marshal(s2c)
//	if err != nil {
//		return api.ErrMarshal
//	}
//	if a.System().IsLocalPid(session.Agent) {
//		entity := session.GetEntity()
//		if entity == nil {
//			return api.ErrNetworkRespond
//		}
//		message := api.NewNetworkMessage(mid, data)
//		return entity.SendMessage(message)
//	} else {
//		netMessage := api.NewNetworkMessage(mid, data)
//		mData, wrong := a.System().Serializer().Marshal(netMessage)
//		if wrong != nil {
//			return api.ErrMarshal
//		}
//		message := api.BuildInnerMessage(a.Self(), session.Agent, "Push", mData)
//		return a.System().PostMessage(session.Agent, message)
//	}
//}

func (a *baseActorContext) AddGroup(name string) {
	a.System().Group().Add(name, a.Process())
	a.groups[name] = struct{}{}
}

func (a *baseActorContext) RemoveGroup(name string) {
	a.System().Group().Remove(name, a.Self())
	delete(a.groups, name)
}

func (a *baseActorContext) BroadcastGroup(name string, msg interface{}) *api.Error {
	return a.System().Group().Broadcast(name, a.Self(), msg)
}

func (a *baseActorContext) OnStop() *api.Error {
	for event, _ := range a.groups {
		a.RemoveGroup(event)
	}
	if a.name != "" {
		_, _ = a.System().UnregisterName(a.name)
	}
	return a.Actor().OnStop()
}
