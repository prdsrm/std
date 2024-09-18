package utils

import (
	"github.com/gotd/td/tg"
)

// GetIDFromPeerClass returns the chat/user id from the provided tg.PeerClass.
func GetIDFromPeerClass(peer tg.PeerClass) int64 {
	switch peer := peer.(type) {
	case *tg.PeerChannel:
		return peer.ChannelID
	case *tg.PeerUser:
		return peer.UserID
	case *tg.PeerChat:
		return peer.ChatID
	default:
		return 0
	}
}
