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
	ActorEmptyMessage    = *struct{}
	ActorProducer        func() IActor
	IActorMessageInvoker interface {
		InvokerMessage(message interface{}) *Error
	}
	IActorMailbox interface {
		PostMessage(msg interface{}) *Error
		RegisterHandlers(invoker IActorMessageInvoker, dispatcher IActorDispatcher)
	}
	IActorDispatcher interface {
		Schedule(node INode, f func(), recoverFun func(err interface{})) *Error
		Throughput() int
	}
	IActorContext interface {
		IZLogger
		Message() *ActorMessage
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
		Response(session *Session, s2c interface{}) *Error
		Push(session *Session, mid uint16, s2c interface{}) *Error
	}

	IProcess interface {
		IStopper
		Pid() *Pid
		Context() IActorContext
		PostMessage(message *ActorMessage) *Error
		PostMessageAndWait(message *ActorMessage) (rsp *RespondMessage)
		Stop() *Error
	}

	IActor interface {
		OnInit(ctx IActorContext) *Error
		OnStop() *Error
	}

	IActorSystem interface {
		IModule

		Spawn(producer ActorProducer, params interface{}, opts ...ProcessOption) (*Pid, *Error)
		SpawnWithName(name string, producer ActorProducer, params interface{}, opts ...ProcessOption) (*Pid, *Error)
		NextPid() *Pid
		Kill(pid *Pid) *Error
		RouterHub() IActorRouterHub
		AddTimer(pid *Pid, d time.Duration, funcName string) *Error
		Find(pid *Pid) IProcess
		Timeout() time.Duration
		RegisterName(name string, pid *Pid) *Error
		UnregisterName(name string) (*Pid, *Error)
		PostMessage(to *Pid, message *ActorMessage) *Error
		Send(from, to *Pid, funcName string, request interface{}) *Error
		Call(from, to *Pid, funcName string, request, reply interface{}) *Error
		SetTimeout(timeout time.Duration)
		SetSerializer(serializer ISerializer)
		Serializer() ISerializer
		IsLocalPid(pid *Pid) bool
	}

	IActorRouterHub interface {
		Set(name string, router IActorRouter)
		GetOrSet(actor IActor) IActorRouter
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

type RespondMessage struct {
	Data []byte
	Err  *Error
}

type RespondFun func(rsp *RespondMessage) *Error

type ActorMessageType = byte

const (
	ActorNetMessage = iota
	ActorInnerMessage
)

type ActorBaseMessage struct {
}

type ActorMessage struct {
	Typ        int
	MethodName string
	From       *Pid
	To         *Pid
	Data       []byte
	Mid        uint16
	Session    *Session
	respond    RespondFun
}

func (m *ActorMessage) Respond(rsp *RespondMessage) {
	if m.respond == nil {
		return
	}
	m.respond(rsp)
}
func (m *ActorMessage) SetRespond(respond RespondFun) {
	m.respond = respond
}

func BuildNetMessage(session *Session, methodName string, mid uint16, data []byte) *ActorMessage {
	return &ActorMessage{
		Typ:        ActorNetMessage,
		MethodName: methodName,
		Data:       data,
		Mid:        mid,
		Session:    session,
	}
}
func BuildInnerMessage(from, to *Pid, methodName string, data []byte) *ActorMessage {
	return &ActorMessage{
		From:       from,
		To:         to,
		Typ:        ActorInnerMessage,
		MethodName: methodName,
		Data:       data,
	}
}
