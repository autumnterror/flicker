package psql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/autumnterror/breezynotes/pkg/utils/format"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNoUser            = errors.New("no user found")
	ErrPasswordIncorrect = errors.New("password incorrect")
	ErrWrongInput        = errors.New("wrong input")
)

// Authentication search user login and password in database and compare
func (d *Driver) Authentication(ctx context.Context, email, login, password string) (string, error) {
	const op = "psql.Authentication"

	ctx, done := context.WithTimeout(ctx, waitTime)
	defer done()

	if password == "" {
		return "", format.Error(op, ErrWrongInput)
	}

	var (
		query string
		arg   string
	)

	switch {
	case login != "":
		query = `SELECT id, password FROM users WHERE login = $1`
		arg = login
	case email != "":
		query = `SELECT id, password FROM users WHERE email = $1`
		arg = email
	default:
		return "", format.Error(op, ErrWrongInput)
	}

	var hashed string
	var id string
	if err := d.driver.QueryRowContext(ctx, query, arg).Scan(&id, &hashed); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNoUser
		}
		return "", format.Error(op, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", ErrPasswordIncorrect
		}
		return "", format.Error(op, err)
	}

	return id, nil
}
