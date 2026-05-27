package database

import (
	"errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
)

func RunMigrations(dsn string) error {
	return RunMigrationsFrom(dsn, "file://migrations")
}

func RunMigrationsFrom(dsn, migrationsPath string) error {
	m, err := migrate.New(migrationsPath, dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	versionBefore, _, _ := m.Version()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Info().Msg("migraciones: sin cambios pendientes")
			return nil
		}
		return err
	}

	versionAfter, _, _ := m.Version()
	log.Info().
		Uint("de", versionBefore).
		Uint("a", versionAfter).
		Uint("aplicadas", versionAfter-versionBefore).
		Msg("migraciones aplicadas")

	return nil
}

func DropAll(dsn, migrationsPath string) error {
	m, err := migrate.New(migrationsPath, dsn)
	if err != nil {
		return err
	}
	defer m.Close()
	return m.Down()
}
