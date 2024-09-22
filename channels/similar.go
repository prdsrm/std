package channels

import (
	"context"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
)

func GetSimilarChannels(ctx context.Context, client *telegram.Client, channelUsername string) ([]*tg.Channel, error) {
	sender := message.NewSender(client.API())
	builder := sender.Resolve(channelUsername)
	inputChannel, err := builder.AsInputChannelClass(ctx)
	if err != nil {
		return nil, err
	}
	msgs, err := client.API().ChannelsGetChannelRecommendations(ctx, &tg.ChannelsGetChannelRecommendationsRequest{
		Channel: inputChannel,
	})
	if err != nil {
		return nil, err
	}
	var channels []*tg.Channel
	for _, chat := range msgs.GetChats() {
		switch channel := chat.(type) {
		case *tg.Channel:
			channels = append(channels, channel)
		}
	}
	return channels, nil
}
