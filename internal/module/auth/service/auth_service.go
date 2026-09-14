package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/arlinggacr/BtechDevCases/internal/module/auth/model"
	"github.com/arlinggacr/BtechDevCases/internal/module/auth/repository"
)

var ErrInvalidCredentials = errors.New("invalid email or password")
var ErrEmailExists = repository.ErrEmailExists
var ErrInvalidRegistration = errors.New("invalid registration details")
var jwtSecret []byte
var tokenTTL time.Duration

func ConfigureAuthService(secret string, ttl time.Duration) {
	jwtSecret = []byte(secret)
	tokenTTL = ttl
}

func RegisterAccount(request model.RegisterRequest) (*model.User, error) {
	email := strings.ToLower(strings.TrimSpace(request.Email))
	if !strings.Contains(email, "@") || len(request.Password) < 8 || request.Password != request.ConfirmPassword {
		return nil, ErrInvalidRegistration
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash account password: %w", err)
	}

	user := &model.User{Email: email, PasswordHash: string(hash)}
	if err := repository.SaveRegisteredAccount(user); err != nil {
		return nil, fmt.Errorf("save registered account: %w", err)
	}
	return user, nil
}

func LoginAccount(request model.LoginRequest) (*model.AuthResponse, error) {
	user, err := repository.FindAccountByEmail(request.Email)
	if err != nil {
		return nil, fmt.Errorf("find account for login: %w", ErrInvalidCredentials)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)); err != nil {
		return nil, fmt.Errorf("compare login password: %w", ErrInvalidCredentials)
	}

	token, err := createAccountToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("sign account token: %w", err)
	}
	return &model.AuthResponse{Token: token, User: user}, nil
}

func createAccountToken(userID int64, email string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"iat":   now.Unix(),
		"exp":   now.Add(tokenTTL).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
}

func AuthenticateRequest(c *fiber.Ctx) error {
	value := c.Get("Authorization")
	if !strings.HasPrefix(value, "Bearer ") {
		return fiber.ErrUnauthorized
	}
	token, err := jwt.Parse(strings.TrimPrefix(value, "Bearer "), func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return fiber.ErrUnauthorized
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fiber.ErrUnauthorized
	}
	email, ok := claims["email"].(string)
	if !ok {
		return fiber.ErrUnauthorized
	}
	userID, ok := claims["sub"].(float64)
	if !ok {
		return fiber.ErrUnauthorized
	}
	refreshedToken, err := createAccountToken(int64(userID), email)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	c.Set("X-Auth-Token", refreshedToken)
	c.Locals("email", email)
	return c.Next()
}
