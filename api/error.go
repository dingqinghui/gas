/**
 * @Author: dingQingHui
 * @Description:
 * @File: error
 * @Version: 1.0.0
 * @Date: 2024/10/24 10:34
 */

package api

import "errors"

var (
	ErrMailBoxNil              = errors.New("mailbox is nil")
	ErrActorStopped            = errors.New("actor is stopped")
	ErrActorRespondEnvIsNil    = errors.New("actor respond env is nil")
	ErrActorRespondSenderIsNil = errors.New("actor respond sender is nil")
	ErrActorCallTimeout        = errors.New("actor call timeout")
	ErrActorNameExist          = errors.New("actor name exist")
	ErrActorNameNotExist       = errors.New("actor name not exist")
	ErrActorNotLocalPid        = errors.New("actor call not local pid ")
	ErrActorRouterIsNil        = errors.New("actor router is nil")
	ErrDiscoveryProviderIsNil  = errors.New("discovery provider is nil")
	ErrPidIsNil                = errors.New("pid is nil")
	ErrInvalidPid              = errors.New("pid invalid")
	ErrNotLocalPid             = errors.New("not local pid")
	ErrProcessNotExist         = errors.New("process not exist")
	ErrActorArgsNum            = errors.New("actor args num error")
	ErrRpcArgsNum              = errors.New("rpc args num error")
)
