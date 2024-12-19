/**
 * @Author: dingQingHui
 * @Description:
 * @File: meta
 * @Version: 1.0.0
 * @Date: 2024/12/12 15:54
 */

package network

import "github.com/dingqinghui/gas/api"

func newMeta(node api.INode, typ api.NetConnectionType, opts *Options) *Meta {
	return &Meta{
		node: node,
		opts: opts,
		typ:  typ,
	}
}

type Meta struct {
	node api.INode
	opts *Options
	typ  api.NetConnectionType
}
