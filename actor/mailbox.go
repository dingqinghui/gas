/**
 * @Author: dingQingHui
 * @Description:
 * @File: mailbox
 * @Version: 1.0.0
 * @Date: 2024/10/15 14:27
 */

package actor

import (
	"github.com/dingqinghui/gas/api"
	"github.com/dingqinghui/gas/extend/mpsc"
	"runtime"
	"sync/atomic"
)

const (
	idle int32 = iota
	running
)

var _ api.IActorMailbox = &mailbox{}

type mailbox struct {
	invoker       api.IActorMessageInvoker
	queue         *mpsc.Queue
	dispatch      api.IActorDispatcher
	dispatchStat  atomic.Int32
	inCnt, outCnt atomic.Uint64
}

var _ api.IActorMailbox = &mailbox{}

func NewMailbox() *mailbox {
	m := &mailbox{
		queue: mpsc.NewQueue(),
	}
	return m
}

func (m *mailbox) RegisterHandlers(invoker api.IActorMessageInvoker, dispatcher api.IActorDispatcher) {
	m.invoker = invoker
	m.dispatch = dispatcher
}

func (m *mailbox) PostMessage(msg interface{}) *api.Error {
	if msg == nil {
		return nil
	}
	m.queue.Push(msg)
	m.inCnt.Add(1)
	return m.schedule()
}

func (m *mailbox) schedule() *api.Error {
	if !m.dispatchStat.CompareAndSwap(idle, running) {
		return nil
	}
	if err := m.dispatch.Schedule(m.process, func(err interface{}) {
		//_ = m.invoker.InvokerMessage(NewMailBoxMessage(PanicFuncName, nil, err))
	}); err != nil {
		return err
	}
	return nil
}

func (m *mailbox) process() {
	m.run()
	m.dispatchStat.CompareAndSwap(running, idle)
}

func (m *mailbox) run() {
	throughput := m.dispatch.Throughput()
	var i int
	for true {
		if m.queue.Empty() {
			return
		}
		if i > throughput {
			i = 0
			runtime.Gosched()
			continue
		}
		i++
		msg := m.queue.Pop()
		if msg != nil {
			_ = m.invokerMessage(msg)
		} else {
			return
		}
	}
}

// invokerMessage
// @Description: 从队列中读取消息，并调用invoker处理
// @receiver m
// @return error
func (m *mailbox) invokerMessage(msg interface{}) error {
	if err := m.invoker.InvokerMessage(msg); err != nil {
		return err
	}
	m.outCnt.Add(1)
	return nil
}
