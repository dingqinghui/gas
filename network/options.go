/**
 * @Author: dingQingHui
 * @Description:
 * @File: options
 * @Version: 1.0.0
 * @Date: 2024/12/6 14:30
 */

package network

import (
	"github.com/dingqinghui/gas/api"
	"github.com/dingqinghui/gas/extend/serializer"
	"github.com/dingqinghui/gas/network/message"
	"github.com/panjf2000/gnet/v2"
	"time"
)

type Option func(*Options)
type HandshakeAuthFunc func(session api.INetEntity, data []byte) ([]byte, *api.Error)
type RouterFunc func(session *api.Session, msg *message.Message) (*api.Pid, string, *api.Error)

func loadOptions(options ...Option) *Options {
	opts := defaultOptions()
	for _, option := range options {
		option(opts)
	}
	return opts
}

func defaultOptions() *Options {
	opts := &Options{
		AgentProducer: func() api.IActor {
			return new(AgentActor)
		},
		HandshakeAuth: func(session api.INetEntity, data []byte) ([]byte, *api.Error) {
			return nil, nil
		},
		HandshakeBody:    nil,
		GNetOpts:         nil,
		Serializer:       serializer.Json,
		HeartBeatTimeout: time.Second * 5,
	}
	return opts
}

type Options struct {
	AgentProducer    api.ActorProducer
	HandshakeAuth    HandshakeAuthFunc
	HandshakeBody    []byte
	GNetOpts         []gnet.Option
	RouterHandler    RouterFunc
	Serializer       api.ISerializer
	HeartBeatTimeout time.Duration
}

func WithRouterHandler(routerHandler RouterFunc) Option {
	return func(op *Options) {
		op.RouterHandler = routerHandler
	}
}

func WithAgentProducer(producer api.ActorProducer) Option {
	return func(op *Options) {
		op.AgentProducer = producer
	}
}

func WithHandshakeAuth(handshake HandshakeAuthFunc) Option {
	return func(op *Options) {
		op.HandshakeAuth = handshake
	}
}

func WithHandshakeBody(handshakeBody []byte) Option {
	return func(op *Options) {
		op.HandshakeBody = handshakeBody
	}
}

func WithSerializer(serializer api.ISerializer) Option {
	return func(op *Options) {
		op.Serializer = serializer
	}
}
func WithHeartBeatTimeout(hearTimeout time.Duration) Option {
	return func(op *Options) {
		op.HeartBeatTimeout = hearTimeout
	}
}
