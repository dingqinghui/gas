/**
 * @Author: dingQingHui
 * @Description:
 * @File: member
 * @Version: 1.0.0
 * @Date: 2024/11/25 16:46
 */

package api

import (
	"github.com/spf13/viper"
)

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
		Cluster() ICluster
		Base() INodeBase
		Submit(fn func(), recoverFun func(err interface{}))
		Try(fn func(), reFun func(err interface{}))
		NextId() int64
		AddModule(modules ...IModule)
		Terminate(reason string)
		//App() IApp
		//SetApp(app IApp)
	}

	BaseNode struct {
		Id            uint64
		Name, Address string
		Port          int
		Tags          []string
		Meta          map[string]string
	}

	Topology struct {
		All    []INodeBase
		Alive  []INodeBase
		Joined []INodeBase
		Left   []INodeBase
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
