/**
 * @Author: dingQingHui
 * @Description:
 * @File: func
 * @Version: 1.0.0
 * @Date: 2024/11/18 11:26
 */

package discovery

import (
	"github.com/dingqinghui/gas/api"
	"github.com/dingqinghui/gas/extend/xerror"
	"github.com/duke-git/lancet/v2/convertor"
	"golang.org/x/exp/slices"
)

func New(node api.INode, clusterName string, provider api.IDiscoveryProvider) api.IDiscovery {
	d := new(discovery)
	xerror.NilAssert(provider)
	d.provider = provider
	d.clusterName = clusterName
	d.SetNode(node)
	d.Init()
	return d
}

type discovery struct {
	api.BuiltinModule
	provider    api.IDiscoveryProvider
	clusterName string
	waitIndex   uint64
	nodeDict    map[string]*api.BaseNode
}

func (d *discovery) Name() string {
	return "discovery"
}

func (d *discovery) Run() {
	if d.provider == nil {
		return
	}
	// watch node
	api.Assert(d.provider.WatchNode(d.clusterName, func(waitIndex uint64, nodeDict map[string]*api.BaseNode) {
		if waitIndex <= d.waitIndex {
			return
		}
		d.waitIndex = waitIndex
		d.nodeDict = nodeDict
	}))

	// add node
	api.Assert(d.AddNode(d.Node().Base()))
}

func (d *discovery) GetById(nodeId string) api.INodeBase {
	v, _ := d.nodeDict[nodeId]
	return v
}

func (d *discovery) GetByKind(kind string) (result []api.INodeBase) {
	for _, node := range d.nodeDict {
		if slices.Contains(node.GetTags(), kind) {
			result = append(result, convertor.DeepClone(node))
		}
	}
	return
}

func (d *discovery) GetAll() (result []api.INodeBase) {
	for _, node := range d.nodeDict {
		result = append(result, convertor.DeepClone(node))
	}
	return
}

func (d *discovery) AddNode(node api.INodeBase) *api.Error {
	if d.provider == nil {
		return api.ErrDiscoveryProviderIsNil
	}
	return d.provider.AddNode(node)
}

func (d *discovery) RemoveNode(nodeId string) *api.Error {
	if d.provider == nil {
		return api.ErrDiscoveryProviderIsNil
	}
	return d.provider.RemoveNode(nodeId)
}

func (d *discovery) Stop() *api.Error {
	if err := d.BuiltinStopper.Stop(); err != nil {
		return err
	}
	if err := d.RemoveNode(convertor.ToString(d.Node().GetID())); err != nil {
		return err
	}
	if d.provider != nil {
		d.provider.Stop()
	}
	d.Log().Info("discovery module stop")
	return nil
}
