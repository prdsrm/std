package messages

import (
	"context"
	"errors"
	"log"
	"regexp"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"

	"github.com/prdsrm/std/utils"
)

type Monitoring struct {
	dispatcher tg.UpdateDispatcher
	id         int64
	routes     []RouteEntry
	strip      bool
}

func NewMonitoring(dispatcher tg.UpdateDispatcher, id int64, strip bool) *Monitoring {
	return &Monitoring{dispatcher: dispatcher, id: id, strip: strip}

}

func (a *Monitoring) SetupMessageMonitoring(messagesChan chan *tg.Message) {
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

type MonitoringContext struct {
	Ctx context.Context
	c   *telegram.Client
	m   *tg.Message
}

func (m *MonitoringContext) GetClient() *telegram.Client {
	return m.c
}

func (m *MonitoringContext) GetMessage() *tg.Message {
	return m.m
}

var (
	EndConversation = errors.New("end")
)

type RouteEntry struct {
	Regex   *regexp.Regexp
	Handler func(ctx MonitoringContext) error
}

func (m *Monitoring) Handle(re *regexp.Regexp, handler func(ctx MonitoringContext) error) {
	m.routes = append(m.routes, RouteEntry{Regex: re, Handler: handler})
}

func (m *Monitoring) Listen(ctx context.Context, client *telegram.Client) error {
	messagesChan := make(chan *tg.Message)
	m.SetupMessageMonitoring(messagesChan)
	for {
		msg := <-messagesChan
		id := utils.GetIDFromPeerClass(msg.PeerID)
		if id == m.id {
			ctx := MonitoringContext{
				Ctx: ctx,
				m:   msg,
				c:   client,
			}
			text := msg.Message
			if m.strip {
				text = utils.RemoveSpacesAndNewlines(text)
			}
			for _, route := range m.routes {
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
				log.Println("Received message but it does not match with our filter: ", text)
			}
		}
	}
}

// MonitorMessages accepts a function to process the message along with the chatID of the channel
// from which it was received.
// This function is deprecated please use the `Monitoring` class for simplicity.
func MonitorMessages(ctx context.Context, client *telegram.Client, dispatcher tg.UpdateDispatcher, work func(chatID int64, m *tg.Message) error, done chan bool) {
	dispatcher.OnEditMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateEditMessage) error {
		switch m := update.Message.(type) {
		case *tg.Message:
			chatID := utils.GetIDFromPeerClass(m.PeerID)
			err := work(chatID, m)
			if err != nil {
				return err
			}
		}
		return nil
	})
	dispatcher.OnNewMessage(func(ctx context.Context, entities tg.Entities, u *tg.UpdateNewMessage) error {
		switch m := u.Message.(type) {
		case *tg.Message: // message#2357bf25
			m, ok := u.Message.(*tg.Message)
			if !ok || m.Out {
				// Outgoing message, not interesting.
				return nil
			}
			chatID := utils.GetIDFromPeerClass(m.PeerID)
			err := work(chatID, m)
			if err != nil {
				return err
			}
		}
		return nil
	})
	dispatcher.OnNewChannelMessage(func(ctx context.Context, entities tg.Entities, u *tg.UpdateNewChannelMessage) error {
		switch m := u.Message.(type) {
		case *tg.MessageEmpty: // messageEmpty#90a6ca84
		case *tg.Message: // message#2357bf25
			if m.Out {
				// Outgoing message, not interesting.
				return nil
			}
			chatID := utils.GetIDFromPeerClass(m.PeerID)
			err := work(chatID, m)
			if err != nil {
				return err
			}
		case *tg.MessageService: // messageService#2b085862
		}

		return nil
	})
	<-done
}
