/**
 * @Author: dingQingHui
 * @Description:
 * @File: rpc
 * @Version: 1.0.0
 * @Date: 2024/11/26 15:31
 */

package rpc

import (
	"fmt"
	"github.com/dingqinghui/gas/api"
	"github.com/dingqinghui/gas/extend/serializer"
	"github.com/dingqinghui/gas/zlog"
	"github.com/duke-git/lancet/v2/convertor"
	"go.uber.org/zap"
	"time"
)

func New(msgque api.IRpcMessageQue) api.IRpc {
	r := new(Rpc)
	r.msgque = msgque
	api.GetNode().AddModule(r)
	return r
}

type Rpc struct {
	api.BuiltinModule
	msgque api.IRpcMessageQue
}

func (r *Rpc) Run() {
	process := func(subj string, data []byte, respondFun api.RpcRespondHandler) {
		if err := r.process(data, respondFun); err != nil {
			zlog.Error("rpc process", zap.Error(err))
			return
		}
	}
	// 订阅本节点topic
	topic := r.genNodeTopic(api.GetNode().GetID())
	r.msgque.Subscribe(topic, process)

	// 订阅广播组
	for _, tag := range api.GetNode().GetTags() {
		topic = r.genBroadcastTopic(tag)
		r.msgque.Subscribe(topic, process)
	}
}

func (r *Rpc) Broadcast(message *api.Message) *api.Error {
	if message == nil || message.To == nil {
		return api.ErrInvalidActorMessage
	}
	if !message.IsBroadcast() {
		return api.ErrInvalidActorMessage
	}
	topic := r.genBroadcastTopic(message.To.Name)
	return r.send(topic, message)
}

func (r *Rpc) PostMessage(to *api.Pid, message *api.Message) *api.Error {
	topic := r.genNodeTopic(to.GetNodeId())
	if message.IsBroadcast() {
		return api.ErrInvalidActorMessage
	}
	return r.send(topic, message)
}

func (r *Rpc) send(topic string, message *api.Message) *api.Error {
	buf, err := api.GetNode().Serializer().Marshal(message)
	if err != nil {
		zlog.Error("rpc marshal request err", zap.Error(err))
		return api.ErrJsonPack
	}
	return r.msgque.Send(topic, buf)
}

func (r *Rpc) Call(to *api.Pid, timeout time.Duration, message *api.Message) (rsp *api.RespondMessage) {
	rsp = new(api.RespondMessage)
	data, err := api.GetNode().Serializer().Marshal(message)
	if err != nil {
		zlog.Error("rpc marshal request err", zap.Error(err))
		rsp.Err = api.ErrJsonPack
		return
	}
	rspData, err := r.msgque.Call(r.genNodeTopic(to.GetNodeId()), data, timeout)
	if err != nil {
		zlog.Error("rpc call  err", zap.Error(err))
		rsp.Err = api.ErrNatsSend
		return nil
	}

	if err = serializer.Json.Unmarshal(rspData, rsp); err != nil {
		zlog.Error("system call", zap.Error(err))
		rsp.Err = api.ErrJsonUnPack
		return
	}
	return
}

func (r *Rpc) process(data []byte, respond api.RpcRespondHandler) *api.Error {
	message := new(api.Message)
	if err := api.GetNode().Serializer().Unmarshal(data, message); err != nil {
		zlog.Error("rpc process  err", zap.Error(err))
		return api.ErrJsonUnPack
	}
	if respond != nil {
		message.SetRespond(func(rsp *api.RespondMessage) *api.Error {
			rspData, err := serializer.Json.Marshal(rsp)
			if err != nil {
				return api.ErrJsonUnPack
			}
			return respond(rspData)
		})
	}
	return api.GetNode().System().PostMessage(message.To, message)
}

func (r *Rpc) genNodeTopic(nodeId uint64) string {
	return "node." + convertor.ToString(nodeId)
}

func (r *Rpc) genBroadcastTopic(service string) string {
	return fmt.Sprintf("broadcast.tag.%s", service)
}

func (r *Rpc) Stop() *api.Error {
	if err := r.BuiltinStopper.Stop(); err != nil {
		return err
	}
	if err := r.msgque.Stop(); err != nil {
		return err
	}
	zlog.Info("rpc module stop")
	return nil
}
