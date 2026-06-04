package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type bcryptHasher struct{}

func NewBcryptHasher() PasswordHasher {
	return &bcryptHasher{}
}

func (h *bcryptHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (h *bcryptHasher) Verify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

type jwtTokenGenerator struct {
	secret    string
	expiresIn time.Duration
}

func NewJWTTokenGenerator(secret string, expiresIn time.Duration) TokenGenerator {
	return &jwtTokenGenerator{
		secret:    secret,
		expiresIn: expiresIn,
	}
}

func (g *jwtTokenGenerator) GenerateToken(userID int64, email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(g.expiresIn).Unix(),
	})

	tokenString, err := token.SignedString([]byte(g.secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
