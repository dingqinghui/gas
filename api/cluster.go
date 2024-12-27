/**
 * @Author: dingQingHui
 * @Description:
 * @File: cluster
 * @Version: 1.0.0
 * @Date: 2024/11/28 17:40
 */

package api

import (
	"time"
)

var (
	ClusterUpdateGroup = "OnUpdateClusterGroup"
)

type (
	IBalancer interface {
		Do(nodes []INodeBase, user interface{}) INodeBase // 负载
	}

	RpcRespondHandler func(data []byte) *Error
	RpcProcessHandler func(subj string, data []byte, respond RpcRespondHandler)
	IRpcMessageQue    interface {
		IModule
		Call(subj string, data []byte, timeout time.Duration) ([]byte, error)
		Send(subj string, data []byte) (err *Error)
		Subscribe(subject string, process RpcProcessHandler)
	}
	IRpc interface {
		IModule
		Call(to *Pid, timeout time.Duration, message *Message) (rsp *RespondMessage)
		PostMessage(to *Pid, message *Message) *Error
		Broadcast(message *Message) *Error
	}

	IDiscovery interface {
		IModule
		GetById(nodeId uint64) INodeBase
		GetByKind(kind string) (result []INodeBase)
		GetAll() (result []INodeBase)
		AddNode(node INodeBase) *Error
		RemoveNode(nodeId string) *Error
	}
	IDiscoveryProvider interface {
		IModule
		WatchNode(clusterName string, f EventNodeUpdateHandler) *Error
		AddNode(node INodeBase) *Error
		RemoveNode(nodeId string) *Error
	}

	EventNodeUpdateHandler func(waitIndex uint64, nodeDict map[uint64]*BaseNode)

	ICluster interface {
		IModule
		Discovery() IDiscovery
		Rpc() IRpc
		NewPid(service string, lb IBalancer, user interface{}) *Pid
	}

	Topology struct {
		All    []INodeBase
		Alive  []INodeBase
		Joined []INodeBase
		Left   []INodeBase
	}
)
