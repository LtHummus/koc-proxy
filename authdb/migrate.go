package authdb

import (
	"embed"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/rs/zerolog/log"
)

//go:embed migrations/*
var migrations embed.FS

func createMigration() (*migrate.Migrate, error) {
	d, err := iofs.New(migrations, "migrations")
	if err != nil {
		log.Error().Err(err).Msg("could not load migration files")
		return nil, fmt.Errorf("db: migrateDatabase: could not load migration files: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, buildConnectionString("pgx"))
	if err != nil {
		return nil, err
	}
	m.Log = &migrationLogger{}

	return m, nil
}

func GetVersionInfo() error {
	m, err := createMigration()
	if err != nil {
		return err
	}

	version, dirty, err := m.Version()
	if err != nil {
		return err
	}

	log.Info().Uint("version", version).Bool("dirty", dirty).Msg("retrieved database version information")
	return nil
}

func MigrateDatabase(version uint) error {
	m, err := createMigration()
	if err != nil {
		return err
	}

	err = m.Migrate(version)
	if err == migrate.ErrNoChange {
		log.Warn().Uint("version", version).Msg("no migration required")
		return nil
	}

	if err != nil {
		return err
	}

	return nil
}

// migrationLogger is a shim logger to allow the database migration code to hook in zerolog
type migrationLogger struct{}

func (m *migrationLogger) Printf(format string, v ...interface{}) {
	log.Printf(strings.TrimSpace(format), v...)
}

func (m *migrationLogger) Verbose() bool {
	return true
}
