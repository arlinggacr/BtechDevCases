package repository

import (
	"errors"
	"strings"
	"sync"

	"github.com/arlinggacr/BtechDevCases/internal/module/auth/model"
)

var ErrUserNotFound = errors.New("user not found")
var ErrEmailExists = errors.New("email already registered")

var (
	mu    sync.RWMutex
	users = make(map[string]*model.User)
)

func SaveRegisteredAccount(user *model.User) error {
	mu.Lock()
	defer mu.Unlock()
	key := strings.ToLower(user.Email)
	if _, exists := users[key]; exists {
		return ErrEmailExists
	}
	users[key] = user
	return nil
}

func FindAccountByEmail(email string) (*model.User, error) {
	mu.RLock()
	defer mu.RUnlock()
	user, ok := users[strings.ToLower(strings.TrimSpace(email))]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}
