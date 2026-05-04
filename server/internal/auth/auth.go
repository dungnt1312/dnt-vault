package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/dnt/vault-server/internal/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

type AuthService struct {
	jwtSecret []byte
	usersFile string
}

func NewAuthService(jwtSecretFile, usersFile string) (*AuthService, error) {
	secret, err := loadOrCreateSecret(jwtSecretFile)
	if err != nil {
		return nil, err
	}

	return &AuthService{
		jwtSecret: secret,
		usersFile: usersFile,
	}, nil
}

func loadOrCreateSecret(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		return data, nil
	}

	if !os.IsNotExist(err) {
		return nil, err
	}

	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}

	secretHex := []byte(hex.EncodeToString(secret))
	if err := os.WriteFile(path, secretHex, 0600); err != nil {
		return nil, err
	}

	return secretHex, nil
}

func (a *AuthService) Login(username, password string) (string, time.Time, error) {
	users, err := a.loadUsers()
	if err != nil {
		return "", time.Time{}, err
	}

	var user *models.User
	for i := range users.Users {
		if users.Users[i].Username == username {
			user = &users.Users[i]
			break
		}
	}

	if user == nil {
		return "", time.Time{}, ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", time.Time{}, ErrInvalidCredentials
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	token, err := a.generateToken(username, expiresAt)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, expiresAt, nil
}

func (a *AuthService) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return a.jwtSecret, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username, ok := claims["username"].(string)
		if !ok {
			return "", errors.New("invalid token claims")
		}
		return username, nil
	}

	return "", errors.New("invalid token")
}

func (a *AuthService) generateToken(username string, expiresAt time.Time) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.jwtSecret)
}

func (a *AuthService) loadUsers() (*models.UserStore, error) {
	data, err := os.ReadFile(a.usersFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &models.UserStore{Users: []models.User{}}, nil
		}
		return nil, err
	}

	var store models.UserStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}

	return &store, nil
}

func (a *AuthService) CreateUser(username, password string) error {
	users, err := a.loadUsers()
	if err != nil {
		return err
	}

	for _, u := range users.Users {
		if u.Username == username {
			return errors.New("user already exists")
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	users.Users = append(users.Users, models.User{
		Username:     username,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	})

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(a.usersFile, data, 0600)
}
