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
func (goroutineDispatcher) Schedule(node api.INode, fn func(), recoverFun func(err interface{})) *api.Error {
	node.Submit(fn, recoverFun)
	return nil
}

func (d goroutineDispatcher) Throughput() int {
	return int(d)
}

// 同步调度器
type synchronizedDispatcher int

func (synchronizedDispatcher) Schedule(node api.INode, fn func(), recoverFun func(err interface{})) *api.Error {
	node.Try(fn, recoverFun)
	return nil
}

func (d synchronizedDispatcher) Throughput() int {
	return int(d)
}

func NewSynchronizedDispatcher(throughput int) api.IActorDispatcher {
	return synchronizedDispatcher(throughput)
}
