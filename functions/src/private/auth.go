package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func getSecretKey() string {
	secret := os.Getenv("SECRET")
	if secret == "" {
		secret = "secret"
	}
	return secret
}

type SignedDetails struct {
	Email string
	jwt.StandardClaims
}

var SECRET_KEY string = getSecretKey()

func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)

	if err != nil {
		msg = err.Error()
		return
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = fmt.Sprintf("the token is invalid")
		msg = err.Error()
		return
	}

	// if claims.ExpiresAt < time.Now().Local().Unix() {
	// 	msg = fmt.Sprintf("token is expired")
	// 	msg = err.Error()
	// 	return
	// }

	return claims, msg
}

func ValidateTokenMiddleware(c *gin.Context) (claims *SignedDetails, validate bool) {
	validate = false
	clientToken := c.Request.Header.Get("Authorization")
	if clientToken == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("No Authorization header provided")})
		c.Abort()
		return
	}

	claims, err := ValidateToken(clientToken)
	if err != "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		c.Abort()
		return
	}
	validate = true
	return claims, validate
}

// Auth validates token and authorizes users
func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken := c.Request.Header.Get("Authorization")
		if clientToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("No Authorization header provided")})
			c.Abort()
			return
		}

		claims, err := ValidateToken(clientToken)
		if err != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			c.Abort()
			return
		}

		c.Set("email", claims.Email)

		c.Next()

	}
}
