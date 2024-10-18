package session

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/time/rate"
	"log"
	"sync"
	"time"

	"github.com/gotd/td/session"
	"github.com/prdsrm/std/utils"

	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/telegram/updates"
	updhook "github.com/gotd/td/telegram/updates/hook"
	"github.com/gotd/td/tg"
)

// MemorySession implements in-memory session storage.
// Goroutine-safe.
type MemorySession struct {
	mux  sync.RWMutex
	data []byte
}

// LoadSession loads session from memory.
func (s *MemorySession) LoadSession(context.Context) ([]byte, error) {
	if s == nil {
		return nil, session.ErrNotFound
	}

	s.mux.RLock()
	defer s.mux.RUnlock()

	if len(s.data) == 0 {
		return nil, session.ErrNotFound
	}

	cpy := append([]byte(nil), s.data...)

	return cpy, nil
}

// StoreSession stores session to memory.
func (s *MemorySession) StoreSession(ctx context.Context, data []byte) error {
	s.mux.Lock()
	s.data = data
	s.mux.Unlock()
	return nil
}

func GetNewDefaultAuthConversator(phone string, password string) auth.Flow {
	userAuthenticator := DefaultAuthConversator{PhoneNumber: phone, Passwd: password}
	authOpt := auth.SendCodeOptions{}
	// Authentication flow handles authentication process, like prompting for code and 2FA password.
	flow := auth.NewFlow(userAuthenticator, authOpt)
	return flow
}

func Connect(f func(ctx context.Context, client *telegram.Client, dispatcher tg.UpdateDispatcher, options telegram.Options) error, device telegram.DeviceConfig, apiID int, apiHash string, sessionString string, proxy string, flow auth.Flow) error {
	storage := &MemorySession{}
	// We only load the session if it isn't empty
	if sessionString != "" {
		loader := session.Loader{Storage: storage}
		// Extracts session data from Telethon session string.
		data, err := session.TelethonSession(sessionString)
		if err != nil {
			return err
		}
		// Save decoded Telethon session as gotd session.
		if err := loader.Save(context.Background(), data); err != nil {
			return err
		}
	}
	// Load proxy
	var resolver dcs.Resolver
	var err error
	if proxy != "" {
		resolver, err = utils.NewResolver(proxy)
		if err != nil {
			return err
		}
	}
	// Finish setting up options
	options := telegram.Options{
		Resolver:       resolver,
		SessionStorage: storage,
		Device:         device,
	}
	waiter := floodwait.NewWaiter()
	// Dispatcher handles incoming updates.
	dispatcher := tg.NewUpdateDispatcher()
	gaps := updates.New(updates.Config{
		Handler: dispatcher,
	})
	options.Middlewares = []telegram.Middleware{
		// Setting up FLOOD_WAIT handler to automatically wait and retry request.
		waiter,
		// Setting up general rate limits to less likely get flood wait errors.
		ratelimit.New(rate.Every(time.Millisecond*100), 5),
		// Setting up update hook
		updhook.UpdateHook(gaps.Handle),
	}
	options.UpdateHandler = gaps
	client := telegram.NewClient(apiID, apiHash, options)

	ctx := context.Background()
	if err := waiter.Run(ctx, func(ctx context.Context) error {
		// Spawning main goroutine.
		return runClient(f, ctx, client, dispatcher, options, flow)
	}); err != nil {
		return err
	}
	return nil
}

func runClient(f func(ctx context.Context, client *telegram.Client, dispatcher tg.UpdateDispatcher, options telegram.Options) error, ctx context.Context, client *telegram.Client, dispatcher tg.UpdateDispatcher, options telegram.Options, flow auth.Flow) error {
	if err := client.Run(ctx, func(ctx context.Context) error {
		authCli := client.Auth()
		// Checking auth status.
		status, err := authCli.Status(ctx)
		if err != nil {
			return err
		}
		// Can be already authenticated if we have valid session in
		// session storage.
		if !status.Authorized {
			if err := client.Auth().IfNecessary(ctx, flow); err != nil {
				return fmt.Errorf("could not authenticate: %w", err)
			}
			// In the telegram/connect.go, line 31, we can see that the client.Run helper does not correctly check
			// for the session authorization.
			// It doesn't return anything if its unauthorized, it doesn't log, because its in a goroutine. We need to check ourselves.
			self, err := client.Self(ctx)
			if err != nil {
				if auth.IsUnauthorized(err) {
					return fmt.Errorf("could not authenticate, client is not authorized: %w", err)
				}
				return err
			}
			log.Println("Logged in to account", self.ID, self.FirstName, self.LastName)
		}
		if err := f(ctx, client, dispatcher, options); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

var (
	CodeRequiredError = errors.New("code is required")
)

type DefaultAuthConversator struct {
	PhoneNumber string
	Passwd      string
}

func (DefaultAuthConversator) SignUp(ctx context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, errors.New("signing up not implemented in Terminal")
}

func (DefaultAuthConversator) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	return &auth.SignUpRequired{TermsOfService: tos}
}

func (k DefaultAuthConversator) Phone(_ context.Context) (string, error) {
	return k.PhoneNumber, nil
}

func (k DefaultAuthConversator) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	return "", CodeRequiredError
}

func (k DefaultAuthConversator) Password(ctx context.Context) (string, error) {
	return k.Passwd, nil
}
