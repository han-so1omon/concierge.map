package auth

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/han-so1omon/concierge.map/data"
	"os"
	"time"
)

var jwtSecretKey []byte

func CreateJWT(email string) (response string, err error) {
	jwtSecretKey = []byte(os.Getenv("JWT_SECRET_KEY"))
	expirationTime := time.Now().Add(21 * time.Minute)
	claims := &data.Claims{
		Email: email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecretKey)
	if err == nil {
		return tokenString, nil
	}

	return "", err
}

func VerifyToken(tokenString string) (email string, err error) {
	claims := &data.Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecretKey, nil
	})

	if token != nil {
		return claims.Email, nil
	}

	return "", err
}
