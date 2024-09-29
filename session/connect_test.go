package session

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"testing"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"

	"github.com/prdsrm/std/bot"
	"github.com/prdsrm/std/channels"
	"github.com/prdsrm/std/messages"
)

func callSelf(ctx context.Context, client *telegram.Client, dispatcher tg.UpdateDispatcher, options telegram.Options) error {
	self, err := client.Self(ctx)
	if err != nil {
		return err
	}
	log.Println("Self", self.ID, self.Username)
	automation, err := bot.NewAutomation(ctx, client, dispatcher, "tgdb_bot", false)
	if err != nil {
		return err
	}
	err = automation.SendTextMessage("/resolve 7513073974")
	if err != nil {
		return err
	}
	automation.Handle(regexp.MustCompile(".*"), defaultHandler)
	err = automation.Listen()
	if err != nil {
		return err
	}
	channels, err := channels.GetSimilarChannels(ctx, client, "durov")
	if err != nil {
		return err
	}
	for _, channel := range channels {
		fmt.Println("Channel: ", channel.ID, channel.Username)
	}
	return nil
}

func defaultHandler(ctx messages.MonitoringContext) error {
	log.Println(ctx.GetMessage().Message)
	return messages.EndConversation
}

func TestConnect(t *testing.T) {
	phone := os.Getenv("PHONE_NUMBER")
	password := os.Getenv("PASSWORD")
	sessionString := os.Getenv("SESSION_STRING")
	flow := GetNewDefaultAuthConversator(phone, password)
	err := Connect(callSelf, Windows(), 2040, "b18441a1ff607e10a989891a5462e627", sessionString, "", flow)
	if err != nil {
		t.Fatalf("can't continue: %s", err.Error())
	}
}
