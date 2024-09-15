package bot

import (
	"context"
	"fmt"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
)

type Automation struct {
	ctx       context.Context
	client    *tg.Client
	Username  string
	InputPeer tg.InputPeerClass
}

// NewAutomation creates a new `Automation` object.
// It requires an username, since all (official) Telegram Bot have an username.
func NewAutomation(ctx context.Context, client *telegram.Client, username string) (*Automation, error) {
	automation := Automation{Username: username, ctx: ctx, client: client.API()}
	sender := message.NewSender(automation.client)
	builder := sender.Resolve(automation.Username)
	inputPeer, err := builder.AsInputPeer(ctx)
	if err != nil {
		return nil, err
	}
	automation.InputPeer = inputPeer
	return &automation, nil
}

func (a *Automation) SendTextMessage(text string) error {
	sender := message.NewSender(a.client)
	builder := sender.To(a.InputPeer)
	_, err := builder.Text(a.ctx, text)
	if err != nil {
		return fmt.Errorf("couldn't send message to bot: %w.", err)
	}
	return nil
}

func (a *Automation) SendCallbackData(msgID int, callbackData string) (*tg.MessagesBotCallbackAnswer, error) {
	return a.client.MessagesGetBotCallbackAnswer(a.ctx, &tg.MessagesGetBotCallbackAnswerRequest{
		Game:  false,
		Peer:  a.InputPeer,
		MsgID: msgID,
		Data:  []byte(callbackData),
	})
}
