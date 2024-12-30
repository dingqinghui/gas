/**
 * @Author: dingQingHui
 * @Description:
 * @File: member
 * @Version: 1.0.0
 * @Date: 2024/11/25 16:46
 */

package api

import (
	"github.com/dingqinghui/gas/extend/xerror"
	"github.com/spf13/viper"
)

var (
	currentNode INode
)

func SetNode(node INode) {
	currentNode = node
}

func GetNode() INode {
	xerror.NilAssert(currentNode)
	return currentNode
}

type (
	INodeBase interface {
		GetName() string
		GetID() uint64
		GetAddress() string
		GetPort() int
		GetTags() []string
		GetMeta() map[string]string
	}

	INode interface {
		INodeBase
		Init()
		Run()
		Wait()
		GetViper() *viper.Viper
		System() IActorSystem
		Discovery() IDiscovery
		Rpc() IRpc
		Base() INodeBase
		Submit(fn func(), recoverFun func(err interface{}))
		Try(fn func(), reFun func(err interface{}))
		NextId() int64
		AddModule(modules ...IModule)
		Terminate(reason string)
		Serializer() ISerializer
	}

	BaseNode struct {
		Id            uint64
		Name, Address string
		Port          int
		Tags          []string
		Meta          map[string]string
	}
)

func (b *BaseNode) GetName() string {
	return b.Name
}

func (b *BaseNode) GetID() uint64 {
	return b.Id
}

func (b *BaseNode) GetAddress() string {
	return b.Address
}

func (b *BaseNode) GetPort() int {
	return b.Port
}

func (b *BaseNode) GetTags() []string {
	return b.Tags
}

func (b *BaseNode) GetMeta() map[string]string {
	return b.Meta
}
