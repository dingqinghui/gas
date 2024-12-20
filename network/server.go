/**
 * @Author: dingQingHui
 * @Description:
 * @File: server
 * @Version: 1.0.0
 * @Date: 2024/11/29 14:49
 */

package network

import (
	"context"
	"github.com/dingqinghui/gas/api"
	"github.com/dingqinghui/gas/extend/netx"
	"github.com/dingqinghui/gas/extend/xerror"
	"github.com/dingqinghui/gas/zlog"
	"github.com/duke-git/lancet/v2/maputil"
	"github.com/panjf2000/gnet/v2"
	"go.uber.org/zap"
)

func NewListener(node api.INode, protoAddr string, options ...Option) api.INetServer {
	return newService(node, protoAddr, options...)
}

func newService(node api.INode, protoAddr string, options ...Option) api.INetServer {
	proto, _, err := netx.ParseProtoAddr(protoAddr)
	xerror.Assert(err)
	opts := loadOptions(options...)
	switch proto {
	case "udp", "udp4", "udp6":
		return newUdpServer(node, api.NetListener, opts, protoAddr)
	case "tcp", "tcp4", "tcp6":
		return newTcpServer(node, api.NetListener, opts, protoAddr)
	}
	return nil
}

func newTcpServer(node api.INode, typ api.NetEntityType, opts *Options, protoAddr string) *tcpServer {
	b := new(tcpServer)
	b.builtinServer = newBuiltinServer(node, typ, opts, protoAddr)
	return b
}

func newUdpServer(node api.INode, typ api.NetEntityType, opts *Options, protoAddr string) *udpServer {
	b := new(udpServer)
	b.builtinServer = newBuiltinServer(node, typ, opts, protoAddr)
	b.dict = maputil.NewConcurrentMap[string, api.INetEntity](10)
	return b
}

type udpServer struct {
	*builtinServer
	dict *maputil.ConcurrentMap[string, api.INetEntity]
}

func (b *udpServer) Run() {
	b.run(b)
}

func (b *udpServer) OnTraffic(c gnet.Conn) (action gnet.Action) {
	entity := b.Ref(c)
	if entity == nil {
		entity = newEntity(b, b.opts, c)
		b.Link(entity, c)
	}
	if err := entity.Traffic(c); err != nil {
		zlog.Error("udp server traffic err",
			zap.String("remote", entity.RemoteAddr()), zap.Error(err))
		return gnet.Close
	}
	return
}

func (b *udpServer) Link(entity api.INetEntity, c gnet.Conn) {
	if c == nil || c.RemoteAddr() == nil {
		return
	}
	remoteAddr := c.RemoteAddr().String()
	b.dict.Set(remoteAddr, entity)
	c.SetContext(entity)
}

func (b *udpServer) Ref(c gnet.Conn) api.INetEntity {
	if c == nil || c.RemoteAddr() == nil {
		return nil
	}
	remoteAddr := c.RemoteAddr().String()
	entity, _ := b.dict.Get(remoteAddr)
	return entity
}

func (b *udpServer) Unlink(c gnet.Conn) {
	if c == nil || c.RemoteAddr() == nil {
		return
	}
	remoteAddr := c.RemoteAddr().String()
	entity := b.Ref(c)
	b.dict.Delete(remoteAddr)
	if entity == nil {
		return
	}
	_ = entity.Closed(nil)
}

type tcpServer struct {
	*builtinServer
}

func (b *tcpServer) Run() {
	b.run(b)
}

func (b *tcpServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	entity := newEntity(b, b.opts, c)
	b.Link(entity, c)
	return nil, gnet.None
}

func (b *tcpServer) OnTraffic(c gnet.Conn) (action gnet.Action) {
	var err error
	entity := b.Ref(c)
	if entity == nil {
		zlog.Warn("tcp server traffic err", zap.Error(api.ErrNetEntityIsNil))
		return gnet.Close
	}
	if err = entity.Traffic(c); err != nil {
		return gnet.Close
	}
	return
}

// OnClose fires when a connection has been closed.
// The parameter err is the last known connection error.
func (b *tcpServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	entity := b.Ref(c)
	if entity == nil {
		zlog.Warn("tcp server onclose err",
			zap.Error(api.ErrNetEntityIsNil))
		return
	}
	if err = entity.Closed(err); err != nil {
		zlog.Error("tcp server onclose err",
			zap.String("remote", entity.RemoteAddr()),
			zap.Error(err))
		return 0
	}
	return
}

func (b *tcpServer) Ref(c gnet.Conn) api.INetEntity {
	if c == nil {
		return nil
	}
	ctx := c.Context()
	if ctx == nil {
		return nil
	}
	return ctx.(api.INetEntity)
}

func (b *tcpServer) Link(entity api.INetEntity, c gnet.Conn) {
	c.SetContext(entity)
}

func newBuiltinServer(node api.INode, typ api.NetEntityType, opts *Options, protoAddr string) *builtinServer {
	b := new(builtinServer)
	b.protoAddr = protoAddr
	b.node = node
	b.typ = typ
	b.opts = opts
	b.Init()
	return b
}

type builtinServer struct {
	gnet.BuiltinEventEngine
	api.BuiltinModule
	node        api.INode
	opts        *Options
	typ         api.NetEntityType
	protoAddr   string
	proto, addr string
	eng         gnet.Engine
}

func (b *builtinServer) Init() {
	proto, addr, err := netx.ParseProtoAddr(b.protoAddr)
	xerror.Assert(err)
	b.proto = proto
	b.addr = addr
	b.SetNode(b.node)
}
func (b *builtinServer) OnBoot(eng gnet.Engine) (action gnet.Action) {
	b.eng = eng
	return gnet.None
}
func (b *builtinServer) OnShutdown(eng gnet.Engine) {}
func (b *builtinServer) Stop() *api.Error {
	if err := b.BuiltinStopper.Stop(); err != nil {
		return err
	}
	_ = b.eng.Stop(context.Background())
	return nil
}
func (b *builtinServer) run(handler gnet.EventHandler) {
	b.Node().Workers().Submit(func() {
		xerror.Assert(gnet.Run(handler, b.protoAddr, b.Options().GNetOpts...))
	}, nil)
}
func (b *builtinServer) Unlink(c gnet.Conn) {}
func (b *builtinServer) Options() *Options {
	return b.opts
}
func (b *builtinServer) Typ() api.NetEntityType {
	return b.typ
}
