package auth

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/niradler/social-lab/src/db"
	"github.com/niradler/social-lab/src/types"
	"github.com/niradler/social-lab/src/utils"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/thoas/go-funk"
	"golang.org/x/crypto/bcrypt"
)

var isProd = os.Getenv("LAMBDA_TASK_ROOT") != ""

func getSecretKey() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "secret"
	}
	return secret
}

type SignedDetails struct {
	Id    string             `json:"id,omitempty"`
	Email string             `json:"email,omitempty"`
	Data  interface{}        `json:"data,omitempty"`
	Orgs  []types.OrgContext `json:"orgs,omitempty"`
	jwt.StandardClaims
}

var SECRET_KEY string = getSecretKey()

func GenerateToken(userContext types.UserContext) (signedToken string, signedRefreshToken string, err error) {
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

		c.Set("email", claims.Email)
		c.Set("id", claims.Id)
		c.Set("data", claims.Data)
		c.Set("orgs", claims.Orgs)

		c.Next()

	}
}

func RoleCheck(ctx *gin.Context, orgId string, role string) bool {
	orgs, _ := ctx.Get("orgs")

	r := funk.Find(orgs.([]types.OrgContext), func(org types.OrgContext) bool {
		return org.Id == orgId && org.Role == role
	})

	if r != nil {
		return true
	}
	return false
}

func askResetPassword(email string) error {
	user, err := db.GetItem(db.ToKey("user", email), "#")
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("Not Found")
	}
	//send reset email
	return nil
}

func getProviderConfiguration() (string, string, string, string) {

	err := godotenv.Load()
	if err != nil {
		utils.Logger.Error("Error loading .env file")
	}

	clientId := os.Getenv("GOOGLE_CLIENT_ID")
	if clientId == "" {
		clientId = "clientId"
	}

	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if clientSecret == "" {
		clientSecret = "clientSecret"
	}

	clientAuthCallback := os.Getenv("GOOGLE_CALLBACK")
	if clientAuthCallback == "" {
		clientAuthCallback = "GOOGLE_CALLBACK"
	}

	clientCallback := os.Getenv("CLIENT_CALLBACK")
	if clientCallback == "" {
		clientCallback = "CLIENT_CALLBACK"
	}

	log.Println(clientId)
	log.Println(clientCallback)
	log.Println(clientAuthCallback)

	return clientCallback, clientId, clientSecret, clientAuthCallback

}

func ProvidersAuthCallback(provider string, ctx *gin.Context) {
	clientCallback, clientId, clientSecret, clientAuthCallback := getProviderConfiguration()
	goth.UseProviders(
		google.New(clientId, clientSecret, clientAuthCallback, "email", "profile"),
	)
	q := ctx.Request.URL.Query()
	q.Add("provider", provider)
	ctx.Request.URL.RawQuery = q.Encode()
	user, err := gothic.CompleteUserAuth(ctx.Writer, ctx.Request)
	if utils.HandlerError(ctx, err, 500) {
		return
	}

	existUser, _ := db.GetItem(db.ToKey("user", user.Email), "#")
	if existUser == nil {
		_, err = db.CreateUser(types.UserPayload{
			Email:    user.Email,
			Password: "",
			Data:     user.RawData,
		})
		if utils.HandlerError(ctx, err, http.StatusBadRequest) {
			return
		}
	}

	userContext, err := db.GetUserContext(user.Email)
	if utils.HandlerError(ctx, err, http.StatusBadRequest) {
		return
	}
	token, refreshToken, _ := GenerateToken(*userContext)

	ctx.Redirect(http.StatusMovedPermanently, clientCallback+"?token="+token+"&refreshToken="+refreshToken)
}

func ProvidersAuthBegin(provider string, ctx *gin.Context) {
	_, clientId, clientSecret, clientAuthCallback := getProviderConfiguration()
	goth.UseProviders(
		google.New(clientId, clientSecret, clientAuthCallback, "email", "profile"),
	)

	q := ctx.Request.URL.Query()
	q.Add("provider", provider)
	ctx.Request.URL.RawQuery = q.Encode()
	gothic.BeginAuthHandler(ctx.Writer, ctx.Request)
}

func GothInit() {
	key := os.Getenv("SESSION_SECRET") // Replace with your SESSION_SECRET or similar
	maxAge := 86400 * 30               // 30 days
	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(maxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = isProd // Set to true when serving over https

	gothic.Store = store
}
