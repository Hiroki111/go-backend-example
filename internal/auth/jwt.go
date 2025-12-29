package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const tokenTTL = time.Hour

func getSecretKey() ([]byte, error) {
	key := os.Getenv("SECRET_KEY")
	if key == "" {
		return nil, errors.New("SECRET_KEY not set")
	}
	return []byte(key), nil
}

func GenerateJWTToken(userID uint) (string, error) {
	secretKey, err := getSecretKey()
	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(tokenTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// Use it when you need to parse and validate a JWT token
func ParseJWTToken(token string) (uint, error) {
	secretKey, err := getSecretKey()
	if err != nil {
		return 0, err
	}

	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(token, claims,
		func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		})
	if err != nil {
		return 0, err
	}
	userID, ok := claims["user_id"].(uint)
	if !ok {
		return 0, errors.New("invalid token")
	}
	return userID, nil
}
