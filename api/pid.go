/**
 * @Author: dingQingHui
 * @Description:
 * @File: pid
 * @Version: 1.0.0
 * @Date: 2024/12/9 14:49
 */

package api

import "github.com/duke-git/lancet/v2/convertor"

type Pid struct {
	NodeId string
	UniqId uint64
	Name   string
}

func (p *Pid) GetNodeId() string {
	return p.NodeId
}
func (p *Pid) GetUniqId() uint64 {
	return p.UniqId
}
func (p *Pid) String() string {
	if p == nil {
		return ""
	}
	return p.NodeId + "." + convertor.ToString(p.UniqId)
}
func (p *Pid) GetName() string {
	return p.Name
}
func NewRemotePid(nodeId string, name string) *Pid {
	return &Pid{
		NodeId: nodeId,
		Name:   name,
	}
}
func NewPidWithName(name string) *Pid {
	return &Pid{
		Name: name,
	}
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
