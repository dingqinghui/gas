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
	"github.com/dingqinghui/gas/workers"
	"github.com/dingqinghui/gas/zlog"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func New(configPath string) api.INode {
	node := &Node{
		configPath: configPath,
		BaseNode:   new(api.BaseNode),
		stopChan:   make(chan string),
	}
	node.workers = workers.New(node)
	node.Init()
	return node
}

var _ api.INode = &Node{}

type Node struct {
	api.BuiltinModule
	*api.BaseNode
	configPath  string
	logger      api.IZLogger
	actorSystem api.IActorSystem
	viper       *viper.Viper
	cluster     api.ICluster
	workers     api.IWorkers
	modules     []api.IModule
	stopChan    chan string
}

func (a *Node) Init() {
	// init config parse
	a.initViper()
	// init node
	a.initBaseNode()
	// init log
	a.initLogger()
	// init actor system
	a.initActorSystem()
	// init cluster
	a.initCluster()
	a.Log().Info("node init finish............")
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
	a.BaseNode.Id = vp.GetString("id")
	a.BaseNode.Tags = vp.GetStringSlice("tags")
	a.BaseNode.Meta = vp.GetStringMapString("meta")

	fmt.Printf("init node id:%s  type:%s\n", a.BaseNode.Id, a.BaseNode.Name)
}

func (a *Node) initLogger() {
	a.logger = zlog.New(a)
	a.AddModule(a.logger)
}

func (a *Node) initActorSystem() {
	a.actorSystem = actor.NewSystem(a)
	a.AddModule(a.actorSystem)
}

func (a *Node) initCluster() {
	a.cluster = cluster.New(a)
	a.AddModule(a.cluster)
}

func (a *Node) Run() {
	for _, module := range a.modules {
		module.Run()
	}
	a.Log().Info("node running............")
	a.wait()
}

func (a *Node) Base() api.INodeBase {
	return a.BaseNode
}

func (a *Node) GetViper() *viper.Viper {
	return a.viper
}

func (a *Node) ActorSystem() api.IActorSystem {
	return a.actorSystem
}

func (a *Node) AddModule(modules ...api.IModule) {
	a.modules = append(a.modules, modules...)
}

func (a *Node) Log() api.IZLogger {
	return a.logger
}
func (a *Node) Cluster() api.ICluster {
	return a.cluster
}
func (a *Node) Workers() api.IWorkers {
	return a.workers
}

func (a *Node) Name() string {
	return "node" + a.GetID()
}

func (a *Node) isLocal(pid *api.Pid) bool {
	return pid.GetNodeId() == a.GetID()
}

func (a *Node) Send(from, pid *api.Pid, funcName string, request interface{}) error {
	if !api.ValidPid(pid) {
		return api.ErrInvalidPid
	}
	if a.isLocal(pid) {
		return a.ActorSystem().Send(from, pid, funcName, request)
	} else {
		return a.Cluster().Send(from, pid, funcName, request)
	}
}

func (a *Node) Call(from, pid *api.Pid, funcName string, timeout time.Duration, request, reply interface{}) error {
	if a.isLocal(pid) {
		return a.ActorSystem().Call(from, pid, funcName, timeout, request, reply)
	} else {
		return a.Cluster().Call(from, pid, funcName, timeout, request, reply)
	}
}

func (a *Node) wait() {
	stopChanForSys := make(chan os.Signal, 1)
	signal.Notify(stopChanForSys, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	select {
	case s := <-stopChanForSys:
		a.terminate("system signal:" + s.String())
	case reason := <-a.stopChan:
		a.terminate(reason)
	}
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
			module.Stop()
		}
	}
	a.Log().Info("node terminate", zap.String("reason", reason))
}
