package postgres

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

const (
	migrationPath = "C:/Users/Leonid/Desktop/catalog/internal/storage/migrations"
)

type Storage struct {
	DB *sql.DB
}

func New(driver, info string) (*Storage, error) {
	const op = "storage.postgres.NewStorage" // Current function

	DB, err := sql.Open(driver, info) // Connet to DB
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{DB: DB}, nil
}

func (s *Storage) UpMigration(info string) error {
	const op = "storage.postgres.UpMigration"

	m, err := migrate.New(
		"file://"+migrationPath,
		"postgres://"+info)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := m.Up(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
