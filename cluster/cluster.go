/**
 * @Author: dingQingHui
 * @Description:
 * @File: cluster
 * @Version: 1.0.0
 * @Date: 2024/11/21 10:36
 */

package cluster

import (
	"github.com/dingqinghui/gas/api"
	"github.com/dingqinghui/gas/cluster/balancer"
	"github.com/dingqinghui/gas/cluster/discovery"
	"github.com/dingqinghui/gas/cluster/discovery/provider/consul"
	"github.com/dingqinghui/gas/cluster/rpc"
	"github.com/dingqinghui/gas/cluster/rpc/provider/nats"
	"github.com/dingqinghui/gas/extend/xerror"
	"github.com/duke-git/lancet/v2/maputil"
)

func New(node api.INode) api.ICluster {
	c := new(cluster)
	c.SetNode(node)
	c.Init()
	return c
}

type cluster struct {
	api.BuiltinModule
	rpc       api.IRpc
	discovery api.IDiscovery
	lbDict    *maputil.ConcurrentMap[string, api.IBalancer]
}

func (c *cluster) Name() string {
	return "cluster"
}

func (c *cluster) Init() {
	c.lbDict = maputil.NewConcurrentMap[string, api.IBalancer](10)
	c.initRpc()
	c.initDiscovery()
}

func (c *cluster) initDiscovery() {
	viper := c.Node().GetViper()
	clusterName := viper.GetString("cluster.name")
	provider, err := consul.NewConsulProvider(c.Node())
	xerror.Assert(err)
	c.discovery = discovery.New(c.Node(), clusterName, provider)
}

func (c *cluster) initRpc() {
	msgque := nats.New(c.Node())
	c.rpc = rpc.New(c.Node(), msgque)
}

func (c *cluster) Run() {
	c.discovery.Run()
	c.rpc.Run()
}

func (c *cluster) SetLB(service string, lb api.IBalancer) {
	c.lbDict.Set(service, lb)
}

func (c *cluster) balance(service string) api.INodeBase {
	lb, ok := c.lbDict.Get(service)
	if !ok || lb == nil {
		lb = balancer.NewRandom()
	}
	nodes := c.discovery.GetByKind(service)
	if lb == nil {
		return nil
	}
	return lb.Do(nodes, nil)
}

func (c *cluster) Discovery() api.IDiscovery {
	return c.discovery
}

func (c *cluster) Rpc() api.IRpc {
	return c.rpc
}

func (c *cluster) NewPid(service string, lb api.IBalancer, user interface{}) *api.Pid {
	nodes := c.Discovery().GetByKind(service)
	node := lb.Do(nodes, user)
	if node == nil {
		return nil
	}
	return api.NewRemotePid(node.GetID(), service)
}

func (c *cluster) Stop() *api.Error {
	if err := c.BuiltinStopper.Stop(); err != nil {
		return err
	}
	if err := c.discovery.Stop(); err != nil {
		return err
	}
	if err := c.rpc.Stop(); err != nil {
		return err
	}
	c.Log().Info("cluster module stop")
	return nil
}
