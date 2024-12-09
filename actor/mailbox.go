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
	"go.uber.org/zap"
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
	node          api.INode
	inCnt, outCnt atomic.Uint64
}

var _ api.IActorMailbox = &mailbox{}

func NewMailbox(node api.INode) *mailbox {
	m := &mailbox{
		queue: mpsc.NewQueue(),
		node:  node,
	}
	return m
}

func (m *mailbox) RegisterHandlers(invoker api.IActorMessageInvoker, dispatcher api.IActorDispatcher) {
	m.invoker = invoker
	m.dispatch = dispatcher
}

func (m *mailbox) PostMessage(msg api.IMailBoxMessage) error {
	if msg == nil {
		return nil
	}
	m.queue.Push(msg)
	m.inCnt.Add(1)
	return m.schedule()
}

func (m *mailbox) schedule() error {
	if !m.dispatchStat.CompareAndSwap(idle, running) {
		return nil
	}
	if err := m.dispatch.Schedule(m.node, m.process, func(err interface{}) {
		m.node.Log().Error("mailbox panic", zap.Stack("stack"), zap.Error(err.(error)))
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
			_ = m.invokerMessage(msg.(api.IMailBoxMessage))
		} else {
			return
		}
	}
}

// invokerMessage
// @Description: 从队列中读取消息，并调用invoker处理
// @receiver m
// @return error
func (m *mailbox) invokerMessage(msg api.IMailBoxMessage) error {
	if err := m.invoker.InvokerMessage(msg); err != nil {
		return err
	}
	m.outCnt.Add(1)
	return nil
}
