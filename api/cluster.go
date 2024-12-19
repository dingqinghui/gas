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
		//IActorSender
		SetSerializer(serializer ISerializer)
		//PostNetMessage(to *Pid, session *Session, methodName string, mid uint16, data []byte) error

		Call(to *Pid, timeout time.Duration, message *ActorMessage) (rsp *RespondMessage)
		PostMessage(to *Pid, message *ActorMessage) *Error
	}

	IDiscovery interface {
		IModule
		GetById(nodeId string) INodeBase
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

	EventNodeUpdateHandler func(waitIndex uint64, nodeDict map[string]*BaseNode)

	ICluster interface {
		IModule
		//IActorSender
		SetLB(service string, lb IBalancer)
		Discovery() IDiscovery
		Rpc() IRpc
		NewPid(service string, lb IBalancer, user interface{}) *Pid
		//Broadcast(from *Pid, service, funcName string, request interface{})
		//PostRemoteMessage(session *Session, to *Pid, methodName string, mid uint16, data []byte) *Error
	}
)
