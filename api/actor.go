/**
 * @Author: dingQingHui
 * @Description:
 * @File: actor
 * @Version: 1.0.0
 * @Date: 2024/11/25 16:36
 */

package api

import (
	"fmt"
	"github.com/dingqinghui/gas/extend/reflectx"
	"time"
)

const (
	ActorNetMessage = iota
	ActorInnerMessage
)

type (
	ActorEmptyMessage = *struct{}

	RespondFun func(rsp *RespondMessage) *Error

	ActorMessageType = byte

	ActorMessage struct {
		Typ        int
		MethodName string
		From       *Pid
		To         *Pid
		Data       []byte
		Mid        uint16
		Session    *Session
		respond    RespondFun
	}
	RespondMessage struct {
		Data []byte
		Err  *Error
	}

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
		Name() string
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
		AddGroup(eventName string)
		RemoveGroup(eventName string)
		BroadcastGroup(eventName string, msg interface{}) *Error
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
		NextPid() *Pid
		Kill(pid *Pid) *Error
		AddTimer(pid *Pid, d time.Duration, funcName string) *Error
		Find(pid *Pid) IProcess
		RegisterName(name string, pid *Pid) *Error
		UnregisterName(name string) (*Pid, *Error)
		PostMessage(to *Pid, message *ActorMessage) *Error
		Send(from, to *Pid, funcName string, request interface{}) *Error
		Call(from, to *Pid, funcName string, request, reply interface{}) *Error
		Timeout() time.Duration
		SetTimeout(timeout time.Duration)
		Serializer() ISerializer
		SetSerializer(serializer ISerializer)
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

	Pid struct {
		NodeId uint64
		UniqId uint64
		Name   string
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

func (m *ActorMessage) Respond(rsp *RespondMessage) *Error {
	if m.respond == nil {
		return nil
	}
	return m.respond(rsp)
}

func (m *ActorMessage) SetRespond(respond RespondFun) {
	m.respond = respond
}

func BuildNetMessage(session *Session, methodName string, msg *NetworkMessage) *ActorMessage {
	return &ActorMessage{
		Typ:        ActorNetMessage,
		MethodName: methodName,
		Session:    session,
		Mid:        msg.GetID(),
		Data:       msg.Data,
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

func NewRemotePid(nodeId uint64, service string) *Pid {
	return &Pid{
		NodeId: nodeId,
		Name:   service,
	}
}

func ValidPid(pid *Pid) bool {
	if pid == nil {
		return false
	}
	if pid.GetUniqId() > 0 {
		return true
	}
	if pid.GetName() != "" {
		return true
	}
	return false
}
func (p *Pid) GetNodeId() uint64 {
	return p.NodeId
}
func (p *Pid) GetUniqId() uint64 {
	return p.UniqId
}
func (p *Pid) String() string {
	if p == nil {
		return ""
	}
	return fmt.Sprintf("%v.%v", p.NodeId, p.UniqId)
}
func (p *Pid) GetName() string {
	return p.Name
}
