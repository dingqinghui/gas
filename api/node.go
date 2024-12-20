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
		Workers() IWorkers
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
		EventId uint64
		Alive   []INodeBase
		Joined  []INodeBase
		Left    []INodeBase
	}

	NodeList struct {
		Dict        map[string]*BaseNode
		LastEventId uint64
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

//func NewNodeList() *NodeList {
//	return &NodeList{
//		Dict: make(map[string]*BaseNode),
//	}
//}
//
//func (m *NodeList) UpdateClusterTopology(nodeDict map[string]*BaseNode, lastEventId uint64) *Topology {
//	if m.LastEventId >= lastEventId {
//		return nil
//	}
//	tplg := &Topology{EventId: lastEventId}
//	for _, node := range nodeDict {
//		if _, ok := m.Dict[node.GetID()]; ok {
//			tplg.Alive = append(tplg.Alive, node)
//		} else {
//			tplg.Joined = append(tplg.Joined, node)
//		}
//	}
//	for id := range m.Dict {
//		if _, ok := nodeDict[id]; !ok {
//			tplg.Left = append(tplg.Left, m.Dict[id])
//		}
//	}
//	m.Dict = nodeDict
//	m.LastEventId = lastEventId
//	return tplg
//}
