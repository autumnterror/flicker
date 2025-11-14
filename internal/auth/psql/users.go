package psql

import (
	"context"
	"database/sql"
	"errors"
	"flicker/internal/views"
	"strings"

	"github.com/autumnterror/breezynotes/pkg/log"
	"github.com/autumnterror/breezynotes/pkg/utils/format"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrAlreadyExist = errors.New("user alreay exist")
)

// GetAll only for test
func (d *Driver) GetAll(ctx context.Context) ([]*views.User, error) {
	const op = "psql.users.GetAll"

	ctx, done := context.WithTimeout(ctx, waitTime)
	defer done()

	var ls []*views.User
	rows, err := d.driver.QueryContext(ctx, `SELECT * FROM users`)
	if err != nil {
		return nil, format.Error(op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var us views.User
		if err := rows.Scan(&us.Id, &us.Login, &us.Email, &us.About, &us.Password, &us.Photo); err != nil {
			log.Error(op, "rows scan error", err)
			continue
		}
		ls = append(ls, &us)
	}

	return ls, nil
}

// Create new user
func (d *Driver) Create(ctx context.Context, u *views.User) error {
	const op = "psql.users.Create"

	ctx, done := context.WithTimeout(ctx, waitTime)
	defer done()

	query := `
				INSERT INTO users (id, login, email, about, password, photo)
				VALUES ($1, $2, $3, $4, $5, $6)
			`

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return format.Error(op, err)
	}

	_, err = d.driver.ExecContext(ctx, query, u.Id, u.Login, u.Email, u.About, hashedPass, u.Photo)
	if err != nil {
		if isDuplicateKeyError(err) {
			return format.Error(op, ErrAlreadyExist)
		}
		return format.Error(op, err)
	}

	return nil
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "23505")
}

// UpdatePassword updates user's password by user ID.
// Returns sql.ErrNoRows if user not found.
func (d *Driver) UpdatePassword(ctx context.Context, id, newPassword string) error {
	const op = "psql.users.UpdatePassword"
	ctx, done := context.WithTimeout(ctx, waitTime)
	defer done()

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return format.Error(op, err)
	}

	res, err := d.driver.ExecContext(ctx, `UPDATE users SET password = $1 WHERE id = $2`, hashedPass, id)
	if err != nil {
		return format.Error(op, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return format.Error(op, err)
	}
	if rows == 0 {
		return format.Error(op, sql.ErrNoRows)
	}

	return nil
}

// UpdatePhoto updates user's photo by user ID.
// Returns sql.ErrNoRows if user not found.
func (d *Driver) UpdatePhoto(ctx context.Context, id, np string) error {
	const op = "psql.users.UpdatePhoto"

	ctx, done := context.WithTimeout(ctx, waitTime)
	defer done()

	res, err := d.driver.ExecContext(ctx, `UPDATE users SET photo = $1 WHERE id = $2`, np, id)
	if err != nil {
		return format.Error(op, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return format.Error(op, err)
	}
	if rows == 0 {
		return format.Error(op, sql.ErrNoRows)
	}

	return nil
}

// UpdateEmail updates user's email by user ID.
// Returns sql.ErrNoRows if user not found.
func (d *Driver) UpdateEmail(ctx context.Context, id, email string) error {
	const op = "psql.users.UpdateEmail"

	res, err := d.driver.ExecContext(ctx, `UPDATE users SET email = $1 WHERE id = $2`, email, id)
	if err != nil {
		return format.Error(op, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return format.Error(op, err)
	}
	if rows == 0 {
		return format.Error(op, sql.ErrNoRows)
	}

	return nil
}

// UpdateAbout updates user's about section by user ID.
// Returns sql.ErrNoRows if user not found.
func (d *Driver) UpdateAbout(ctx context.Context, id, about string) error {
	const op = "psql.users.UpdateAbout"

	res, err := d.driver.ExecContext(ctx, `UPDATE users SET about = $1 WHERE id = $2`, about, id)
	if err != nil {
		return format.Error(op, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return format.Error(op, err)
	}
	if rows == 0 {
		return format.Error(op, sql.ErrNoRows)
	}

	return nil
}

// Delete user. May send sql.ErrNoRows
func (d *Driver) Delete(ctx context.Context, id string) error {
	const op = "psql.users.Delete"

	query := `
				DELETE FROM users
				WHERE id = $1
			`
	res, err := d.driver.ExecContext(ctx, query, id)
	if err != nil {
		return format.Error(op, err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return format.Error(op, err)
	}
	if rowsAffected == 0 {
		return format.Error(op, sql.ErrNoRows)
	}
	return nil
}

// GetInfo get info about user by login. May send sql.ErrNoRows
func (d *Driver) GetInfo(ctx context.Context, id string) (*views.User, error) {
	const op = "psql.users.GetInfo"
	query := `
		SELECT login,email,about FROM users
		WHERE id = $1
	`
	var u views.User
	if err := d.driver.QueryRowContext(ctx, query, id).Scan(&u.Login, &u.Email, &u.About); err != nil {
		return nil, format.Error(op, err)
	}

	return &u, nil
}

// CheckUserExists checks if a user with given login or email exists.
// Returns nil if user is found by either field, otherwise returns sql.ErrNoRows.
//func (d *driver) CheckUserExists(login, email string) error {
//	const op = "psql.users.CheckUserExists"
//
//	var (
//		query string
//		args  []any
//	)
//
//	switch {
//	case login != "" && email != "":
//		query = `SELECT 1 FROM users WHERE login = $1 OR email = $2 LIMIT 1`
//		args = []any{login, email}
//	case login != "":
//		query = `SELECT 1 FROM users WHERE login = $1 LIMIT 1`
//		args = []any{login}
//	case email != "":
//		query = `SELECT 1 FROM users WHERE email = $1 LIMIT 1`
//		args = []any{email}
//	default:
//		return format.Error(op, errors.New("no input provided"))
//	}
//
//	row := d.driver.QueryRow(query, args...)
//	var dummy int
//	if err := row.Scan(&dummy); err != nil {
//		return format.Error(op, err)
//	}
//
//	return nil
//}

// Update user. May send sql.ErrNoRows
//func (d *driver) Update(u *views.User, id string) error {
//	const op = "psql.users.Update"
//
//	query := `
//				UPDATE users
//				SET login = $1, email = $2, about = $3, password = $4, photo = $5
//				WHERE id = $6
//			`
//
//	hashedPass, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
//	if err != nil {
//		return format.Error(op, err)
//	}
//
//	res, err := d.driver.Exec(query, u.GetLogin(), u.GetEmail(), u.GetAbout(), hashedPass, u.GetPhoto(), id)
//	if err != nil {
//		return format.Error(op, err)
//	}
//
//	ra, err := res.RowsAffected()
//	if err != nil {
//		return format.Error(op, err)
//	}
//
//	if ra == 0 {
//		return format.Error(op, sql.ErrNoRows)
//	}
//	return nil
//}
