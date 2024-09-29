package main

import (
	"context"
	"fmt"
	"log"

	examples "github.com/gotd/td/examples"
	tdsession "github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"

	"github.com/prdsrm/std/session"
)

func listen(ctx context.Context, client *telegram.Client, dispatcher tg.UpdateDispatcher, options telegram.Options) error {
	loader := &tdsession.Loader{Storage: options.SessionStorage}
	data, err := loader.Load(ctx)
	if err != nil {
		return err
	}
	sessionString, err := session.EncodeSessionToTelethonString(data)
	if err != nil {
		return err
	}
	fmt.Println("Your telethon session string: ", sessionString)
	return nil
}

func main() {
	authOpt := auth.SendCodeOptions{}
	flow := auth.NewFlow(examples.Terminal{}, authOpt)
	err := session.Connect(listen, session.Windows(), 2040, "b18441a1ff607e10a989891a5462e627", "", "", flow)
	if err != nil {
		log.Fatalln("can't connect: ", err)
	}
}
