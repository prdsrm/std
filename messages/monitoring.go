package messages

import (
	"context"
	"errors"
	"log"
	"regexp"

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
	m *tg.Message
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

func (m *Monitoring) Handle(expr string, handler func(ctx MonitoringContext) error) {
	regex, err := regexp.Compile(expr)
	if err != nil {
		log.Fatalf("Regex expression: `%s` is invalid and failed to compile: %s\n", expr, err)
	}
	m.routes = append(m.routes, RouteEntry{Regex: regex, Handler: handler})
}

func (m *Monitoring) Listen() error {
	messagesChan := make(chan *tg.Message)
	m.SetupMessageMonitoring(messagesChan)
	for {
		msg := <-messagesChan
		id := utils.GetIDFromPeerClass(msg.PeerID)
		if id == m.id {
			ctx := MonitoringContext{
				m: msg,
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
			}
		}
	}
}
