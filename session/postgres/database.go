package postgres

import (
	"embed"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

const (
	// 0, illimited number of conns
	MaxConn = 0
	// 0, letting them idle forever
	MaxIdleConn = 0
	// 0, connections are reused forever
	MaxLifetimeConn = 0
)

//go:embed migrations/*.sql
var fs embed.FS

// OpenDBConnection func for opening database connection.
func OpenDBConnection(databaseURL string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", databaseURL)
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
		log.Fatalln("Couldn't find migrations on disk.", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, databaseURL)
	if err != nil {
		log.Fatalln("Couldn't start a new migrator.", err)
	}
	err = m.Up()
	if err != nil && err.Error() != "no change" {
		return nil, err
	}
	_, _, err = m.Version()
	if err != nil {
		log.Fatalln("Couldn't get database version.", err)
	}

	return db, nil
}
