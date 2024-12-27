/**
 * @Author: dingQingHui
 * @Description:
 * @File: app
 * @Version: 1.0.0
 * @Date: 2024/9/20 17:50
 */

package node

import (
	"fmt"
	"github.com/dingqinghui/gas/actor"
	"github.com/dingqinghui/gas/api"
	"github.com/dingqinghui/gas/cluster/discovery"
	"github.com/dingqinghui/gas/cluster/discovery/provider/consul"
	"github.com/dingqinghui/gas/cluster/rpc"
	"github.com/dingqinghui/gas/cluster/rpc/provider/nats"
	"github.com/dingqinghui/gas/extend/serializer"
	"github.com/dingqinghui/gas/extend/snowflake"
	"github.com/dingqinghui/gas/extend/xerror"
	"github.com/dingqinghui/gas/zlog"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/panjf2000/ants/v2"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
)

func New(configPath string) api.INode {
	node := &Node{
		configPath: configPath,
		BaseNode:   new(api.BaseNode),
		stopChan:   make(chan string),
	}
	api.SetNode(node)
	node.Init()
	return node
}

var _ api.INode = &Node{}

type Node struct {
	api.BuiltinModule
	*api.BaseNode
	configPath  string
	actorSystem api.IActorSystem
	viper       *viper.Viper
	rpc         api.IRpc
	discovery   api.IDiscovery
	modules     []api.IModule
	serializer  api.ISerializer
	idWorker    *snowflake.IdWorker
	stopChan    chan string
	goCount     atomic.Int64
	panicCount  atomic.Uint64
	pool        *ants.Pool
}

func (a *Node) Init() {
	// init config parse
	a.initViper()
	// init serializer
	a.initSerializer()
	// init goroutine pool
	a.initGoPool()
	// init node
	a.initBaseNode()
	// init log
	a.initLogger()
	// init actor system
	a.initActorSystem()
	// init discovery
	a.initDiscovery()
	// init rpc
	a.initRpc()

	for _, module := range a.modules {
		module.Init()
	}
	zlog.Info("node init finish............")
}

func (a *Node) initGoPool() {
	pool, err := ants.NewPool(1000)
	xerror.Assert(err)
	a.pool = pool
}

func (a *Node) initViper() {
	a.viper = viper.New()
	a.viper.SetConfigFile(a.configPath)
	err := a.viper.ReadInConfig() // 读取配置文件
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	fmt.Printf("init viper path:%s\n", a.configPath)
}

func (a *Node) initBaseNode() {
	vp := a.viper.Sub("node")
	a.BaseNode.Name = a.viper.GetString("cluster.name")
	a.BaseNode.Id = vp.GetUint64("id")
	a.BaseNode.Tags = vp.GetStringSlice("tags")
	a.BaseNode.Meta = vp.GetStringMapString("meta")

	fmt.Printf("init node id:%d  type:%s\n", a.BaseNode.Id, a.BaseNode.Name)

	idWorker, err := snowflake.NewIdWorker(int64(a.GetID()))
	xerror.Assert(err)
	a.idWorker = idWorker
}

func (a *Node) initLogger() {
	zlog.Init()
}

func (a *Node) initActorSystem() {
	a.actorSystem = actor.NewSystem()
}

func (a *Node) initSerializer() {
	a.serializer = serializer.Json
}

func (a *Node) initDiscovery() {
	vp := a.GetViper()
	clusterName := vp.GetString("cluster.name")
	provider, err := consul.NewConsulProvider()
	xerror.Assert(err)
	a.discovery = discovery.New(clusterName, provider)
}

func (a *Node) initRpc() {
	msgque := nats.New()
	a.rpc = rpc.New(msgque)
}

func (a *Node) Run() {
	for _, module := range a.modules {
		module.Run()
	}
	zlog.Info("node running............")
}

func (a *Node) Base() api.INodeBase {
	return a.BaseNode
}

func (a *Node) GetViper() *viper.Viper {
	return a.viper
}

func (a *Node) System() api.IActorSystem {
	return a.actorSystem
}

func (a *Node) AddModule(modules ...api.IModule) {
	a.modules = append(a.modules, modules...)
}

func (a *Node) Discovery() api.IDiscovery {
	return a.discovery
}

func (a *Node) Rpc() api.IRpc {
	return a.rpc
}

func (a *Node) Serializer() api.ISerializer {
	return a.serializer
}

func (a *Node) Name() string {
	return "node" + convertor.ToString(a.GetID())
}

func (a *Node) Wait() {
	stopChanForSys := make(chan os.Signal, 1)
	signal.Notify(stopChanForSys, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	select {
	case s := <-stopChanForSys:
		a.terminate("system signal:" + s.String())
	case reason := <-a.stopChan:
		a.terminate(reason)
	}
}

func (a *Node) NextId() int64 {
	if a.idWorker == nil {
		return 0
	}
	id, err := a.idWorker.NextId()
	if err != nil {
		zlog.Error("nextId", zap.Error(err))
	}
	return id
}

func (a *Node) Terminate(reason string) {
	a.stopChan <- reason
}

func (a *Node) terminate(reason string) {
	if a.modules != nil {
		for i := len(a.modules) - 1; i > 0; i-- {
			module := a.modules[i]
			if module == nil {
				continue
			}
			_ = module.Stop()
		}
	}
	zlog.Info("node terminate", zap.String("reason", reason))
}

func (a *Node) Submit(fn func(), recoverFun func(err interface{})) {
	err := a.pool.Submit(func() {
		a.goCount.Add(1)
		a.Try(fn, recoverFun)
		a.goCount.Add(-1)
	})
	if err != nil {
		return
	}
}

func (a *Node) Try(fn func(), reFun func(err interface{})) {
	defer func() {
		if err := recover(); err != nil {
			a.panicCount.Add(1)
			if reFun != nil {
				reFun(err)
			}
			zlog.Error("panic", zap.Error(err.(error)), zap.Stack("stack"))
		}
	}()
	fn()
}
