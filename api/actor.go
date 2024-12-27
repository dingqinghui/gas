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

const (
	ActorNetMessage = iota
	ActorInnerMessage
	ActorBroadcastMessage
)

type (
	ActorEmptyMessage    = *struct{}
	ActorMessageType     = byte
	ActorProducer        func() IActor
	IActorMessageInvoker interface {
		InvokerMessage(message interface{}) *Error
	}
	IActorMailbox interface {
		PostMessage(msg interface{}) *Error
		RegisterHandlers(invoker IActorMessageInvoker, dispatcher IActorDispatcher)
	}
	IActorDispatcher interface {
		Schedule(f func(), recoverFun func(err interface{})) *Error
		Throughput() int
	}
	IActorContext interface {
		Name() string
		Message() *Message
		Process() IProcess
		System() IActorSystem
		Actor() IActor
		Self() *Pid
		InitParams() interface{}
		RegisterName(name string) *Error
		UnregisterName(name string) (*Pid, *Error)
		Router() IActorRouter
		Send(to *Pid, funcName string, request interface{}) *Error
		Call(to *Pid, funcName string, request, reply interface{}) *Error
		AddGroup(eventName string)
		RemoveGroup(eventName string)
		BroadcastGroup(eventName string, msg interface{}) *Error
	}

	IProcess interface {
		IStopper
		Pid() *Pid
		Context() IActorContext
		PostMessage(message *Message) *Error
		PostMessageAndWait(message *Message) (rsp *RespondMessage)
		Stop() *Error
	}

	IActor interface {
		OnInit(ctx IActorContext) *Error
		OnStop() *Error
	}

	IActorSystem interface {
		IModule
		Spawn(producer ActorProducer, params interface{}, opts ...ProcessOption) (*Pid, *Error)
		NextPid() *Pid
		Kill(pid *Pid) *Error
		AddTimer(pid *Pid, d time.Duration, funcName string) *Error
		Find(pid *Pid) IProcess
		RegisterName(name string, pid *Pid) *Error
		UnregisterName(name string) (*Pid, *Error)
		PostMessage(to *Pid, message *Message) *Error
		Send(from, to *Pid, funcName string, request interface{}) *Error
		Call(from, to *Pid, funcName string, request, reply interface{}) *Error
		Timeout() time.Duration
		SetTimeout(timeout time.Duration)
		IsLocalPid(pid *Pid) bool
		SetRouter(name string, router IActorRouter)
		GetOrSetRouter(actor IActor) IActorRouter
		Group() IGroup
	}

	IGroup interface {
		Add(name string, process IProcess)
		Remove(name string, pid *Pid)
		Broadcast(name string, from *Pid, msg interface{}) *Error
		Range(name string, f func(IProcess) bool)
	}

	IActorRouter interface {
		Get(name string) *reflectx.Method
		Set(name string, method *reflectx.Method)
	}

	BuiltinActor struct {
		Ctx IActorContext
	}
	ProcessOption       func(o *ActorProcessOptions)
	ActorProcessOptions struct {
		Dispatcher IActorDispatcher
		Mailbox    IActorMailbox
		Name       string
	}
)

func (r *BuiltinActor) OnInit(ctx IActorContext) *Error {
	r.Ctx = ctx
	return nil
}

func (r *BuiltinActor) OnStop() *Error {
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
func WithActorName(name string) ProcessOption {
	return func(b *ActorProcessOptions) {
		b.Name = name
	}
}
