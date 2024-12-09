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
	"go.uber.org/zap"
	"reflect"
	"time"
)

type baseActorContext struct {
	api.IZLogger
	actor      api.IActor
	process    api.IProcess
	system     api.IActorSystem
	mbm        api.IMailBoxMessage
	router     api.IActorRouter
	pid        *api.Pid
	initParams interface{}
}

var _ api.IActorContext = &baseActorContext{}
var _ api.IActorMessageInvoker = &baseActorContext{}

func NewBaseActorContext() *baseActorContext {
	return new(baseActorContext)
}

func (a *baseActorContext) InvokerMessage(env api.IMailBoxMessage) error {
	a.System().Node().Workers().Try(func() {
		a.mbm = env
		if err := a.invokerMessage(env); err != nil {
			a.Error("actor invoker message error",
				zap.Uint64("pid", a.Self().GetUniqId()),
				zap.String("actor", reflect.TypeOf(a.actor).String()),
				zap.String("methodName", env.MethodName()),
				zap.Error(err))
			return
		}
	}, func(err interface{}) {
		a.Panic("actor invoker message panic",
			zap.Error(err.(error)), zap.Stack("stack"))
	})
	return nil
}

func (a *baseActorContext) invokerMessage(mbm api.IMailBoxMessage) error {
	a.mbm = mbm
	if a.router == nil {
		return api.ErrActorRouterIsNil
	}
	if err := a.router.Call(a, mbm); err != nil {
		return err
	}
	if mbm.Waiter() != nil {
		mbm.Waiter().Done()
	}
	return nil
}

func (a *baseActorContext) Message() api.IMailBoxMessage {
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

func (a *baseActorContext) Send(to *api.Pid, funcName string, request interface{}) error {
	return a.System().Node().Send(a.Self(), to, funcName, request)
}

func (a *baseActorContext) Call(to *api.Pid, funcName string, timeout time.Duration, request, reply interface{}) (err error) {
	return a.System().Node().Call(a.Self(), to, funcName, timeout, request, reply)
}

func (a *baseActorContext) RegisterName(name string) error {
	return a.System().RegisterName(name, a.Self())
}

func (a *baseActorContext) UnregisterName(name string) (*api.Pid, error) {
	return a.System().UnregisterName(name)
}
