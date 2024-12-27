/**
 * @Author: dingQingHui
 * @Description:
 * @File: BlueprintOptionsFunc
 * @Version: 1.0.0
 * @Date: 2024/10/24 10:16
 */

package actor

import (
	"github.com/dingqinghui/gas/api"
)

func loadOptions(options ...api.ProcessOption) *api.ActorProcessOptions {
	opts := new(api.ActorProcessOptions)
	for _, option := range options {
		option(opts)
	}
	return opts
}

func getDispatcher(b *api.ActorProcessOptions) api.IActorDispatcher {
	if b.Dispatcher == nil {
		b.Dispatcher = NewDefaultDispatcher(50)
	}
	return b.Dispatcher
}

func getMailBox(b *api.ActorProcessOptions) api.IActorMailbox {
	if b.Mailbox == nil {
		b.Mailbox = NewMailbox()
	}
	return b.Mailbox
}
