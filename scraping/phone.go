package scraping

import (
	"context"
	"errors"
	"log"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

var exists = errors.New("user doesn't exist")

// CheckPhoneNumber checks for the account associated with a phone number. It does this
// by adding the user has a contact
func CheckPhoneNumber(ctx context.Context, client *telegram.Client, phone string) (*tg.User, error) {
	user, err := AddContact(ctx, client, phone)
	if err != nil {
		return nil, err
	}
	err = DeleteContact(ctx, client, []tg.InputUserClass{user.AsInput()})
	if err != nil {
		return nil, err
	}
	return user, nil
}

// AddContact adds a contact to your contacts list
func AddContact(ctx context.Context, client *telegram.Client, phone string) (*tg.User, error) {
	contacts, err := client.API().ContactsImportContacts(ctx, []tg.InputPhoneContact{
		// TODO: random name & last name, in case there is a detection mechanism
		{ClientID: 0, Phone: phone, FirstName: "", LastName: ""},
	})
	if err != nil {
		return nil, err
	}
	for _, user := range contacts.Users {
		switch u := user.(type) {
		case *tg.User: // user#83314fca
			return u, nil
		default:
			log.Println("[INFO] Unknown case happened, maybe the user has deleted his account / been banned: ", phone)
			return nil, exists
		}
	}
	return nil, exists
}

// DeleteContact deletes the contact that has been imported in AddContact.
func DeleteContact(ctx context.Context, client *telegram.Client, id []tg.InputUserClass) error {
	_, err := client.API().ContactsDeleteContacts(ctx, id)
	if err != nil {
		return err
	}
	return nil
}
