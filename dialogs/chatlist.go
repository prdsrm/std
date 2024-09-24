package dialogs

import (
	"context"
	"fmt"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

func JoinChatlist(ctx context.Context, client *telegram.Client, slug string) error {
	chatlistInvite, err := client.API().ChatlistsCheckChatlistInvite(ctx, slug)
	if err != nil {
		return fmt.Errorf("can't get chatlist %s: %w", slug, err)
	}
	switch invite := chatlistInvite.(type) {
	case *tg.ChatlistsChatlistInviteAlready:
		return nil
	case *tg.ChatlistsChatlistInvite:
		var inputPeers []tg.InputPeerClass
		for _, chat := range invite.GetChats() {
			switch channel := chat.(type) {
			case *tg.Channel:
				inputPeers = append(inputPeers, channel.AsInputPeer())
			}
		}

		_, err := client.API().ChatlistsJoinChatlistInvite(ctx, &tg.ChatlistsJoinChatlistInviteRequest{
			Slug:  slug,
			Peers: inputPeers,
		})
		if err != nil {
			return fmt.Errorf("can't join chatlist: %w", err)
		}
	}
	return nil
}

func ExportShareableFolder(ctx context.Context, client *telegram.Client, slug string, categoryID int) ([]tg.ChatClass, error) {
	chatlistInvite, err := client.API().ChatlistsCheckChatlistInvite(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("can't get chatlist %s: %w", slug, err)
	}
	// TODO: add some helpers for object like []tg.UserClass, []tg.ChatClass...
	return chatlistInvite.GetChats(), nil
}
