package main

import (
	"context"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"

	"github.com/prdsrm/std/channels"
	"github.com/prdsrm/std/messages"
	"github.com/prdsrm/std/session"
	"github.com/prdsrm/std/utils"
)

type DSNewPair struct {
	Chain        string
	Address      string
	PumpFun      bool
	PriceUSD     string
	PriceSOL     string
	FDV          string
	Liquidity    string
	Ticker       string
	PooledToken  string
	PooledSOL    string
	PooledSOLUSD string
}

var re = regexp.MustCompile(`(?m)NewPaironSolana:.+\/(SOL).+Tokenaddress:([1-9A-HJ-NP-Za-km-z]{32,44})PriceUSD:\$([0-9.,]+)Price:([0-9.,]+)SOLFDV:\$([0-9.,]+)Totalliquidity:\$([0-9.,]+)Pooled($|)([A-Z]+):([0-9.,]+)PooledSOL:([0-9.,]+)\(\$([0-9.,]+)\)WARNING:buyingnewpairsisextremelyrisky,pleasereadpinnedmessagebeforeproceeding!`)
var counter = 0

func ParseDSNewPair(message messages.MonitoringContext) error {
	text := utils.RemoveSpacesAndNewlines(message.GetMessage().Message)
	for _, match := range re.FindAllStringSubmatch(text, -1) {
		chain := match[1]
		address := match[2]
		pumpfun := false
		if strings.Contains(address, "pump") {
			pumpfun = true
		}
		priceUSD := match[3]
		priceSOL := match[4]
		fdv := match[5]
		liquidity := match[6]
		ticker := match[8]
		pooledToken := match[9]
		pooledSOL := match[10]
		pooledSOLUSD := match[11]

		pair := DSNewPair{
			Chain:        chain,
			Address:      address,
			PumpFun:      pumpfun,
			PriceUSD:     priceUSD,
			PriceSOL:     priceSOL,
			FDV:          fdv,
			Liquidity:    liquidity,
			Ticker:       ticker,
			PooledToken:  pooledToken,
			PooledSOL:    pooledSOL,
			PooledSOLUSD: pooledSOLUSD,
		}
		log.Println("New Dexscreener listing on Solana: ", pair)
	}
	// We do not run this indefinitely, because this is an example.
	if counter > 5 {
		return messages.EndConversation
	}
	counter += 1
	return nil
}

func listen(ctx context.Context, client *telegram.Client, dispatcher tg.UpdateDispatcher, options telegram.Options) error {
	log.Println("Successfully connected to Telegram account")
	monitoring, err := channels.NewChannelMonitoring(ctx, client, "DSNewPairsSolana", dispatcher, true)
	if err != nil {
		return err
	}
	monitoring.Handle(re, ParseDSNewPair)
	log.Println("Starting to listen for new dexscreener listings")
	err = monitoring.Listen()
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
