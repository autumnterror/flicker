package main

import (
	"database/sql"
	"errors"
	"flag"
	"flicker/internal/config"
	"fmt"
	"log"

	"github.com/autumnterror/breezynotes/pkg/utils/format"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	typeMigration := flag.String("type", "up", "type of migration action")
	flag.Parse()

	if err := executeMigrate(*typeMigration); err != nil {
		log.Fatal(err)
	}
}

func executeMigrate(TYPE string) error {
	const op = "migrator.executeMigrate"

	cfg := config.MustSetup()
	db, err := sql.Open("postgres", cfg.Uri)
	if err != nil {
		return format.Error(op, err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(format.Error(op, err))
		}
	}(db)

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return format.Error(op, err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		return format.Error(op, err)
	}
	defer func(m *migrate.Migrate) {
		err, _ := m.Close()
		if err != nil {
			log.Fatal(format.Error(op, err))
		}
	}(m)

	switch TYPE {
	case "up":
		err = m.Up()
		if err != nil {
			log.Fatal(format.Error(op, err))
		}
		log.Println("Migrations applied successfully!")
		return nil
	case "down":
		if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return format.Error(op, err)
		}
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("No migrations to rollback")
		}
		log.Println("Migration rolled back successfully!")
		return nil
	default:
		return fmt.Errorf("flag --type not recognized")
	}

}
