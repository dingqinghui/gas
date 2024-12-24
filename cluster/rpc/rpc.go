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
	"github.com/dingqinghui/gas/zlog"
	"github.com/duke-git/lancet/v2/convertor"
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
		r.Node().Submit(func() {
			if err := r.process(data, respondFun); err != nil {
				zlog.Error("rpc process", zap.Error(err))
				return
			}

		}, nil)

	})
	zlog.Info("rpc subscribe", zap.String("topic", topic))
}

func (r *Rpc) PostMessage(to *api.Pid, message *api.ActorMessage) *api.Error {
	buf, err := r.serializer.Marshal(message)
	if err != nil {
		zlog.Error("rpc marshal request err", zap.Error(err))
		return api.ErrJsonPack
	}
	return r.msgque.Send(genTopic(to.GetNodeId()), buf)
}

func (r *Rpc) Call(to *api.Pid, timeout time.Duration, message *api.ActorMessage) (rsp *api.RespondMessage) {
	rsp = new(api.RespondMessage)
	data, err := r.serializer.Marshal(message)
	if err != nil {
		zlog.Error("rpc marshal request err", zap.Error(err))
		rsp.Err = api.ErrJsonPack
		return
	}
	rspData, err := r.msgque.Call(genTopic(to.GetNodeId()), data, timeout)
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
	message := new(api.ActorMessage)
	if err := r.serializer.Unmarshal(data, message); err != nil {
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
	return r.Node().System().PostMessage(message.To, message)
}

func (r *Rpc) SetSerializer(serializer api.ISerializer) {
	r.serializer = serializer
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

func genTopic(nodeId uint64) string {
	return "node." + convertor.ToString(nodeId)
}
