/**
 * @Author: dingQingHui
 * @Description:
 * @File: session
 * @Version: 1.0.0
 * @Date: 2024/12/27 14:50
 */

package api

import (
	"github.com/dingqinghui/gas/network/message"
)

const (
	sessionPush  = "Push"
	sessionClose = "Close"
)

func NewSession(entity INetEntity) *Session {
	return &Session{
		entity: entity,
		ctx:    nil,
	}
}

type Session struct {
	Agent  *Pid
	Mid    uint32
	Index  uint32
	ctx    IActorContext
	entity INetEntity
}

func (s *Session) SetContext(ctx IActorContext) {
	s.ctx = ctx
}

func (s *Session) Response(payload interface{}) *Error {
	if GetNode() == nil || GetNode().Serializer() == nil {
		return nil
	}
	body, err := GetNode().Serializer().Marshal(payload)
	if err != nil {
		return ErrMarshal
	}
	m := message.NewWithData(body)
	m.ID = uint16(s.Mid)
	m.Index = s.Index
	return s.push(m)
}

func (s *Session) ResponseErr(err *Error) *Error {
	m := message.NewErr(err.Id)
	m.ID = uint16(s.Mid)
	m.Index = s.Index
	return s.push(m)
}

func (s *Session) Push(mid uint16, payload interface{}) *Error {
	if GetNode() == nil || GetNode().Serializer() == nil {
		return nil
	}
	body, err := GetNode().Serializer().Marshal(payload)
	if err != nil {
		return ErrMarshal
	}
	m := message.New()
	m.ID = mid
	m.Data = body
	return s.push(m)
}

func (s *Session) push(m *message.Message) *Error {
	entity := s.entity
	if entity != nil {
		return entity.SendMessage(m)
	} else {
		return s.forwardToAgent(sessionPush, m)
	}
}

func (s *Session) forwardToAgent(method string, payload interface{}) *Error {
	if GetNode() == nil || GetNode().Serializer() == nil {
		return nil
	}
	data, err := GetNode().Serializer().Marshal(payload)
	if err != nil {
		return ErrMarshal
	}
	actMessage := &Message{
		Session: s,
		Typ:     MessageEnumInner,
		Method:  method,
		Data:    data,
		To:      s.Agent,
		From:    s.ctx.Self(),
	}
	return GetNode().System().PostMessage(s.Agent, actMessage)
}

func (s *Session) Close(reason *Error) *Error {
	entity := s.entity
	if entity == nil {
		return entity.Close(reason)
	} else {
		return s.forwardToAgent(sessionClose, reason)
	}
}
