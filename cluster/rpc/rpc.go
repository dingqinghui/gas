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
	r.serializer = serializer.Json
	r.Init()
	return r
}

type Rpc struct {
	api.BuiltinModule
	msgque     api.IRpcMessageQue
	serializer api.ISerializer
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
		data, err := r.serializer.Marshal(request)
		if err != nil {
			r.Log().Error("rpc marshal request err", zap.Error(err))
			return nil, err
		}
		msg.Data = data
	}

	buf, err := r.serializer.Marshal(msg)
	if err != nil {
		r.Log().Error("rpc marshal request err", zap.Error(err))
		return nil, err
	}
	return buf, nil
}

func (r *Rpc) Send(from, to *api.Pid, methodName string, request interface{}) error {
	msg := newMessage(from, to, methodName, false, 0)
	buf, err := r.marshalMsgWithRequest(msg, request)
	if err != nil {
		r.Log().Error("rpc send  err", zap.Error(err))
		return err
	}
	return r.msgque.Send(genTopic(to.GetNodeId()), buf)
}

func (r *Rpc) Call(from, to *api.Pid, methodName string, timeout time.Duration, request interface{}, reply interface{}) error {
	msg := newMessage(from, to, methodName, true, timeout)

	buf, err := r.marshalMsgWithRequest(msg, request)
	if err != nil {
		r.Log().Error("rpc call  err", zap.Error(err))
		return err
	}

	replyBuf, err := r.msgque.Call(genTopic(to.GetNodeId()), buf, timeout)
	if err != nil {
		r.Log().Error("rpc call  err", zap.Error(err))
		return err
	}

	if err = r.serializer.Unmarshal(replyBuf, reply); err != nil {
		r.Log().Error("rpc call  err", zap.Error(err))
		return err
	}
	return nil
}

func (r *Rpc) process(data []byte, respond api.RpcRespondHandler) error {
	msg := new(Message)
	if err := r.serializer.Unmarshal(data, msg); err != nil {
		r.Log().Error("rpc process  err", zap.Error(err))
		return err
	}

	process := r.Node().ActorSystem().Find(msg.To)
	if process == nil || process.Context() == nil {
		r.Log().Error("rpc process  err", zap.Error(api.ErrPidIsNil))
		return api.ErrPidIsNil
	}
	router := process.Context().Router()
	if router == nil {
		r.Log().Error("rpc process  err", zap.Error(api.ErrActorRouterIsNil))
		return api.ErrActorRouterIsNil
	}
	request, reply, err := router.NewArgs(msg.MethodName)
	if err != nil {
		r.Log().Error("rpc process  err",
			zap.String("methodName", msg.MethodName), zap.Error(err))
		return err
	}
	if err = r.serializer.Unmarshal(msg.Data, request); err != nil {
		r.Log().Error("rpc process  err",
			zap.String("methodName", msg.MethodName), zap.Error(err))
		return err
	}
	if reply == nil {
		if err = process.Send(msg.From, msg.MethodName, request); err != nil {
			r.Log().Error("rpc process  err",
				zap.String("methodName", msg.MethodName), zap.Error(err))
			return err
		}
		return nil
	} else {
		wait, callErr := process.Call(msg.From, msg.MethodName, msg.Timeout, request, reply)
		if callErr != nil {
			r.Log().Error("rpc process  err",
				zap.String("methodName", msg.MethodName), zap.Error(err))
			return err
		}
		r.Node().Workers().Submit(func() {
			if waitErr := wait.Wait(); waitErr != nil {
				r.Log().Error("rpc process  err",
					zap.String("methodName", msg.MethodName), zap.Error(waitErr))
				return
			}
			if err = r.respond(respond, reply); err != nil {
				r.Log().Error("rpc process  err",
					zap.String("methodName", msg.MethodName), zap.Error(err))
				return
			}
		}, nil)
	}
	return nil
}

func (r *Rpc) respond(respond api.RpcRespondHandler, reply interface{}) error {
	data, err := r.serializer.Marshal(reply)
	if err != nil {
		r.Log().Error("rpc respond  err", zap.Error(err))
		return err
	}
	respond(data)
	return nil
}

func (r *Rpc) SetSerializer(serializer api.ISerializer) {
	r.serializer = serializer
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
