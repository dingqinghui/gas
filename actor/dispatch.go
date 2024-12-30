// Package actor
// @Description:

package actor

import (
	"github.com/dingqinghui/gas/api"
)

// 协程调度器
type goroutineDispatcher int

func NewDefaultDispatcher(throughput int) api.IActorDispatcher {
	return goroutineDispatcher(throughput)
}
func (goroutineDispatcher) Schedule(fn func(), recoverFun func(err interface{})) *api.Error {
	if api.GetNode() == nil {
		return nil
	}
	api.GetNode().Submit(fn, recoverFun)
	return nil
}

func (d goroutineDispatcher) Throughput() int {
	return int(d)
}

// 同步调度器
type synchronizedDispatcher int

func (synchronizedDispatcher) Schedule(fn func(), recoverFun func(err interface{})) *api.Error {
	if api.GetNode() == nil {
		return nil
	}
	api.GetNode().Try(fn, recoverFun)
	return nil
}

func (d synchronizedDispatcher) Throughput() int {
	return int(d)
}

func NewSynchronizedDispatcher(throughput int) api.IActorDispatcher {
	return synchronizedDispatcher(throughput)
}
