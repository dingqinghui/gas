/**
 * @Author: dingQingHui
 * @Description:
 * @File: actor
 * @Version: 1.0.0
 * @Date: 2024/11/25 16:36
 */

package api

import (
	"github.com/dingqinghui/gas/extend/reflectx"
	"time"
)

type (
	ActorEmptyMessage = *struct{}

	IActorWaiter interface {
		Wait() error
		Done()
	}

	ActorProducer        func() IActor
	IActorMessageInvoker interface {
		InvokerMessage(message IMailBoxMessage) error
	}
	IActorMailbox interface {
		PostMessage(msg IMailBoxMessage) error
		RegisterHandlers(invoker IActorMessageInvoker, dispatcher IActorDispatcher)
	}
	IActorDispatcher interface {
		Schedule(node INode, f func(), recoverFun func(err interface{})) error
		Throughput() int
	}

	IMailBoxMessage interface {
		MethodName() string
		Args() []interface{}
		Waiter() IActorWaiter
		From() *Pid
	}

	IActorContext interface {
		IZLogger
		Message() IMailBoxMessage
		Process() IProcess
		System() IActorSystem
		Actor() IActor
		Self() *Pid
		InitParams() interface{}
		RegisterName(name string) error
		UnregisterName(name string) (*Pid, error)
		Router() IActorRouter
		Send(to *Pid, funcName string, request interface{}) error
		Call(to *Pid, funcName string, timeout time.Duration, request, reply interface{}) error
	}

	IProcess interface {
		IStopper
		Pid() *Pid
		Context() IActorContext
		Send(from *Pid, funcName string, request interface{}) error
		CallAndWait(from *Pid, funcName string, timeout time.Duration, request, reply interface{}) error
		Call(from *Pid, methodName string, timeout time.Duration, request, reply any) (IActorWaiter, error)
		AsyncStop() error
	}

	IActor interface {
		OnInit(ctx IActorContext) error
		OnStop() error
	}

	IActorSystem interface {
		IModule
		Send(from, to *Pid, funcName string, request interface{}) error
		Call(from, to *Pid, funcName string, timeout time.Duration, request, reply interface{}) error
		Spawn(producer ActorProducer, params interface{}, opts ...ProcessOption) (*Pid, error)
		SpawnWithName(name string, producer ActorProducer, params interface{}, opts ...ProcessOption) (*Pid, error)
		RegisterName(name string, pid *Pid) error
		UnregisterName(name string) (*Pid, error)
		NextPid() *Pid
		Kill(pid *Pid) error
		RouterHub() IActorRouterHub
		AddTimer(pid *Pid, d time.Duration, funcName string) error
		Find(pid *Pid) IProcess
	}

	IActorSender interface {
		Send(from, to *Pid, funcName string, request interface{}) error
		Call(from, to *Pid, funcName string, timeout time.Duration, request, reply interface{}) error
	}

	IActorRouterHub interface {
		Set(name string, router IActorRouter)
		GetOrSet(actor IActor) IActorRouter
	}
	IActorRouter interface {
		Get(name string) *reflectx.Method
		Call(ctx IActorContext, env IMailBoxMessage) error
		NewArgs(methodName string) (request, reply interface{}, err error)
	}

	BuiltinActor struct {
		Ctx IActorContext
	}
	ProcessOption       func(o *ActorProcessOptions)
	ActorProcessOptions struct {
		Dispatcher IActorDispatcher
		Mailbox    IActorMailbox
	}
)

func (r *BuiltinActor) OnInit(ctx IActorContext) error {
	r.Ctx = ctx
	return nil
}
func (r *BuiltinActor) OnStop() error {
	return nil
}

func WithActorDispatcher(dispatcher IActorDispatcher) ProcessOption {
	return func(b *ActorProcessOptions) {
		b.Dispatcher = dispatcher
	}
}
func WithActorMailBox(mailbox IActorMailbox) ProcessOption {
	return func(b *ActorProcessOptions) {
		b.Mailbox = mailbox
	}
}
