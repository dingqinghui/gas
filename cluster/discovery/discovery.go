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
	"github.com/dingqinghui/gas/zlog"
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
	list        *api.NodeList
}

func (d *discovery) Init() {
	d.list = api.NewNodeList()
}

func (d *discovery) Name() string {
	return "discovery"
}

func (d *discovery) Run() {
	if d.provider == nil {
		return
	}
	// watch node
	api.Assert(d.provider.WatchNode(d.clusterName, func(waitIndex uint64, nodeDict map[uint64]*api.BaseNode) {
		if waitIndex <= d.list.LastEventId {
			return
		}
		topology := d.list.UpdateClusterTopology(nodeDict, waitIndex)
		if len(topology.Left) != 0 || len(topology.Joined) != 0 {
			_ = d.Node().System().EventBus().Notify(api.EventUpdateCluster, nil, topology)
		}
	}))
	// add node
	api.Assert(d.AddNode(d.Node().Base()))
}

func (d *discovery) GetById(nodeId uint64) api.INodeBase {
	v, _ := d.list.Dict[nodeId]
	return v
}

func (d *discovery) GetByKind(kind string) (result []api.INodeBase) {
	for _, node := range d.list.Dict {
		if slices.Contains(node.GetTags(), kind) {
			result = append(result, convertor.DeepClone(node))
		}
	}
	return
}

func (d *discovery) GetAll() (result []api.INodeBase) {
	for _, node := range d.list.Dict {
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
		return d.provider.Stop()
	}
	zlog.Info("discovery module stop")
	return nil
}
