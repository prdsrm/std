package channels

import (
	"context"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

func GetUpdateFromUpdates(updates tg.UpdatesClass) ([]tg.UpdateClass, []tg.ChatClass, []tg.UserClass) {
	switch u := updates.(type) {
	case *tg.Updates:
		return u.Updates, u.Chats, u.Users
	case *tg.UpdatesCombined:
		return u.Updates, u.Chats, u.Users
	case *tg.UpdateShort:
		return []tg.UpdateClass{u.Update}, tg.ChatClassArray{}, tg.UserClassArray{}
	default:
		return nil, nil, nil
	}
}

// The following code is taken from https://github.com/celestix/gotgproto
// Check it out if you like wrappers. It can also be used for official bots.
// I just edited it in order to use standard `td` objects.

func CreateChannel(ctx context.Context, client *telegram.Client, title string, about string, broadcast bool) (*tg.Channel, error) {
	udps, err := client.API().ChannelsCreateChannel(ctx, &tg.ChannelsCreateChannelRequest{
		Title:     title,
		About:     about,
		Broadcast: broadcast,
	})
	if err != nil {
		return nil, err
	}
	// Highly experimental value from ChatClass array
	_, chats, _ := GetUpdateFromUpdates(udps)
	return chats[0].(*tg.Channel), nil
}

func CreateChat(ctx context.Context, client *telegram.Client, title string, users []tg.InputUserClass) (*tg.Chat, error) {
	udps, err := client.API().MessagesCreateChat(ctx, &tg.MessagesCreateChatRequest{
		Users: users,
		Title: title,
	})
	if err != nil {
		return nil, err
	}
	// Highly experimental value from ChatClass map
	_, chats, _ := GetUpdateFromUpdates(udps.Updates)
	return chats[0].(*tg.Chat), nil
}

func AddChatMembers(ctx context.Context, client *telegram.Client, chatPeer tg.InputPeerClass, users []tg.InputUserClass, forwardLimit int) (bool, error) {
	switch c := chatPeer.(type) {
	case *tg.InputPeerChat:
		for _, user := range users {
			user, ok := user.(*tg.InputUser)
			if ok {
				_, err := client.API().MessagesAddChatUser(ctx, &tg.MessagesAddChatUserRequest{
					ChatID: c.ChatID,
					UserID: &tg.InputUser{
						UserID:     user.UserID,
						AccessHash: user.AccessHash,
					},
					FwdLimit: forwardLimit,
				})
				if err != nil {
					return false, err
				}
			}
		}
		return true, nil
	case *tg.InputPeerChannel:
		_, err := client.API().ChannelsInviteToChannel(ctx, &tg.ChannelsInviteToChannelRequest{
			Channel: &tg.InputChannel{
				ChannelID:  c.ChannelID,
				AccessHash: c.AccessHash,
			},
			Users: users,
		})
		return err == nil, err
	}
	return false, nil
}

func PromoteChatMember(ctx context.Context, client *telegram.Client, chat *tg.InputChannel, user *tg.InputUser, rights tg.ChatAdminRights, title string) (bool, error) {
	rights.Other = true
	if chat.AccessHash != 0 {
		_, err := client.API().ChannelsEditAdmin(ctx, &tg.ChannelsEditAdminRequest{
			Channel:     chat,
			UserID:      user,
			AdminRights: rights,
			Rank:        title,
		})
		return err == nil, err
	} else {
		_, err := client.API().MessagesEditChatAdmin(ctx, &tg.MessagesEditChatAdminRequest{
			ChatID:  chat.ChannelID,
			UserID:  user,
			IsAdmin: true,
		})
		return err == nil, err
	}
}

func DemoteChatMember(ctx context.Context, client *telegram.Client, chat *tg.InputChannel, user *tg.InputUser, rights tg.ChatAdminRights, title string) (bool, error) {
	rights.Other = false
	if chat.AccessHash != 0 {
		_, err := client.API().ChannelsEditAdmin(ctx, &tg.ChannelsEditAdminRequest{
			Channel:     chat,
			UserID:      user,
			AdminRights: rights,
			Rank:        title,
		})
		return err == nil, err
	} else {
		_, err := client.API().MessagesEditChatAdmin(ctx, &tg.MessagesEditChatAdminRequest{
			ChatID:  chat.ChannelID,
			UserID:  user,
			IsAdmin: false,
		})
		return err == nil, err
	}
}
