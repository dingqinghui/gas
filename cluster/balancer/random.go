package balancer

import (
	"github.com/dingqinghui/gas/api"
	"math/rand"
)

func NewRandom() *random {
	return new(random)
}

type random struct {
}

func (b *random) Do(nodeArray []api.INodeBase, user interface{}) api.INodeBase {
	if len(nodeArray) <= 0 {
		return nil
	}
	return nodeArray[rand.Intn(len(nodeArray))]
}
