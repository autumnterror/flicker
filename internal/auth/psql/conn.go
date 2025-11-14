package psql

import (
	"context"
	"database/sql"
	"errors"
	"flicker/internal/config"
	"flicker/internal/views"
	"time"

	"github.com/autumnterror/breezynotes/pkg/log"
	"github.com/autumnterror/breezynotes/pkg/utils/format"
	_ "github.com/lib/pq"
)

type PostgresDb struct {
	Driver *sql.DB
}

type SqlRepo interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

const (
	waitTime = 3 * time.Second
)

type Driver struct {
	driver SqlRepo
}

func NewDriver(driver SqlRepo) *Driver {
	return &Driver{
		driver: driver,
	}
}

func MustConnect(cfg *config.Config) *PostgresDb {
	db, err := NewConnect(cfg)
	if err != nil {
		log.Panic(err)
	}
	return db
}

// NewConnect is constructor of PostgresDb. Construct with connection
func NewConnect(cfg *config.Config) (*PostgresDb, error) {
	const op = "psql.NewConnect"

	db, err := sql.Open("postgres", cfg.Uri)
	if err != nil {
		return nil, format.Error(op, err)
	}
	err = db.Ping()
	if err != nil {
		return nil, format.Error(op, err)
	}

	log.Success(op, "Connection to postgresSQL is established")
	return &PostgresDb{Driver: db}, nil
}

func (d *PostgresDb) Disconnect() error {
	const op = "psql.PostgresDb.Disconnect"
	err := d.Driver.Close()
	if err != nil {
		return format.Error(op, err)
	}
	err = d.Driver.Ping()
	if err == nil {
		return format.Error(op, errors.New("failed to disconnect"))
	}

	log.Success(op, "Connection to postgresSQL terminated")
	return nil
}

type AuthRepo interface {
	Authentication(ctx context.Context, email, login, password string) (string, error)
	GetAll(ctx context.Context) ([]*views.User, error)
	Create(ctx context.Context, u *views.User) error
	UpdatePhoto(ctx context.Context, id, np string) error
	UpdatePassword(ctx context.Context, id, newPassword string) error
	UpdateEmail(ctx context.Context, id, email string) error
	UpdateAbout(ctx context.Context, id, about string) error
	Delete(ctx context.Context, id string) error
	GetInfo(ctx context.Context, id string) (*views.User, error)
}
