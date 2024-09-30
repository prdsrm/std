# Phronesis

[![Nix](https://img.shields.io/badge/Nix-5277C3?logo=nixos&logoColor=fff)](#)
[![Go](https://img.shields.io/badge/Go-%2300ADD8.svg?&logo=go&logoColor=white)](#)
[![Postgres](https://img.shields.io/badge/Postgres-%23316192.svg?logo=postgresql&logoColor=white)](#)


Telegram helper library and tools, tailored for bot & channel automation, groups or channel monitoring, managing sessions and scraping.

## Features

- Session helpers for [gotd](https://github.com/gotd/td)
  - allows converting a Telethon SQLite Session, or TDATA session, to a Telethon string session
  - export all of your sessions to Telethon string session
  - **Examples**: [simply generate a telethon string session for your account, with manual login, or an existing SQLite / TDATA session](https://github.com/prdsrm/std/blob/main/cmd/generate/main.go)
- Telegram bot automation helpers
  - Structured object in order to automate Telegram bot.
  - **Examples**: scrape data from [SpyDefi Bot](https://docs.spydefi.org/spydefi-docs/spydefi-guides/how-to-use-spydefi/spydefi-bot), [check out the code](https://github.com/prdsrm/std/blob/main/examples/spydefi/main.go).
- Group and channel monitoring / scraping
  - Custom object tailored for parsing message from channels.
  - **Examples**: monitor newly listed tokens on [DexScreener](https://dexscreener.com), [check out the code](https://github.com/prdsrm/std/blob/main/examples/dslisting/main.go).
- Common helpers for [gotd](https://github.com/gotd/td)
  - Create shareable folders(<https://t.me/addlist/random>), export chats in them, join them.
	- Scrape similar channels, export them into a Maltego file
  - Join channel, add views, reactions, create channels and groups, promote & demote members, add members to your channel, easily.
  - Automate posting in your channel.
  - Export all members from a channel you own.

## Installation

Run the following in your Go project:
```bash
go get github.com/prdsrm/std
```

## Usage

Scrape similar channels example, (here, we get the channels similar to Pavel Durov personal channel):
```go

package main

import (
	"context"
	"fmt"
	"log"

	examples "github.com/gotd/td/examples"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"

	"github.com/prdsrm/std/channels"
	"github.com/prdsrm/std/session"
)

func run(ctx context.Context, client *telegram.Client, dispatcher tg.UpdateDispatcher, options telegram.Options) error {
	channels, err := channels.GetSimilarChannels(ctx, client, "durov")
	if err != nil {
		return err
	}
	for _, channel := range channels {
		fmt.Println("Channel: ", channel.ID, channel.Username)
	}
	return nil
}

func main() {
	authOpt := auth.SendCodeOptions{}
	flow := auth.NewFlow(examples.Terminal{}, authOpt)
	err := session.Connect(run, session.Windows(), 2040, "b18441a1ff607e10a989891a5462e627", "", "", flow)
	if err != nil {
		log.Fatalln("can't connect: ", err)
	}
}
```
You can run this example, just create a `main.go` file, and install the library with `go get`, then run `go run`.

Bot automation example, extracting KOLs data from <https://t.me/spydefi_bot>:

```go
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
	err = automation.Listen()
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
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first
to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

BSD-3-Clause license, because FreeBSD & OpenBSD is great
