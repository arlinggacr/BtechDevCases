package repository

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/arlinggacr/BtechDevCases/internal/module/auth/model"
)

var ErrUserNotFound = errors.New("user not found")
var ErrEmailExists = errors.New("email already registered")

var database *sql.DB

func ConfigureAuthRepository(db *sql.DB) {
	database = db
}

func SaveRegisteredAccount(user *model.User) error {
	if database == nil {
		return errors.New("auth repository is not configured")
	}
	err := database.QueryRow(
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`,
		strings.ToLower(strings.TrimSpace(user.Email)),
		user.PasswordHash,
	).Scan(&user.ID)
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
		return ErrEmailExists
	}
	return err
}

func FindAccountByEmail(email string) (*model.User, error) {
	if database == nil {
		return nil, errors.New("auth repository is not configured")
	}
	user := &model.User{}
	err := database.QueryRow(
		`SELECT id, email, password_hash FROM users WHERE email = $1`,
		strings.ToLower(strings.TrimSpace(email)),
	).Scan(&user.ID, &user.Email, &user.PasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}
