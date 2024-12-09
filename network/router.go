/**
 * @Author: dingQingHui
 * @Description:
 * @File: router
 * @Version: 1.0.0
 * @Date: 2024/12/6 15:05
 */

package network

import (
	"sync"
)

func NewRouter() *Router {
	return &Router{}
}

type Router struct {
	dict sync.Map
}

func (b *Router) Register(msgId uint16, methodName string) {
	b.dict.Store(msgId, methodName)
}

func (b *Router) Get(msgId uint16) string {
	v, ok := b.dict.Load(msgId)
	if !ok {
		return ""
	}
	return v.(string)
}
