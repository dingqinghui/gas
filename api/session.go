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
	ses := new(Session)
	ses.Agent = entity.GetAgent()
	ses.entity = entity
	return ses
}

type Session struct {
	Agent   *Pid
	Message *message.Message
	entity  INetEntity
	ctx     IActorContext
}

func (s *Session) GetEntity() INetEntity {
	return s.entity
}

func (s *Session) Dup() *Session {
	return &Session{
		Agent:   s.Agent,
		Message: nil,
		entity:  s.entity,
	}
}

func (s *Session) SetContext(ctx IActorContext) {
	s.ctx = ctx
}

func (s *Session) Response(payload interface{}) *Error {
	body, err := GetNode().Serializer().Marshal(payload)
	if err != nil {
		return ErrMarshal
	}
	m := message.NewWithData(body)
	m.Copy(s.Message)
	return s.push(m)
}

func (s *Session) ResponseErr(err *Error) *Error {
	m := message.NewErr(err.Id)
	m.Copy(s.Message)
	return s.push(m)
}

func (s *Session) Push(mid uint16, payload interface{}) *Error {
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
	entity := s.GetEntity()
	if entity != nil {
		return entity.SendMessage(m)
	} else {
		return s.forwardToAgent(sessionPush, m)
	}
}

func (s *Session) forwardToAgent(method string, payload interface{}) *Error {
	mData, wrong := GetNode().Serializer().Marshal(payload)
	if wrong != nil {
		return ErrMarshal
	}
	actorMsg := BuildInnerMessage(s.ctx.Self(), s.Agent, method, mData)
	return GetNode().System().PostMessage(s.Agent, actorMsg)
}

func (s *Session) Close(reason *Error) *Error {
	if s.entity == nil {
		return s.entity.Close(reason)
	} else {
		return s.forwardToAgent(sessionClose, reason)
	}
}
