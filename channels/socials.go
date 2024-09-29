package channels

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"

	"github.com/prdsrm/std/utils"
)

// JoinChannel joins a public channel / group chat, that has an username.
func JoinChannel(ctx context.Context, client *telegram.Client, username string) error {
	sender := message.NewSender(client.API())
	builder := sender.Resolve(username)
	_, err := builder.Join(ctx)
	if err != nil {
		return err
	}
	return nil
}

func JoinPrivateChannel(ctx context.Context, client *telegram.Client, hash string) error {
	_, err := client.API().MessagesImportChatInvite(ctx, hash)
	if err != nil && !strings.Contains(err.Error(), "INVITE_REQUEST_SENT") {
		return fmt.Errorf("can't import chat invite: %w", err)
	}
	return nil
}

// AddView increments the number of view of a channel post.
// It accepts a message deep link, such as: https://t.me/channel_username/<message_id: number>
func AddView(ctx context.Context, client *telegram.Client, link string) error {
	messageDeepLink, err := utils.ParseMessageDeepLink(link)
	if err != nil {
		return err
	}
	sender := message.NewSender(client.API())
	inputPeer, err := sender.Resolve(messageDeepLink.Username).AsInputPeer(ctx)
	if err != nil {
		return fmt.Errorf("couldn't resolve channel %s: %w", messageDeepLink.Username, err)
	}
	views, err := client.API().MessagesGetMessagesViews(ctx, &tg.MessagesGetMessagesViewsRequest{
		Peer:      inputPeer,
		ID:        []int{messageDeepLink.MessageID},
		Increment: true,
	})
	for _, view := range views.GetViews() {
		fmt.Println("Current number of views: ", view.Views)
	}
	return nil
}

func toReactionClass(something tg.ReactionClass) tg.ReactionClass {
	return something
}

// AddReaction adds a reaction as well as a view to the current post.
// It accepts a message deep link, such as: https://t.me/channel_username/<message_id: number>
func AddReaction(ctx context.Context, client *telegram.Client, link string, emojiChar string) error {
	messageDeepLink, err := utils.ParseMessageDeepLink(link)
	if err != nil {
		return err
	}
	sender := message.NewSender(client.API())
	inputPeer, err := sender.Resolve(messageDeepLink.Username).AsInputPeer(ctx)
	if err != nil {
		return err
	}
	// NOTE: only adding a reaction without viewing might trigger some weird behavior.
	views, err := client.API().MessagesGetMessagesViews(ctx, &tg.MessagesGetMessagesViewsRequest{
		Peer:      inputPeer,
		ID:        []int{messageDeepLink.MessageID},
		Increment: true,
	})
	for _, view := range views.GetViews() {
		fmt.Println("Current number of views: ", view.Views)
	}
	var reactions []tg.ReactionClass
	emoji := toReactionClass(&tg.ReactionEmoji{Emoticon: emojiChar})
	reactions = append(reactions, emoji)
	_, err = client.API().MessagesSendReaction(ctx, &tg.MessagesSendReactionRequest{
		Peer:     inputPeer,
		MsgID:    messageDeepLink.MessageID,
		Reaction: reactions,
	})
	if err != nil {
		return err
	}
	return nil
}

// ForwardMessageFromPublicChannel forwards a message from a public channel thanks to the message deep link
func ForwardMessageFromPublicChannel(ctx context.Context, client *telegram.Client, deepLink string, destinationChannelUsername string) error {
	messageDeepLink, err := utils.ParseMessageDeepLink(deepLink)
	if err != nil {
		return err
	}
	sender := message.NewSender(client.API())
	inputPeer, err := sender.Resolve(messageDeepLink.Username).AsInputPeer(ctx)
	if err != nil {
		return fmt.Errorf("can't get as input peer: %w", err)
	}
	builder := sender.Resolve(destinationChannelUsername)
	_, err = builder.Join(ctx)
	if err != nil {
		return fmt.Errorf("can't join channel %s: %w", destinationChannelUsername, err)
	}
	_, err = builder.ForwardIDs(inputPeer, messageDeepLink.MessageID).Send(ctx)
	if err != nil {
		return fmt.Errorf("can't forward message to %s: %w", destinationChannelUsername, err)
	}
	return nil
}
