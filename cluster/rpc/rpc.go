/**
 * @Author: dingQingHui
 * @Description:
 * @File: rpc
 * @Version: 1.0.0
 * @Date: 2024/11/26 15:31
 */

package rpc

import (
	"github.com/dingqinghui/gas/api"
	"github.com/dingqinghui/gas/extend/serializer"
	"go.uber.org/zap"
	"time"
)

func New(node api.INode, msgque api.IRpcMessageQue) api.IRpc {
	r := new(Rpc)
	r.SetNode(node)
	r.msgque = msgque
	r.Init()
	return r
}

type Rpc struct {
	api.BuiltinModule
	msgque api.IRpcMessageQue
}

func (r *Rpc) Run() {
	nodeId := r.Node().GetID()
	// 订阅本节点topic
	topic := genTopic(nodeId)
	r.msgque.Subscribe(topic, func(subj string, data []byte, respondFun api.RpcRespondHandler) {
		r.Node().Workers().Submit(func() {
			if err := r.process(data, respondFun); err != nil {
				r.Log().Error("rpc process", zap.Error(err))
				return
			}

		}, nil)

	})
	r.Log().Info("rpc subscribe", zap.String("topic", topic))
}

func (r *Rpc) marshalMsgWithRequest(msg *Message, request interface{}) ([]byte, error) {
	if request != nil {
		data, err := serializer.Json.Marshal(request)
		if err != nil {
			return nil, err
		}
		msg.Data = data
	}

	buf, err := serializer.Json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (r *Rpc) Send(from, to *api.Pid, methodName string, request interface{}) error {

	msg := newMessage(from, to, methodName, false, 0)
	buf, err := r.marshalMsgWithRequest(msg, request)
	if err != nil {
		return err
	}
	return r.msgque.Send(genTopic(to.GetNodeId()), buf)
}

func (r *Rpc) Call(from, to *api.Pid, methodName string, timeout time.Duration, request interface{}, reply interface{}) error {
	msg := newMessage(from, to, methodName, true, timeout)

	buf, err := r.marshalMsgWithRequest(msg, request)
	if err != nil {
		return err
	}

	replyBuf, err := r.msgque.Call(genTopic(to.GetNodeId()), buf, timeout)
	if err != nil {
		return err
	}

	if err = serializer.Json.Unmarshal(replyBuf, reply); err != nil {
		return err
	}
	return nil
}

func (r *Rpc) process(data []byte, respond api.RpcRespondHandler) error {
	msg := new(Message)
	if err := serializer.Json.Unmarshal(data, msg); err != nil {
		return err
	}

	process := r.Node().ActorSystem().Find(msg.To)
	if process == nil || process.Context() == nil {
		return api.ErrPidIsNil
	}
	router := process.Context().Router()
	if router == nil {
		return api.ErrActorRouterIsNil
	}
	request, reply, err := router.NewArgs(msg.MethodName)
	if err != nil {
		return err
	}
	if err = serializer.Json.Unmarshal(msg.Data, request); err != nil {
		return err
	}
	if reply == nil {
		if err := r.Node().ActorSystem().Send(msg.From, msg.To, msg.MethodName, request); err != nil {
			return err
		}
		return nil
	} else {
		if err = r.Node().ActorSystem().Call(msg.From, msg.To, msg.MethodName, msg.Timeout, request, reply); err != nil {
			return err
		}
		return r.respond(respond, reply)
	}
}

func (r *Rpc) respond(respond api.RpcRespondHandler, reply interface{}) error {
	buf, err := serializer.Json.Marshal(reply)
	if err != nil {
		return err
	}
	respond(buf)
	return nil
}

func (r *Rpc) Stop() error {
	if err := r.BuiltinStopper.Stop(); err != nil {
		return err
	}
	if err := r.msgque.Stop(); err != nil {
		return err
	}
	r.Log().Info("rpc module stop")
	return nil
}

func genTopic(nodeId string) string {
	return "node." + nodeId
}
