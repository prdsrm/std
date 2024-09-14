package session

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/gotd/td/telegram"
)

func callSelf(ctx context.Context, client *telegram.Client) error {
	self, err := client.Self(ctx)
	if err != nil {
		return err
	}
	log.Println("Self", self.ID, self.Username)
	return nil
}

func TestConnect(t *testing.T) {
	phone := os.Getenv("PHONE_NUMBER")
	password := os.Getenv("PASSWORD")
	sessionString := os.Getenv("SESSION_STRING")
	err := Connect(callSelf, phone, password, Windows(), 2040, "b18441a1ff607e10a989891a5462e627", sessionString, "")
	if err != nil {
		t.Fatalf("can't continue: %s", err.Error())
	}
}
