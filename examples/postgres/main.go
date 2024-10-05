package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"

	"github.com/prdsrm/std/session/postgres"
)

func listen(ctx context.Context, client *telegram.Client, dispatcher tg.UpdateDispatcher, options telegram.Options) error {
	user, err := client.Self(ctx)
	if err != nil {
		return err
	}
	log.Println("Connected to user account: ", user.ID, user.FirstName)
	return nil
}

func main() {
	userID := os.Getenv("USER_ID")
	id, err := strconv.Atoi(userID)
	if err != nil {
		log.Fatalln("user ID is not an integer: ", err)
	}
	db, err := postgres.OpenDBConnection()
	if err != nil {
		log.Fatalln("can't connect to database: ", err)
	}
	bot, err := postgres.GetBotByUserID(db, int64(id))
	if err != nil {
		log.Fatalln("can't get bot: ", err)
	}
	log.Println("Bot from the db: ", bot.UserID)
	err = postgres.ConnectToBotFromDatabase(db, bot, listen)
	if err != nil {
		log.Fatalln("can't connect: ", err)
	}
}
