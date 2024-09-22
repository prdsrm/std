package channels

import (
	"context"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"

	"github.com/prdsrm/std/messages"
)

type ChannelMonitoring struct {
	*messages.Monitoring
}

// NewChannelMonitoring creates a new `ChannelMonitoring` struct.
// NOTE: If the channel is private, activate the parameters to show ID of entities
// you're interacting with on Telegram Desktop, or use an alternative client like NekoGram on
// Android, or NiceGram on iOS. You should be subscribed to the channel to do so.
// NOTE: The strip parameter means that you want to remove special characters and spaces from
// messages you will be parsing, because you don't need them, and, its easier to make regular
// expressions this way.
func NewChannelMonitoring(ctx context.Context, client *telegram.Client, username string, dispatcher tg.UpdateDispatcher, strip bool) (*ChannelMonitoring, error) {
	sender := message.NewSender(client.API())
	builder := sender.Resolve(username)
	_, err := builder.Join(ctx)
	if err != nil {
		// NOTE: an error is returned if it's not a channel, so we don't need to check ourselves.
		return nil, err
	}
	channel, err := builder.AsInputChannel(ctx)
	if err != nil {
		return nil, err
	}
	monitoring := messages.NewMonitoring(dispatcher, channel.ChannelID, strip)
	return &ChannelMonitoring{Monitoring: monitoring}, nil
}
