package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/thoas/go-funk"
	"golang.org/x/crypto/bcrypt"
)

func getSecretKey() string {
	secret := os.Getenv("SECRET")
	if secret == "" {
		secret = "secret"
	}
	return secret
}

type SignedDetails struct {
	Id    string       `json:"id,omitempty"`
	Email string       `json:"email,omitempty"`
	Data  interface{}  `json:"data,omitempty"`
	Orgs  []OrgContext `json:"orgs,omitempty"`
	jwt.StandardClaims
}

var SECRET_KEY string = getSecretKey()

func GenerateToken(userContext UserContext) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Id:    userContext.Id,
		Email: userContext.Email,
		Data:  userContext.Data,
		Orgs:  userContext.Orgs,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	refreshClaims := &SignedDetails{
		Email: userContext.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return
	}

	return token, refreshToken, err
}

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("login or passowrd is incorrect")
		check = false
	}

	return check, msg
}

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

	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = fmt.Sprintf("token is expired")
		msg = err.Error()
		return
	}

	return claims, msg
}

func ValidateTokenMiddleware(c *gin.Context) (claims *SignedDetails, valid bool) {
	valid = false
	clientToken := c.Request.Header.Get("Authorization")
	if clientToken == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No Authorization header provided"})
		c.Abort()
		return
	}

	claims, err := ValidateToken(clientToken)
	if err != "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		c.Abort()
		return
	}
	valid = true
	return claims, valid
}

// Auth validates token and authorizes users
func AuthenticationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken := c.Request.Header.Get("Authorization")
		if clientToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("No Authorization header provided")})
			c.Abort()
			return
		}

		claims, err := ValidateToken(clientToken)
		if err != "" {
			c.JSON(http.StatusForbidden, gin.H{"error": err})
			c.Abort()
			return
		}

		// dump(claims)

		c.Set("email", claims.Email)
		c.Set("id", claims.Id)
		c.Set("data", claims.Data)
		c.Set("orgs", claims.Orgs)

		c.Next()

	}
}

func roleCheck(ctx *gin.Context, orgId string, role string) bool {
	orgs, _ := ctx.Get("orgs")

	r := funk.Find(orgs.([]OrgContext), func(org OrgContext) bool {
		return org.Id == orgId && org.Role == role
	})

	if r != nil {
		return true
	}
	return false
}
