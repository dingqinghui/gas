/**
 * @Author: dingQingHui
 * @Description:
 * @File: router
 * @Version: 1.0.0
 * @Date: 2024/12/6 15:05
 */

package network

import (
	"github.com/dingqinghui/gas/api"
	"sync"
)

func NewRouters() *Routers {
	return &Routers{}
}

type Routers struct {
	dict sync.Map
}

func (b *Routers) Add(mid uint16, router api.INetRouter) {
	b.dict.Store(mid, router)
}

func (b *Routers) Get(msgId uint16) api.INetRouter {
	v, ok := b.dict.Load(msgId)
	if !ok {
		return nil
	}
	return v.(api.INetRouter)
}

type Router struct {
	Service string
	ActorId uint64
	Method  string
}

func (b *Router) GetService() string {
	return b.Service
}

func (b *Router) GetActorId() uint64 {
	return b.ActorId
}

func (b *Router) GetMethod() string {
	return b.Method
}
