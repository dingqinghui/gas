package balancer

import (
	"github.com/dingqinghui/gas/api"
	"sync"
)

func NewRoundRobin() *roundRobin {
	return new(roundRobin)
}

type roundRobin struct {
	sync.Mutex
	curIndex int
}

func (b *roundRobin) Do(nodeArray []api.INodeBase, user interface{}) api.INodeBase {
	b.Lock()
	defer b.Unlock()
	if len(nodeArray) <= 0 {
		return nil
	}
	if b.curIndex >= len(nodeArray) {
		b.curIndex = 0
	}

	inst := nodeArray[b.curIndex]
	b.curIndex++
	return inst
}
