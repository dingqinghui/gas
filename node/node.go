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
	"github.com/dingqinghui/gas/cluster"
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
	cluster     api.ICluster
	modules     []api.IModule
	idWorker    *snowflake.IdWorker
	stopChan    chan string
	goCount     atomic.Int64
	panicCount  atomic.Uint64
	pool        *ants.Pool
}

func (a *Node) Init() {
	// init config parse
	a.initViper()
	// init goroutine pool
	a.initGoPool()
	// init node
	a.initBaseNode()
	// init log
	a.initLogger()
	// init actor system
	a.initActorSystem()
	// init cluster
	a.initCluster()
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
	zlog.Init(a)
}

func (a *Node) initActorSystem() {
	a.actorSystem = actor.NewSystem(a)
}

func (a *Node) initCluster() {
	a.cluster = cluster.New(a)
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

func (a *Node) Cluster() api.ICluster {
	return a.cluster
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
