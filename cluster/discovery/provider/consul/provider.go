/**
 * @Author: dingQingHui
 * @Description:
 * @File: consul_provider
 * @Version: 1.0.0
 * @Date: 2024/4/25 14:09
 */

package consul

import (
	api2 "github.com/dingqinghui/gas/api"
	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
	"time"
)

type consulProvider struct {
	api2.BuiltinModule
	address   string
	client    *api.Client
	waitIndex uint64
	status    string
	cfg       *config
}

func NewConsulProvider(node api2.INode) (api2.IDiscoveryProvider, error) {
	c := new(consulProvider)
	c.SetNode(node)
	c.Init()
	return c, nil
}

func (c *consulProvider) Init() {
	c.cfg = initConfig(c.Node())
	c.status = "pass"

	if err := c.connect(c.cfg.address); err != nil {
		c.Log().Panic("consul connect err", zap.Error(err))
		return
	}
}

func (c *consulProvider) connect(consulAddress string) error {
	apiConfig := api.DefaultConfig()
	apiConfig.Address = consulAddress
	client, err := api.NewClient(apiConfig)
	if err != nil {
		c.Log().Error("consul new client err", zap.Error(err))
		return err
	}
	c.client = client
	c.Log().Info("consul connect success..", zap.String("consulAddress", consulAddress))
	return nil
}

func (c *consulProvider) Name() string {
	return "consul"
}

func (c *consulProvider) WatchNode(clusterName string, f api2.EventNodeUpdateHandler) error {
	if err := c.monitorMemberStatusChanges(clusterName, f); err != nil {
		return err
	}
	go func() {
		c.Log().Info("consul watch begin", zap.String("clusterName", clusterName))
		for !c.IsStop() {
			if err := c.monitorMemberStatusChanges(clusterName, f); err != nil {
				return
			}
		}
		c.Log().Info("consul watch end", zap.String("clusterName", clusterName))
	}()
	return nil
}

func (c *consulProvider) monitorMemberStatusChanges(clusterName string, f api2.EventNodeUpdateHandler) error {
	opt := &api.QueryOptions{
		WaitIndex: c.waitIndex,
		WaitTime:  c.cfg.watchWaitTime,
	}
	services, meta, err := c.client.Health().Service(clusterName, "", true, opt)
	if err != nil {
		c.Log().Error("consul discovery agent err", zap.String("clusterName", clusterName), zap.Error(err))
		return err
	}
	nodeDict := make(map[string]*api2.BaseNode)
	for _, service := range services {
		node := &api2.BaseNode{
			Id:      service.Service.ID,
			Name:    clusterName,
			Address: service.Service.Address,
			Port:    service.Service.Port,
			Tags:    service.Service.Tags,
			Meta:    service.Service.Meta,
		}
		nodeDict[node.Id] = node
	}
	c.waitIndex = meta.LastIndex
	f(c.waitIndex, nodeDict)
	return nil
}

func (c *consulProvider) AddNode(node api2.INodeBase) error {
	// 注册服务
	check := &api.AgentServiceCheck{
		TTL:                            (c.cfg.healthTtl).String(),
		DeregisterCriticalServiceAfter: (c.cfg.deregister).String(),
	}
	registration := &api.AgentServiceRegistration{
		ID:      node.GetID(),
		Name:    node.GetName(),
		Address: node.GetAddress(),
		Port:    node.GetPort(),
		Tags:    node.GetTags(),
		Meta:    node.GetMeta(),
		Check:   check,
	}
	if err := c.client.Agent().ServiceRegister(registration); err != nil {
		c.Log().Error("consul node  register err", zap.String("nodeId", node.GetID()), zap.Error(err))
		return err
	}

	go c.healthCheckActor()

	c.Log().Info("consul node  register ", zap.String("nodeId", node.GetID()),
		zap.String("nodeName", node.GetName()), zap.String("address", node.GetAddress()),
		zap.Int("port", node.GetPort()), zap.Strings("tags", node.GetTags()))
	return nil
}

func (c *consulProvider) healthCheckActor() {
	c.Log().Info("consul health check begin")
	for !c.IsStop() {
		if err := c.client.Agent().UpdateTTL("service:"+c.Node().GetID(), "", c.status); err != nil {
			c.Node().Log().Error("consul health agent err", zap.String("nodeId", c.Node().GetID()),
				zap.String("status", c.status), zap.Error(err))
			return
		}
		time.Sleep(c.cfg.healthTtl / 2)
	}
	c.Log().Info("consul health check end")
}

func (c *consulProvider) UpdateStatus(status string) {
	c.status = status
}

func (c *consulProvider) RemoveNode(nodeId string) error {
	if err := c.client.Agent().ServiceDeregister(nodeId); err != nil {
		c.Log().Error("consul service deregister", zap.String("nodeId", nodeId), zap.Error(err))
		return err
	}

	c.Log().Info("consul service deregister", zap.String("nodeId", nodeId))
	return nil
}

func (c *consulProvider) Stop() error {
	if err := c.BuiltinStopper.Stop(); err != nil {
		return err
	}
	c.Log().Info("consul module stop")
	return nil
}
