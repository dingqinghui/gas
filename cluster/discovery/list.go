/**
 * @Author: dingQingHui
 * @Description:
 * @File: list
 * @Version: 1.0.0
 * @Date: 2024/12/27 17:48
 */

package discovery

import "github.com/dingqinghui/gas/api"

type NodeList struct {
	Dict        map[uint64]*api.BaseNode
	LastEventId uint64
}

func NewNodeList() *NodeList {
	return &NodeList{
		Dict: make(map[uint64]*api.BaseNode),
	}
}

func (m *NodeList) UpdateClusterTopology(nodeDict map[uint64]*api.BaseNode, lastEventId uint64) *api.Topology {
	topology := &api.Topology{}
	if m.LastEventId >= lastEventId {
		return topology
	}
	for _, node := range nodeDict {
		if _, ok := m.Dict[node.GetID()]; ok {
			topology.Alive = append(topology.Alive, node)
		} else {
			topology.Joined = append(topology.Joined, node)
		}
		topology.All = append(topology.All, node)
	}
	for id := range m.Dict {
		if _, ok := nodeDict[id]; !ok {
			topology.Left = append(topology.Left, m.Dict[id])
		}
	}
	m.Dict = nodeDict
	m.LastEventId = lastEventId
	return topology
}
