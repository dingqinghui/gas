/**
 * @Author: dingQingHui
 * @Description:
 * @File: cluster
 * @Version: 1.0.0
 * @Date: 2024/11/28 17:40
 */

package api

import "time"

type (
	IBalancer interface {
		Do(nodes []INodeBase, user interface{}) INodeBase // 负载
	}

	RpcRespondHandler func(data []byte)
	RpcProcessHandler func(subj string, data []byte, respond RpcRespondHandler)
	IRpcMessageQue    interface {
		IModule
		Call(subj string, data []byte, timeout time.Duration) ([]byte, error)
		Send(subj string, data []byte) (err error)
		Subscribe(subject string, process RpcProcessHandler)
	}
	IRpc interface {
		IModule
		IActorSender
	}

	IDiscovery interface {
		IModule
		GetById(nodeId string) INodeBase
		GetByKind(kind string) (result []INodeBase)
		GetAll() (result []INodeBase)
		AddNode(node INodeBase) error
		RemoveNode(nodeId string) error
	}
	IDiscoveryProvider interface {
		IModule
		WatchNode(clusterName string, f EventNodeUpdateHandler) error
		AddNode(node INodeBase) error
		RemoveNode(nodeId string) error
	}

	EventNodeUpdateHandler func(waitIndex uint64, nodeDict map[string]*BaseNode)

	ICluster interface {
		IModule
		IActorSender
		SetLB(service string, lb IBalancer)
		Discovery() IDiscovery
		Rpc() IRpc
		Broadcast(from *Pid, service, funcName string, request interface{})
	}
)
