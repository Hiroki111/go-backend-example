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

func ParseJWTToken(token string) (uint, error) {
	secretKey, err := getSecretKey()
	if err != nil {
		return 0, err
	}

	parsedToken, err := jwt.ParseWithClaims(token, &jwt.MapClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return secretKey, nil
		})
	if err != nil {
		return 0, err
	}

	claims, ok := parsedToken.Claims.(*jwt.MapClaims)
	if !ok {
		return 0, errors.New("unknown claims type, cannot parse the token")
	}

	userID, ok := (*claims)["user_id"].(float64)
	if !ok {
		return 0, errors.New("invalid token")
	}
	return uint(userID), nil
}
