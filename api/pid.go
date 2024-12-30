/**
 * @Author: dingQingHui
 * @Description:
 * @File: pid
 * @Version: 1.0.0
 * @Date: 2024/12/9 14:49
 */

package api

type Pid struct {
	NodeId uint64
	UniqId uint64
	Name   string
}

func (x *Pid) GetNodeId() uint64 {
	if x != nil {
		return x.NodeId
	}
	return 0
}

func (x *Pid) GetUniqId() uint64 {
	if x != nil {
		return x.UniqId
	}
	return 0
}

func (x *Pid) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
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
