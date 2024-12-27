/**
 * @Author: dingQingHui
 * @Description:
 * @File: pid
 * @Version: 1.0.0
 * @Date: 2024/12/9 14:49
 */

package api

import "fmt"

func NewRemotePid(nodeId uint64, service string) *Pid {
	return &Pid{
		NodeId: nodeId,
		Name:   service,
	}
}

type Pid struct {
	NodeId uint64
	UniqId uint64
	Name   string
}

func ValidPid(pid *Pid) bool {
	if pid == nil {
		return false
	}
	if pid.GetUniqId() > 0 {
		return true
	}
	if pid.GetName() != "" {
		return true
	}
	return false
}
func (p *Pid) GetNodeId() uint64 {
	return p.NodeId
}
func (p *Pid) GetUniqId() uint64 {
	return p.UniqId
}
func (p *Pid) String() string {
	if p == nil {
		return ""
	}
	return fmt.Sprintf("%v.%v", p.NodeId, p.UniqId)
}
func (p *Pid) GetName() string {
	return p.Name
}
