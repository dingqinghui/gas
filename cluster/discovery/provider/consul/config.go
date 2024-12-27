/**
 * @Author: dingQingHui
 * @Description:
 * @File: config
 * @Version: 1.0.0
 * @Date: 2024/12/5 9:58
 */

package consul

import (
	"github.com/dingqinghui/gas/api"
	"time"
)

func initConfig() *config {
	node := api.GetNode()
	c := new(config)
	vp := node.GetViper().Sub("cluster.consul")
	c.address = vp.GetString("address")
	c.watchWaitTime = vp.GetDuration("watchWaitTime")
	c.healthTtl = vp.GetDuration("healthTtl")
	c.deregister = vp.GetDuration("deregister")
	return c
}

type config struct {
	address       string
	watchWaitTime time.Duration
	healthTtl     time.Duration
	deregister    time.Duration
}
