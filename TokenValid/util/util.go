package util

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"log"
	"os"
	"time"
)

var secretKey []byte
var expirationTime time.Duration

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	viper.AutomaticEnv()
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "defaultSecretKey"
	}
	secretKey = []byte(secret)
	expirationtime := os.Getenv("JWT_EXPIRATION")
	if expirationtime == "" {
		expirationtime = "5m"
	}
	expirationTime, _ = time.ParseDuration(expirationtime)
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func GenerateToken(username string) (string, error) {
	expirationTime := time.Now().Add(expirationTime)
	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func ValidateToken(tokenString, username string) (bool, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil || !token.Valid || claims.Username != username {
		return false, err
	}

	return true, nil
}
