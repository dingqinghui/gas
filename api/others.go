/**
 * @Author: dingQingHui
 * @Description:
 * @File: others
 * @Version: 1.0.0
 * @Date: 2024/12/5 11:17
 */

package api

import (
	"sync/atomic"
)

type (
	IWorkers interface {
		Submit(fn func(), recoverFun func(err interface{}))
		Try(fn func(), reFun func(err interface{}))
	}

	IStopper interface {
		IsStop() bool
		Stop() *Error
	}

	BuiltinStopper struct {
		stop atomic.Bool
	}

	Marshaler interface {
		Marshal(interface{}) ([]byte, error)
	}
	Unmarshaler interface {
		Unmarshal([]byte, interface{}) error
	}
	ISerializer interface {
		Marshaler
		Unmarshaler
	}

	IModule interface {
		IModuleLifecycle
		Name() string
		SetNode(node INode)
		Node() INode
	}

	IModuleLifecycle interface {
		IStopper
		Init()
		Run()
	}
	BuiltinModule struct {
		BuiltinStopper
		node INode
	}
)

func (b *BuiltinStopper) IsStop() bool {
	return b.stop.CompareAndSwap(true, true)
}

func (b *BuiltinStopper) Stop() *Error {
	if !b.stop.CompareAndSwap(false, true) {
		return ErrStopped
	}
	return nil
}

func (b *BuiltinModule) Init()              {}
func (b *BuiltinModule) Run()               {}
func (b *BuiltinModule) Name() string       { return "" }
func (b *BuiltinModule) SetNode(node INode) { b.node = node }
func (b *BuiltinModule) Node() INode        { return b.node }
