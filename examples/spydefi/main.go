package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"

	"github.com/prdsrm/std/bot"
	"github.com/prdsrm/std/messages"
	"github.com/prdsrm/std/session"
	"github.com/prdsrm/std/utils"
)

var expr = `(?m)Statsfor\@([a-zA-Z0-9_]{4,32}):Consistency:([0-9.]{1,3})%NumberofAlphaCalls:([0-9]+)BestCall:([a-zA-Z0-9.]+)(\(x[0-9.]+\))LastCall:([a-zA-Z0-9]+)(\(x[0-9.]+\))AverageCallMarketcap:\$([0-9.]+)(K|M|B)AverageXPerCall:x([0-9.]+)TotalCallsTracked:(\d+)NumberofAchievementCalls:calls:(\d+)calls:(\d+)TotalAchievements:(\d+)AchievementCallsrefertocallsthathavegonex2\+asperourparameters`
var re = regexp.MustCompile(expr)

func saveSpydefiChannelData(ctx messages.MonitoringContext) error {
	cleanedMessage := utils.RemoveSpacesAndNewlines(ctx.GetMessage().Message)
	for _, match := range re.FindAllStringSubmatch(cleanedMessage, -1) {
		username := match[1]
		averageCallMC := match[8]
		numberMC, err := strconv.ParseFloat(averageCallMC, 2)
		if err != nil {
			return err
		}
		mc := int(numberMC)
		averageCallMCUnit := match[9]
		var numberAverageCallMC int
		if averageCallMCUnit == "M" {
			numberAverageCallMC = mc * 1_000_000
		}
		if averageCallMCUnit == "K" {
			numberAverageCallMC = mc * 1_000
		} else {
			numberAverageCallMC = mc * 1_000_000_000
		}
		averageCallX := match[10]
		numberAverageCallX, _ := strconv.ParseFloat(averageCallX, 2)

		// NOTE: starting from here, you could start saving the result in a database or something similar.
		fmt.Println("Data: ", username, numberAverageCallMC, numberAverageCallX)
		return messages.EndConversation
	}
	if cleanedMessage == "ThisKOLChannelwasnotfoundinournetworkMakesureyouhaveenteredtheusernamecorrectly!Youcanalsousethe/trackme<@username>featuretosubmitanychannelstobetracked." {
		fmt.Println("This channel does not exist, in the spydefi database.")
		return messages.EndConversation
	}
	return nil
}

func listen(ctx context.Context, client *telegram.Client, dispatcher tg.UpdateDispatcher, options telegram.Options) error {
	fmt.Println("Successfully connected to the bot")
	automation, err := bot.NewAutomation(ctx, client, dispatcher, "spydefi_bot", true)
	if err != nil {
		return err
	}
	// Here is an example, on how to use custom TG parameters to start the bot, like for this URL:
	// https://t.me/SpyDefi_bot?start=tggemarcheologisttelegram-1001942713434
	// This channel is fully arbitrary, I'm not a crypto trader, I just do programming, I selected
	// it from the t.me/spydefi channel, randomly.
	fmt.Println("Sending start message with params")
	err = automation.SendStartMessageWithParams("tggemarcheologisttelegram-1001942713434")
	if err != nil {
		return err
	}
	// NOTE: you can also do it like this:
	// username := "LeclercCalls"
	// // ^ same, this channel has been taken randomly, from my scraped database, because it has some good "KOL".
	// err = automation.SendTextMessage("/start")
	// err = automation.SendTextMessage(fmt.Sprintf("/checkstats @%s", username))
	// if err != nil {
	// 	return err
	// }
	automation.Handle(re, saveSpydefiChannelData)
	err = automation.Listen(ctx, client)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	phone := os.Getenv("PHONE_NUMBER")
	password := os.Getenv("PASSWORD")
	sessionString := os.Getenv("SESSION_STRING")
	flow := session.GetNewDefaultAuthConversator(phone, password)
	err := session.Connect(listen, session.Windows(), 2040, "b18441a1ff607e10a989891a5462e627", sessionString, "", flow)
	if err != nil {
		log.Fatalln("can't connect: ", err)
	}
}
