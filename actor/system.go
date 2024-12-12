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
	"github.com/duke-git/lancet/v2/maputil"
	"sync/atomic"
	"time"
)

type System struct {
	api.BuiltinModule
	uniqId      atomic.Uint64
	nameDict    *maputil.ConcurrentMap[string, *api.Pid]
	processDict *maputil.ConcurrentMap[uint64, api.IProcess]
	routerHub   api.IActorRouterHub
}

func NewSystem(node api.INode) api.IActorSystem {
	s := new(System)
	s.SetNode(node)
	s.Init()
	return s
}

func (s *System) Init() {
	s.nameDict = maputil.NewConcurrentMap[string, *api.Pid](10)
	s.processDict = maputil.NewConcurrentMap[uint64, api.IProcess](10)
	s.routerHub = NewRouterHub()
}

func (s *System) Name() string {
	return "actorSystem"
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

func (s *System) RegisterName(name string, pid *api.Pid) error {
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

func (s *System) UnregisterName(name string) (*api.Pid, error) {
	v, ok := s.nameDict.GetAndDelete(name)
	if !ok {
		return nil, api.ErrActorNameNotExist
	}
	return v, nil
}

func (s *System) Spawn(producer api.ActorProducer, params interface{}, opts ...api.ProcessOption) (*api.Pid, error) {
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

func (s *System) SpawnWithName(name string, producer api.ActorProducer, params interface{}, opts ...api.ProcessOption) (*api.Pid, error) {
	pid, err := s.Spawn(producer, params, opts...)
	if err != nil || pid == nil {
		return nil, err
	}
	return pid, s.RegisterName(name, pid)
}

func (s *System) Send(from, to *api.Pid, funcName string, request interface{}) error {
	if !api.ValidPid(to) {
		return api.ErrInvalidPid
	}
	if !s.isLocalPid(to) {
		return api.ErrNotLocalPid
	}
	process := s.Find(to)
	if process == nil {
		return api.ErrProcessNotExist
	}
	return process.Send(from, funcName, request)
}

func (s *System) Call(from, to *api.Pid, funcName string, timeout time.Duration, request, reply interface{}) error {
	if !api.ValidPid(to) {
		return api.ErrInvalidPid
	}
	if !s.isLocalPid(to) {
		return api.ErrNotLocalPid
	}
	process := s.Find(to)
	if process == nil {
		return api.ErrProcessNotExist
	}
	return process.CallAndWait(from, funcName, timeout, request, reply)
}

func (s *System) isLocalPid(pid *api.Pid) bool {
	return pid.GetNodeId() == s.Node().GetID()
}

func (s *System) NextPid() *api.Pid {
	return &api.Pid{
		NodeId: s.Node().GetID(),
		UniqId: s.uniqId.Add(1),
	}
}

func (s *System) spawn(producer api.ActorProducer, params interface{}, opts ...api.ProcessOption) (api.IProcess, error) {
	opt := loadOptions(opts...)
	mb := getMailBox(s.Node(), opt)

	actor := producer()
	r := s.routerHub.GetOrSet(actor)

	context := NewBaseActorContext()
	context.actor = actor
	context.system = s
	context.router = r
	context.pid = s.NextPid()
	context.initParams = params
	context.IZLogger = s.Log()

	process := NewBaseProcess(context, mb)
	context.process = process

	mb.RegisterHandlers(context, getDispatcher(opt))
	// notify actor start
	if err := process.Send(process.Pid(), InitFuncName, context); err != nil {
		return nil, err
	}
	return process, nil
}

func (s *System) RouterHub() api.IActorRouterHub {
	return s.routerHub
}

func (s *System) AddTimer(pid *api.Pid, d time.Duration, funcName string) error {
	if pid == nil {
		return api.ErrPidIsNil
	}
	asynctime.AfterFunc(d, func() {
		_ = s.Send(pid, pid, funcName, nil)
	})
	return nil
}

func (s *System) Kill(pid *api.Pid) error {
	if !api.ValidPid(pid) {
		return api.ErrInvalidPid
	}
	if !s.isLocalPid(pid) {
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

func (s *System) Stop() error {
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
	s.Log().Info("actor system stop")
	return nil
}
