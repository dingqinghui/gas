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
	"github.com/panjf2000/gnet/v2"
)

type Option func(*Options)
type HandshakeAuthFunc func(ctx api.IActorContext, data []byte) ([]byte, error)

type Options struct {
	AgentProducer api.ActorProducer
	HandshakeAuth HandshakeAuthFunc
	HandshakeBody []byte
	GNetOpts      []gnet.Option
	Router        api.INetRouter
	Serializer    api.ISerializer
}

func WithAgentProducer(producer api.ActorProducer) Option {
	return func(op *Options) {
		op.AgentProducer = producer
	}
}

func WithRouter(router api.INetRouter) Option {
	return func(op *Options) {
		op.Router = router
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

func loadOptions(options ...Option) *Options {
	opts := new(Options)
	for _, option := range options {
		option(opts)
	}
	if opts.AgentProducer == nil {
		opts.AgentProducer = func() api.IActor {
			return new(AgentActor)
		}
	}
	if opts.HandshakeAuth == nil {
		opts.HandshakeAuth = func(ctx api.IActorContext, data []byte) ([]byte, error) {
			return nil, nil
		}
	}
	if opts.Router == nil {
		opts.Router = new(Router)
	}

	if opts.Serializer == nil {
		opts.Serializer = serializer.Json
	}
	return opts
}
