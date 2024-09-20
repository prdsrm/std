package bot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"regexp"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/prdsrm/std/utils"
)

type Automation struct {
	ctx        context.Context
	client     *tg.Client
	dispatcher tg.UpdateDispatcher
	Username   string
	ID         int64
	InputPeer  tg.InputPeerClass
	routes     []RouteEntry
	strip      bool
}

// NewAutomation creates a new `Automation` object.
// It requires an username, since all (official) Telegram Bot have an username.
func NewAutomation(ctx context.Context, client *telegram.Client, dispatcher tg.UpdateDispatcher, username string, strip bool) (*Automation, error) {
	automation := Automation{Username: username, ctx: ctx, client: client.API(), dispatcher: dispatcher}
	sender := message.NewSender(automation.client)
	builder := sender.Resolve(automation.Username)
	inputPeer, err := builder.AsInputPeer(ctx)
	if err != nil {
		return nil, err
	}
	automation.InputPeer = inputPeer
	automation.ID = inputPeer.(*tg.InputPeerUser).UserID
	automation.strip = false
	if strip {
		automation.strip = true
	}
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

type AutomationContext struct {
	m *tg.Message
}

func (a AutomationContext) GetMessage() *tg.Message {
	return a.m
}

var (
	EndConversation = errors.New("end")
)

type RouteEntry struct {
	Regex   *regexp.Regexp
	Handler func(ctx AutomationContext) error
}

func (a *Automation) Handle(expr string, handler func(ctx AutomationContext) error) {
	regex, err := regexp.Compile(expr)
	if err != nil {
		log.Fatalf("Regex expression: `%s` is invalid and failed to compile: %w\n", expr, err)
	}
	a.routes = append(a.routes, RouteEntry{Regex: regex, Handler: handler})
}

func (a *Automation) Listen() error {
	messagesChan := make(chan *tg.Message)
	a.SetupMessageMonitoring(messagesChan)
	for {
		msg := <-messagesChan
		id := utils.GetIDFromPeerClass(msg.PeerID)
		if id == a.ID {
			ctx := AutomationContext{
				m: msg,
			}
			text := msg.Message
			if a.strip {
				text = utils.RemoveSpacesAndNewlines(text)
			}
			for _, route := range a.routes {
				exists := route.Regex.Match([]byte(text))
				if exists {
					err := route.Handler(ctx)
					if err == EndConversation {
						log.Println("[INFO] Shutting down...")
						return nil
					}
					if err != nil {
						log.Println("[ERROR] while processing request: ", err)
					}
					break
				}
			}
		}
	}
}
