package session

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/net/proxy"
	"golang.org/x/time/rate"
	"net/url"
	"sync"
	"time"

	"github.com/gotd/td/session"

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

func newDialer(proxyConnStr string) (proxy.Dialer, error) {
	url, err := url.Parse(proxyConnStr)
	if err != nil {
		return nil, err
	}
	socks5, err := proxy.FromURL(url, proxy.Direct)
	if err != nil {
		return nil, err
	}
	return socks5, nil
}

func newResolver(proxyConnStr string) (dcs.Resolver, error) {
	var resolver dcs.Resolver
	socks5, err := newDialer(proxyConnStr)
	if err != nil {
		return nil, err
	}
	dc := socks5.(proxy.ContextDialer)
	resolver = dcs.Plain(dcs.PlainOptions{
		Dial: dc.DialContext,
	})
	return resolver, nil
}

func Connect(f func(ctx context.Context, client *telegram.Client, dispatcher tg.UpdateDispatcher, options telegram.Options) error, phone string, password string, device telegram.DeviceConfig, apiID int, apiHash string, sessionString string, proxy string) error {
	data, err := session.TelethonSession(sessionString)
	if err != nil {
		return err
	}
	var resolver dcs.Resolver
	if proxy != "" {
		resolver, err = newResolver(proxy)
		if err != nil {
			return err
		}
	}
	storage := &MemorySession{}
	loader := session.Loader{Storage: storage}
	// Save decoded Telethon session as gotd session.
	if err := loader.Save(context.Background(), data); err != nil {
		return err
	}
	options := telegram.Options{
		Resolver:       resolver,
		SessionStorage: storage,
		Device:         device,
	}
	userAuthenticator := DefaultAuthConversator{PhoneNumber: phone, Passwd: password}
	authOpt := auth.SendCodeOptions{}
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
	// Authentication flow handles authentication process, like prompting for code and 2FA password.
	flow := auth.NewFlow(userAuthenticator, authOpt)

	ctx := context.Background()
	if err := waiter.Run(ctx, func(ctx context.Context) error {
		// Spawning main goroutine.
		if err := client.Run(ctx, func(ctx context.Context) error {
			// Checking auth status.
			status, err := client.Auth().Status(ctx)
			if err != nil {
				return err
			}
			// Can be already authenticated if we have valid session in
			// session storage.
			if !status.Authorized {
				if err := client.Auth().IfNecessary(ctx, flow); err != nil {
					return fmt.Errorf("could not authenticate: %w", err)
				}
			}
			if err := f(ctx, client, dispatcher, options); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return err
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
