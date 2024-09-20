package channels

import (
	"context"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

func GetSimilarChannels(ctx context.Context, client *telegram.Client, channelID int64, channelUsername string) ([]*tg.Channel, error) {
	var inputChannel tg.InputChannelClass
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

// TODO: export the channels to maltego, and set multiple backends for the Maltego definition:
// - Byte(Default)
// - Write to File
// func ExportChannelsToMaltego(mainChannel *tg.Channel, subChannels []*tg.Channel) ([]byte, error) {}
