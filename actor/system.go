/**
 * @Author: dingQingHui
 * @Description:
 * @File: system
 * @Version: 1.0.0
 * @Date: 2023/12/7 14:54
 */

package actor

import (
	"github.com/dingqinghui/gas/api"
	"github.com/dingqinghui/gas/extend/asynctime"
	"github.com/dingqinghui/gas/extend/reflectx"
	"github.com/dingqinghui/gas/extend/serializer"
	"github.com/dingqinghui/gas/zlog"
	"github.com/duke-git/lancet/v2/maputil"
	"go.uber.org/zap"
	"sync/atomic"
	"time"
)

type System struct {
	api.BuiltinModule
	uniqId      atomic.Uint64
	nameDict    *maputil.ConcurrentMap[string, *api.Pid]
	processDict *maputil.ConcurrentMap[uint64, api.IProcess]
	timeout     time.Duration
	serializer  api.ISerializer
	routerDict  *maputil.ConcurrentMap[string, api.IActorRouter]
	group       *Group
}

func NewSystem(node api.INode) api.IActorSystem {
	s := new(System)
	s.SetNode(node)
	s.Init()
	node.AddModule(s)
	return s
}

func (s *System) Init() {
	s.nameDict = maputil.NewConcurrentMap[string, *api.Pid](10)
	s.processDict = maputil.NewConcurrentMap[uint64, api.IProcess](10)
	s.routerDict = maputil.NewConcurrentMap[string, api.IActorRouter](10)
	s.timeout = time.Second * 1
	s.serializer = serializer.Json
	s.group = NewGroup(s)
}

func (s *System) Name() string {
	return "actorSystem"
}

func (s *System) SetRouter(name string, router api.IActorRouter) {
	s.routerDict.Set(name, router)
}

func (s *System) GetOrSetRouter(actor api.IActor) api.IActorRouter {
	name := reflectx.TypeFullName(actor)
	v, ok := s.routerDict.Get(name)
	if !ok {
		v, ok = s.routerDict.GetOrSet(name, NewRouter(name, actor))
	}
	return v
}

func (s *System) Timeout() time.Duration {
	return s.timeout
}
func (s *System) SetTimeout(timeout time.Duration) {
	s.timeout = timeout
}
func (s *System) SetSerializer(serializer api.ISerializer) {
	s.serializer = serializer
}
func (s *System) Serializer() api.ISerializer {
	return s.serializer
}

func (s *System) Find(pid *api.Pid) api.IProcess {
	if pid == nil {
		return nil
	}
	if pid.GetNodeId() != s.Node().GetID() {
		return nil
	}
	if pid.GetUniqId() > 0 {
		return s.FindById(pid.GetUniqId())
	}
	if pid.GetName() != "" {
		return s.FindByName(pid.GetName())
	}
	return nil
}

func (s *System) FindById(uniqId uint64) api.IProcess {
	v, _ := s.processDict.Get(uniqId)
	return v
}

func (s *System) FindByName(name string) api.IProcess {
	pid, _ := s.nameDict.Get(name)
	if pid == nil {
		return nil
	}
	return s.FindById(pid.GetUniqId())
}

func (s *System) RegisterName(name string, pid *api.Pid) *api.Error {
	if name == "" {
		return nil
	}
	if pid == nil {
		return api.ErrPidIsNil
	}
	_, ok := s.nameDict.GetOrSet(name, pid)
	if ok {
		return api.ErrActorNameExist
	}
	pid.Name = name
	return nil
}

func (s *System) UnregisterName(name string) (*api.Pid, *api.Error) {
	v, ok := s.nameDict.GetAndDelete(name)
	if !ok {
		return nil, api.ErrActorNameNotExist
	}
	return v, nil
}

func (s *System) Spawn(producer api.ActorProducer, params interface{}, opts ...api.ProcessOption) (*api.Pid, *api.Error) {
	process, err := s.spawn(producer, params, opts...)
	if err != nil || process == nil {
		return nil, err
	}
	pid := process.Pid()
	if pid == nil {
		return nil, api.ErrPidIsNil
	}
	s.processDict.Set(pid.GetUniqId(), process)
	return pid, nil
}

func (s *System) PostMessage(to *api.Pid, message *api.ActorMessage) *api.Error {
	if !api.ValidPid(to) {
		return api.ErrInvalidPid
	}
	if s.IsLocalPid(to) || message.IsBroadcast() {
		process := s.Find(to)
		if process == nil {
			return api.ErrProcessNotExist
		}
		return process.PostMessage(message)
	} else {
		return s.Node().Rpc().PostMessage(to, message)
	}
}

func (s *System) Send(from, to *api.Pid, funcName string, request interface{}) *api.Error {
	data, err := s.Serializer().Marshal(request)
	if err != nil {
		return api.ErrJsonPack
	}
	message := api.BuildInnerMessage(from, to, funcName, data)
	return s.PostMessage(to, message)
}

func (s *System) Call(from, to *api.Pid, funcName string, request, reply interface{}) *api.Error {
	if !api.ValidPid(to) {
		return api.ErrInvalidPid
	}
	requestData, e := s.Serializer().Marshal(request)
	if e != nil {
		zlog.Error("system call", zap.Error(api.ErrJsonPack))
		return api.ErrJsonPack
	}
	message := api.BuildInnerMessage(from, to, funcName, requestData)

	var rsp *api.RespondMessage
	if s.IsLocalPid(to) {
		process := s.Find(to)
		if process == nil {
			return api.ErrProcessNotExist
		}
		rsp = process.PostMessageAndWait(message)
	} else {
		rsp = s.Node().Rpc().Call(to, s.timeout, message)
	}
	if rsp == nil {
		return nil
	}
	if !api.IsOk(rsp.Err) {
		return rsp.Err
	}
	if err := s.unmarshalRsp(rsp, reply); err != nil {
		zlog.Error("system call", zap.Error(err))
		return err
	}
	return nil
}

func (s *System) unmarshalRsp(rsp *api.RespondMessage, reply interface{}) *api.Error {
	if rsp.Data == nil {
		return nil
	}
	if err := s.Serializer().Unmarshal(rsp.Data, reply); err != nil {
		return api.ErrJsonUnPack
	}
	return nil
}

func (s *System) IsLocalPid(pid *api.Pid) bool {
	return pid.GetNodeId() == s.Node().GetID()
}

func (s *System) NextPid() *api.Pid {
	return &api.Pid{
		NodeId: s.Node().GetID(),
		UniqId: s.uniqId.Add(1),
	}
}

func (s *System) spawn(producer api.ActorProducer, params interface{}, opts ...api.ProcessOption) (api.IProcess, *api.Error) {
	opt := loadOptions(opts...)
	mb := getMailBox(s.Node(), opt)

	actor := producer()
	r := s.GetOrSetRouter(actor)
	name := opt.Name
	context := NewBaseActorContext()
	context.actor = actor
	context.system = s
	context.router = r
	context.pid = s.NextPid()
	context.initParams = params
	context.name = name

	process := NewBaseProcess(context, mb)
	context.process = process

	_ = s.RegisterName(name, context.pid)

	mb.RegisterHandlers(context, getDispatcher(opt))
	// notify actor start
	message := buildInitMessage()
	if err := process.PostMessage(message); err != nil {
		return nil, err
	}
	return process, nil
}

func (s *System) AddTimer(pid *api.Pid, d time.Duration, funcName string) *api.Error {
	if pid == nil {
		return api.ErrPidIsNil
	}
	asynctime.AfterFunc(d, func() {
		_ = s.Send(pid, pid, funcName, nil)
	})
	return nil
}

func (s *System) Kill(pid *api.Pid) *api.Error {
	if !api.ValidPid(pid) {
		return api.ErrInvalidPid
	}
	if !s.IsLocalPid(pid) {
		return api.ErrNotLocalPid
	}
	process := s.Find(pid)
	if process == nil {
		return nil
	}
	if err := process.Stop(); err != nil {
		return err
	}
	s.nameDict.Delete(pid.GetName())
	s.processDict.Delete(pid.GetUniqId())
	return nil
}

func (s *System) Stop() *api.Error {
	if err := s.BuiltinStopper.Stop(); err != nil {
		return err
	}
	s.processDict.Range(func(_ uint64, process api.IProcess) bool {
		if process == nil {
			return true
		}
		_ = process.Stop()
		return true
	})
	s.nameDict = maputil.NewConcurrentMap[string, *api.Pid](10)
	s.processDict = maputil.NewConcurrentMap[uint64, api.IProcess](10)
	zlog.Info("actor system stop")
	return nil
}

func (s *System) Group() api.IGroup {
	return s.group
}
