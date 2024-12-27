/**
 * @Author: dingQingHui
 * @Description:
 * @File: cluster
 * @Version: 1.0.0
 * @Date: 2024/11/21 10:36
 */

package cluster

import (
	"github.com/dingqinghui/gas/api"
)

func NewPid(service string, lb api.IBalancer, user interface{}) *api.Pid {
	nodes := api.GetNode().Discovery().GetByKind(service)
	selectNode := lb.Do(nodes, user)
	if selectNode == nil {
		return nil
	}
	return api.NewRemotePid(selectNode.GetID(), service)
}
