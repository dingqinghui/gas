/**
 * @Author: dingQingHui
 * @Description:
 * @File: message
 * @Version: 1.0.0
 * @Date: 2024/11/20 10:38
 */

package rpc

import (
	"github.com/dingqinghui/gas/api"
	"time"
)

type Message struct {
	From       *api.Pid
	To         *api.Pid
	MethodName string
	IsSync     bool
	Timeout    time.Duration
	Data       []byte
	Session    *api.Session
	Mid        uint16
}

func newMessage(from *api.Pid, to *api.Pid, methodName string, isSync bool, timeout time.Duration) *Message {
	return &Message{
		From:       from,
		To:         to,
		MethodName: methodName,
		IsSync:     isSync,
		Timeout:    timeout,
	}
}
