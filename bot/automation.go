package bot

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
)

type Automation struct {
	ctx        context.Context
	client     *tg.Client
	dispatcher tg.UpdateDispatcher
	Username   string
	InputPeer  tg.InputPeerClass
}

// NewAutomation creates a new `Automation` object.
// It requires an username, since all (official) Telegram Bot have an username.
func NewAutomation(ctx context.Context, client *telegram.Client, dispatcher tg.UpdateDispatcher, username string) (*Automation, error) {
	automation := Automation{Username: username, ctx: ctx, client: client.API(), dispatcher: dispatcher}
	sender := message.NewSender(automation.client)
	builder := sender.Resolve(automation.Username)
	inputPeer, err := builder.AsInputPeer(ctx)
	if err != nil {
		return nil, err
	}
	automation.InputPeer = inputPeer
	return &automation, nil
}

// SendStartMessageWithParams will start a chat with a bot with the appropriate parameters.
// For instance, if you receive the link https://t.me/randombot?start=hello, the "hello" parameter
// will be sent as well.
func (a *Automation) SendStartMessageWithParams(params string) error {
	randomID := rand.Int63()
	_, err := a.client.MessagesStartBot(a.ctx, &tg.MessagesStartBotRequest{
		Bot:        &tg.InputUser{UserID: a.InputPeer.(*tg.InputPeerUser).UserID, AccessHash: a.InputPeer.(*tg.InputPeerUser).AccessHash},
		Peer:       a.InputPeer,
		RandomID:   randomID,
		StartParam: params,
	})
	if err != nil {
		return err
	}
	return nil
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

func (a *Automation) SetupMessageMonitoring(messagesChan chan *tg.Message) {
	a.dispatcher.OnEditMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateEditMessage) error {
		m, ok := u.Message.(*tg.Message)
		if !ok || m.Out {
			// Outgoing message, not interesting.
			return nil
		}
		messagesChan <- m
		return nil
	})
	a.dispatcher.OnNewMessage(func(ctx context.Context, entities tg.Entities, u *tg.UpdateNewMessage) error {
		m, ok := u.Message.(*tg.Message)
		if !ok || m.Out {
			// Outgoing message, not interesting.
			return nil
		}
		messagesChan <- m
		return nil
	})
	a.dispatcher.OnNewChannelMessage(func(ctx context.Context, entities tg.Entities, u *tg.UpdateNewChannelMessage) error {
		m, ok := u.Message.(*tg.Message)
		if !ok || m.Out {
			// Outgoing message, not interesting.
			return nil
		}
		messagesChan <- m
		return nil
	})
}
