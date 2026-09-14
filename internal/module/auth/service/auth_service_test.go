package service

import (
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arlinggacr/BtechDevCases/internal/module/auth/model"
	"github.com/arlinggacr/BtechDevCases/internal/module/auth/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func setupAuthServiceTest(t *testing.T) sqlmock.Sqlmock {
	t.Helper()
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create database mock: %v", err)
	}
	repository.ConfigureAuthRepository(database)
	ConfigureAuthService("test-secret", 15*time.Minute)
	t.Cleanup(func() { database.Close() })
	return mock
}

func TestRegisterAccountSavesHashedAccount(t *testing.T) {
	mock := setupAuthServiceTest(t)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id")).WithArgs("user@example.com", sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	user, err := RegisterAccount(model.RegisterRequest{
		Email:           " USER@example.com ",
		Password:        "password123",
		ConfirmPassword: "password123",
	})
	if err != nil {
		t.Fatalf("register account: %v", err)
	}
	if user.ID != 1 || user.Email != "user@example.com" {
		t.Fatalf("unexpected user: %+v", user)
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("password123")) != nil {
		t.Fatal("password was not hashed correctly")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("database expectations: %v", err)
	}
}

func TestRegisterAccountRejectsInvalidDetails(t *testing.T) {
	setupAuthServiceTest(t)

	_, err := RegisterAccount(model.RegisterRequest{
		Email:           "invalid-email",
		Password:        "short",
		ConfirmPassword: "different",
	})
	if err == nil {
		t.Fatal("expected invalid registration to fail")
	}
}

func TestLoginAccountReturnsJWT(t *testing.T) {
	mock := setupAuthServiceTest(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash test password: %v", err)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password_hash FROM users WHERE email = $1")).WithArgs("user@example.com").WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash"}).AddRow(7, "user@example.com", string(hash)))

	response, err := LoginAccount(model.LoginRequest{Email: "user@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("login account: %v", err)
	}
	token, err := jwt.Parse(response.Token, func(token *jwt.Token) (interface{}, error) { return []byte("test-secret"), nil })
	if err != nil || !token.Valid {
		t.Fatalf("parse returned JWT: %v", err)
	}
	claims := token.Claims.(jwt.MapClaims)
	if claims["email"] != "user@example.com" || claims["sub"] != float64(7) {
		t.Fatalf("unexpected JWT claims: %#v", claims)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("database expectations: %v", err)
	}
}

func TestLoginAccountRejectsWrongPassword(t *testing.T) {
	mock := setupAuthServiceTest(t)
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password_hash FROM users WHERE email = $1")).WithArgs("user@example.com").WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash"}).AddRow(7, "user@example.com", string(hash)))

	_, err := LoginAccount(model.LoginRequest{Email: "user@example.com", Password: "wrong-password"})
	if err == nil || !strings.Contains(err.Error(), ErrInvalidCredentials.Error()) {
		t.Fatalf("expected invalid credentials, got: %v", err)
	}
}

func TestAuthenticatedEndpointReturnsWelcomeMessage(t *testing.T) {
	setupAuthServiceTest(t)
	claims := jwt.MapClaims{"sub": int64(7), "email": "user@example.com", "exp": time.Now().Add(time.Minute).Unix()}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("create test JWT: %v", err)
	}

	app := fiber.New()
	app.Get("/api/auth/me", AuthenticateRequest, func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Hello " + c.Locals("email").(string) + ", welcome back"})
	})
	request := httptest.NewRequest("GET", "/api/auth/me", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("call authenticated endpoint: %v", err)
	}
	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}
}

func TestAuthenticatedEndpointRejectsMissingToken(t *testing.T) {
	setupAuthServiceTest(t)
	app := fiber.New()
	app.Get("/api/auth/me", AuthenticateRequest)

	response, err := app.Test(httptest.NewRequest("GET", "/api/auth/me", nil))
	if err != nil {
		t.Fatalf("call protected endpoint: %v", err)
	}
	if response.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", response.StatusCode)
	}
}
