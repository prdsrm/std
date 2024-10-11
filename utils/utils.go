package utils

import (
	"errors"
	"fmt"
	"golang.org/x/net/proxy"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/gotd/td/telegram/dcs"
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

type MessageDeepLink struct {
	Username   string
	MessageID  int
	Parameters string
}

func ParseMessageDeepLink(link string) (MessageDeepLink, error) {
	parsedURL, err := url.Parse(link)
	if err != nil {
		return MessageDeepLink{}, fmt.Errorf("Couldn't parse message deep link %s: %w", link, err)
	}
	paths := strings.Split(parsedURL.Path, "/")
	if len(paths) >= 3 {
		username := paths[1]
		id, err := strconv.Atoi(paths[2])
		if err != nil {
			return MessageDeepLink{}, fmt.Errorf(
				"Message ID isn't an integer %s: %w",
				paths[2],
				err,
			)
		}
		return MessageDeepLink{Username: username, MessageID: id}, nil
	} else if len(paths) == 2 {
		username := paths[1]
		values := parsedURL.Query()
		startParams := values["start"][0]
		return MessageDeepLink{Username: username, MessageID: 0, Parameters: startParams}, nil
	} else {
		return MessageDeepLink{}, errors.New("can't parse")
	}
}

func RemoveSpacesAndNewlines(s string) string {
	// Use strings.Map to remove non-printable characters
	cleanText := strings.Map(func(r rune) rune {
		if r <= unicode.MaxASCII {
			return r
		}
		return -1
	}, s)
	return strings.Join(strings.Fields(cleanText), "")
}

func NewDialer(proxyConnStr string) (proxy.Dialer, error) {
	url, err := url.Parse(proxyConnStr)
	if err != nil {
		return nil, err
	}
	socks5, err := proxy.FromURL(url, proxy.Direct)
	if err != nil {
		return nil, err
	}
	return socks5, nil
}

func NewResolver(proxyConnStr string) (dcs.Resolver, error) {
	var resolver dcs.Resolver
	socks5, err := NewDialer(proxyConnStr)
	if err != nil {
		return nil, err
	}
	dc := socks5.(proxy.ContextDialer)
	resolver = dcs.Plain(dcs.PlainOptions{
		Dial: dc.DialContext,
	})
	return resolver, nil
}
