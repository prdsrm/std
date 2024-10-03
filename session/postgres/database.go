package postgres

import (
	"context"
	"embed"
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/prdsrm/std/session"
)

const (
	// 0, illimited number of conns
	MaxConn = 0
	// 0, letting them idle forever
	MaxIdleConn = 0
	// 0, connections are reused forever
	MaxLifetimeConn = 0
)

func GetEnvVariable(name string) (string, error) {
	env, exists := os.LookupEnv(name)
	if !exists {
		return "", fmt.Errorf("Please set the %s environment variable", name)
	}
	return env, nil
}

//go:embed migrations/*.sql
var fs embed.FS

// OpenDBConnection func for opening database connection.
func OpenDBConnection(log *zap.Logger) (*sqlx.DB, error) {
	// Define database connection for PostgreSQL.
	connStr, err := GetEnvVariable("DATABASE_URL")
	if err != nil {
		return nil, err
	}

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error, not connected to database, %w", err)
	}

	// Set database connection settings.
	db.SetMaxOpenConns(MaxConn)
	db.SetMaxIdleConns(MaxIdleConn)
	db.SetConnMaxLifetime(time.Duration(MaxLifetimeConn))

	// Try to ping database.
	if err := db.Ping(); err != nil {
		defer db.Close() // close database connection
		return nil, fmt.Errorf("error, not sent ping to database, %w", err)
	}

	// Run migrations scripts
	d, err := iofs.New(fs, "migrations")
	if err != nil {
		log.Fatal("Couldn't find migrations on disk.", zap.Error(err))
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, connStr)
	if err != nil {
		log.Fatal("Couldn't start a new migrator.", zap.Error(err))
	}
	err = m.Up()
	if err != nil && err.Error() != "no change" {
		return nil, err
	}
	version, dirty, err := m.Version()
	if err != nil {
		log.Fatal("Couldn't get database version.", zap.Error(err))
	}
	if dirty {
		log.Fatal("Database is dirty.")
	}
	log.Info("Checking version.", zap.Uint("version", version))

	return db, nil
}

func ConnectToBotFromDatabase(db *sqlx.DB, botModel Bot, f func(ctx context.Context, client *telegram.Client, dispatcher tg.UpdateDispatcher, options telegram.Options) error, phone string, password string) error {
	device, err := GetDevice(db, botModel.UserID)
	if err != nil {
		return err
	}
	flow := session.GetNewDefaultAuthConversator(phone, password)
	err = session.Connect(f, session.Windows(), device.ApiID, device.ApiHash, device.SessionString, device.Proxy.String, flow)
	if err != nil {
		return err
	}
	return nil
}
