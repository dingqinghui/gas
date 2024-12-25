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

type EventHandler func(context api.IActorContext, msg interface{}) error

type EventBus struct {
	system api.IActorSystem
	events *maputil.ConcurrentMap[string, *maputil.ConcurrentMap[*api.Pid, api.IProcess]]
}

func NewEventBus(system api.IActorSystem) *EventBus {
	return &EventBus{
		system: system,
		events: maputil.NewConcurrentMap[string, *maputil.ConcurrentMap[*api.Pid, api.IProcess]](1),
	}
}
func (m *EventBus) Register(eventName string, process api.IProcess) {
	event, _ := m.events.GetOrSet(eventName, maputil.NewConcurrentMap[*api.Pid, api.IProcess](10))
	if event == nil {
		return
	}
	event.Set(process.Pid(), process)
}

func (m *EventBus) UnRegister(eventName string, pid *api.Pid) {
	event, ok := m.events.Get(eventName)
	if !ok {
		return
	}
	event.Delete(pid)
}

func (m *EventBus) Notify(eventName string, from *api.Pid, msg interface{}) *api.Error {
	event, ok := m.events.Get(eventName)
	if !ok {
		return nil
	}
	data, err := m.system.Serializer().Marshal(msg)
	if err != nil {
		return api.ErrMarshal
	}
	event.Range(func(pid *api.Pid, process api.IProcess) bool {
		message := api.BuildInnerMessage(from, pid, eventName, data)
		_ = process.PostMessage(message)
		return true
	})
	return nil
}

func (m *EventBus) Range(eventName string, f func(api.IProcess) bool) {
	event, ok := m.events.Get(eventName)
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
