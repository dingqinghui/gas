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
	"github.com/duke-git/lancet/v2/maputil"
	"github.com/panjf2000/gnet/v2"
)

func NewListener(node api.INode, protoAddr string, options ...Option) api.INetServer {
	return newService(node, protoAddr, options...)
}

func newService(node api.INode, protoAddr string, options ...Option) api.INetServer {
	proto, _, err := netx.ParseProtoAddr(protoAddr)
	xerror.Assert(err)
	opts := loadOptions(options...)
	meta := newMeta(node, api.NetListener, opts)
	switch proto {
	case "udp", "udp4", "udp6":
		return newUdpServer(meta, protoAddr)
	case "tcp", "tcp4", "tcp6":
		return newTcpServer(meta, protoAddr)
	}
	return nil
}

func newTcpServer(meta *Meta, protoAddr string) *tcpServer {
	b := new(tcpServer)
	b.builtinServer = newBuiltinServer(protoAddr, meta)
	return b
}

func newUdpServer(meta *Meta, protoAddr string) *udpServer {
	b := new(udpServer)
	b.builtinServer = newBuiltinServer(protoAddr, meta)
	b.dict = maputil.NewConcurrentMap[string, api.ISession](10)
	return b
}

type udpServer struct {
	*builtinServer
	dict *maputil.ConcurrentMap[string, api.ISession]
}

func (b *udpServer) Run() {
	b.run(b)
}

func (b *udpServer) OnTraffic(c gnet.Conn) (action gnet.Action) {
	session := b.Ref(c)
	if session == nil {
		session = newSession(b, b.meta, c)
	}
	if err := session.Traffic(c); err != nil {
		return gnet.Close
	}
	return
}

func (b *udpServer) Link(session api.ISession, c gnet.Conn) {
	remoteAddr := c.RemoteAddr().String()
	b.dict.Set(remoteAddr, session)
	c.SetContext(session)
}

func (b *udpServer) Ref(c gnet.Conn) api.ISession {
	remoteAddr := c.RemoteAddr().String()
	session, _ := b.dict.Get(remoteAddr)
	return session
}

func (b *udpServer) Unlink(c gnet.Conn) {
	remoteAddr := c.RemoteAddr().String()
	session := b.Ref(c)
	b.dict.Delete(remoteAddr)
	if session == nil {
		return
	}
	_ = session.Closed(nil)
}

type tcpServer struct {
	*builtinServer
}

func (b *tcpServer) Run() {
	b.run(b)
}

func (b *tcpServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	newSession(b, b.meta, c)
	return nil, gnet.None
}

func (b *tcpServer) OnTraffic(c gnet.Conn) (action gnet.Action) {
	var err error
	session := b.Ref(c)
	if session == nil {
		return gnet.Close
	}
	if err = session.Traffic(c); err != nil {
		return gnet.Close
	}
	return
}

// OnClose fires when a connection has been closed.
// The parameter err is the last known connection error.
func (b *tcpServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	session := b.Ref(c)
	if session == nil {
		return
	}
	if err = session.Closed(err); err != nil {
		return 0
	}
	return
}

func (b *tcpServer) Ref(c gnet.Conn) api.ISession {
	if c == nil {
		return nil
	}
	ctx := c.Context()
	if ctx == nil {
		return nil
	}
	return ctx.(api.ISession)
}

func (b *tcpServer) Link(session api.ISession, c gnet.Conn) {
	c.SetContext(session)
}

func newBuiltinServer(protoAddr string, meta *Meta) *builtinServer {
	b := new(builtinServer)
	b.protoAddr = protoAddr
	b.meta = meta
	b.Init()
	return b
}

type builtinServer struct {
	gnet.BuiltinEventEngine
	api.BuiltinModule
	meta        *Meta
	protoAddr   string
	proto, addr string
	eng         gnet.Engine
}

func (b *builtinServer) Init() {
	proto, addr, err := netx.ParseProtoAddr(b.protoAddr)
	xerror.Assert(err)
	b.proto = proto
	b.addr = addr
	b.SetNode(b.meta.node)
}
func (b *builtinServer) OnBoot(eng gnet.Engine) (action gnet.Action) {
	b.eng = eng
	return gnet.None
}
func (b *builtinServer) OnShutdown(eng gnet.Engine) {}
func (b *builtinServer) Stop() error {
	if err := b.BuiltinStopper.Stop(); err != nil {
		return err
	}
	return b.eng.Stop(context.Background())
}
func (b *builtinServer) run(handler gnet.EventHandler) {
	b.Node().Workers().Submit(func() {
		xerror.Assert(gnet.Run(handler, b.protoAddr, b.meta.opts.GNetOpts...))
	}, nil)
}
func (b *builtinServer) Unlink(c gnet.Conn) {}
