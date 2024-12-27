/**
 * @Author: dingQingHui
 * @Description:
 * @File: config
 * @Version: 1.0.0
 * @Date: 2024/12/5 10:11
 */

package nats

import "github.com/dingqinghui/gas/api"

func initConfig() *config {
	node := api.GetNode()
	c := new(config)
	vp := node.GetViper().Sub("cluster.nats")
	c.urls = vp.GetString("urls")
	c.server = vp.GetStringSlice("server")
	c.recChanSize = vp.GetInt("recChanSize")
	return c
}

type config struct {
	urls        string
	server      []string
	recChanSize int
}
