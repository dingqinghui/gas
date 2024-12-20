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
	"github.com/dingqinghui/gas/cluster/discovery"
	"github.com/dingqinghui/gas/cluster/discovery/provider/consul"
	"github.com/dingqinghui/gas/cluster/rpc"
	"github.com/dingqinghui/gas/cluster/rpc/provider/nats"
	"github.com/dingqinghui/gas/extend/xerror"
	"github.com/dingqinghui/gas/zlog"
)

func New(node api.INode) api.ICluster {
	c := new(cluster)
	c.SetNode(node)
	c.Init()
	node.AddModule(c)
	return c
}

type cluster struct {
	api.BuiltinModule
	rpc       api.IRpc
	discovery api.IDiscovery
}

func (c *cluster) Name() string {
	return "cluster"
}

func (c *cluster) Init() {
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
	zlog.Info("cluster module stop")
	return nil
}
