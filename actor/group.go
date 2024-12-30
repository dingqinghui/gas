/**
 * @Author: dingQingHui
 * @Description:
 * @File: eventbus
 * @Version: 1.0.0
 * @Date: 2024/4/23 15:46
 */

package actor

import (
	"github.com/dingqinghui/gas/api"
	"github.com/duke-git/lancet/v2/maputil"
)

type Group struct {
	system api.IActorSystem
	dict   *maputil.ConcurrentMap[string, *maputil.ConcurrentMap[*api.Pid, api.IProcess]]
}

func NewGroup(system api.IActorSystem) *Group {
	return &Group{
		system: system,
		dict:   maputil.NewConcurrentMap[string, *maputil.ConcurrentMap[*api.Pid, api.IProcess]](1),
	}
}
func (m *Group) Add(name string, process api.IProcess) {
	dict, _ := m.dict.GetOrSet(name, maputil.NewConcurrentMap[*api.Pid, api.IProcess](10))
	if dict == nil {
		return
	}
	dict.Set(process.Pid(), process)
}

func (m *Group) Remove(name string, pid *api.Pid) {
	dict, ok := m.dict.Get(name)
	if !ok {
		return
	}
	dict.Delete(pid)
}

func (m *Group) Broadcast(name string, from *api.Pid, msg interface{}) *api.Error {
	if api.GetNode() == nil {
		return nil
	}
	event, ok := m.dict.Get(name)
	if !ok {
		return nil
	}
	data, err := api.GetNode().Serializer().Marshal(msg)
	if err != nil {
		return api.ErrMarshal
	}
	event.Range(func(pid *api.Pid, process api.IProcess) bool {
		message := api.BuildInnerMessage(from, pid, name, data)
		_ = process.PostMessage(message)
		return true
	})
	return nil
}

func (m *Group) Range(eventName string, f func(api.IProcess) bool) {
	event, ok := m.dict.Get(eventName)
	if !ok {
		return
	}
	if f == nil {
		return
	}
	event.Range(func(pid *api.Pid, process api.IProcess) bool {
		return f(process)
	})
}
