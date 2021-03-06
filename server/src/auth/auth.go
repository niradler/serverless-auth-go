package auth

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/bitbucket"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/gitlab"
	"github.com/markbates/goth/providers/google"
	"github.com/niradler/social-lab/src/db"
	"github.com/niradler/social-lab/src/types"
	"github.com/niradler/social-lab/src/utils"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/thoas/go-funk"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var isProd = os.Getenv("LAMBDA_TASK_ROOT") != ""

func getSecretKey() string {
	secret := os.Getenv("SLS_AUTH_JWT_SECRET")
	if secret == "" {
		secret = "secret"
	}
	return secret
}

type RefreshTokenClaims struct {
	Id      string `json:"id,omitempty"`
	Email   string `json:"email,omitempty"`
	Refresh bool   `json:"refresh,omitempty"`
	jwt.StandardClaims
}

type TokenClaims struct {
	Id    string             `json:"id,omitempty"`
	Email string             `json:"email,omitempty"`
	Data  interface{}        `json:"data,omitempty"`
	Orgs  []types.OrgContext `json:"orgs,omitempty"`
	jwt.StandardClaims
}

var SECRET_KEY string = getSecretKey()

func GenerateToken(userContext types.UserContext) (signedToken string, signedRefreshToken string, err error) {
	claims := &TokenClaims{
		Id:    userContext.Id,
		Email: userContext.Email,
		Data:  userContext.Data,
		Orgs:  userContext.Orgs,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	refreshClaims := &RefreshTokenClaims{
		Id:      userContext.Id,
		Email:   userContext.Email,
		Refresh: true,
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

func VerifyPassword(userPassword string, providedPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true

	if err != nil {
		check = false
		return check, err
	}

	return check, nil
}

func ValidateToken(clientToken string) (claims *TokenClaims, err error) {
	if clientToken == "" {
		return nil, errors.New("No Authorization header provided")
	}

	token, err := jwt.ParseWithClaims(
		clientToken,
		&TokenClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)

	if err != nil {
		return nil, errors.New("Failed to parse token")
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return nil, errors.New("Failed to get claims")
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		return nil, errors.New("Token expired")
	}

	return claims, nil
}

func ValidateRefreshToken(clientToken string) (claims *RefreshTokenClaims, err error) {
	if clientToken == "" {
		return nil, errors.New("No Authorization header provided")
	}

	token, err := jwt.ParseWithClaims(
		clientToken,
		&RefreshTokenClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)

	if err != nil {
		return nil, errors.New("Failed to parse token")
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)
	if !ok {
		return nil, errors.New("Failed to get claims")
	}

	if claims.Refresh != true {
		return nil, errors.New("Not a refresh token")
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		return nil, errors.New("Token expired")
	}

	return claims, nil
}

func AuthenticationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken := c.Request.Header.Get("Authorization")
		if clientToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("No Authorization header provided")})
			c.Abort()
			return
		}

		claims, err := ValidateToken(c.Request.Header.Get("Authorization"))
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
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

func RoleCheck(orgId string, userId string, role string) bool {
	utils.Logger.Info("RoleCheck", zap.String("orgId", orgId), zap.String("userId", userId), zap.String("role", role))
	orgUser, err := db.GetItem(db.ToKey("user", userId), db.GenerateKey("org", orgId))
	if err != nil {
		return false
	}
	if orgUser["role"].(string) == role && orgUser["sk"].(string) == db.GenerateKey("org", orgId) {
		return true
	}

	return false
}

func JWTRoleCheck(ctx *gin.Context, orgId string, role string) bool {
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

func getClientCallback() string {

	clientCallback := os.Getenv("SLS_AUTH_CLIENT_CALLBACK")
	if clientCallback == "" {
		clientCallback = "CLIENT_CALLBACK"
	}

	log.Println(clientCallback)

	return clientCallback

}

func getProviderConfiguration(provider string) (string, string, string) {
	providerUpperCase := strings.ToUpper(provider)

	err := godotenv.Load()
	if err != nil {
		utils.Logger.Error("Error loading .env file")
	}

	clientId := os.Getenv("SLS_AUTH_" + providerUpperCase + "_CLIENT_ID")
	if clientId == "" {
		clientId = "clientId"
	}

	clientSecret := os.Getenv("SLS_AUTH_" + providerUpperCase + "_CLIENT_SECRET")
	if clientSecret == "" {
		clientSecret = "clientSecret"
	}

	clientAuthCallback := os.Getenv("SLS_AUTH_" + providerUpperCase + "_CALLBACK")
	if clientAuthCallback == "" {
		clientAuthCallback = "clientAuthCallback"
	}

	log.Println(clientId)
	log.Println(clientAuthCallback)

	return clientId, clientSecret, clientAuthCallback

}

func ProvidersAuthCallback(provider string, ctx *gin.Context) {
	utils.Logger.Info("ProvidersAuthCallback:", zap.String("provider", provider))
	q := ctx.Request.URL.Query()
	q.Add("provider", provider)
	ctx.Request.URL.RawQuery = q.Encode()
	user, err := gothic.CompleteUserAuth(ctx.Writer, ctx.Request)
	if utils.HandlerError(ctx, err, 500) {
		return
	}
	if utils.Debug {
		utils.Logger.Info("user:", zap.Any("user", user))
	}
	if user.Email == "" || strings.Contains(user.Email, "@") == false {
		utils.HandlerError(ctx, errors.New("missing email address."), http.StatusBadRequest)
		return

	}
	utils.Logger.Info("ProvidersAuthCallback:", zap.String("email", user.Email))
	existUser, _ := db.GetItem(db.ToKey("user", user.Email), "#")
	if existUser == nil {
		_, err = db.CreateUser(types.UserPayload{
			Email:    user.Email,
			Password: "",
			Data: map[string]string{
				"email":          user.Email,
				"provider":       user.Provider,
				"name":           user.Name,
				"firstName":      user.FirstName,
				"lastName":       user.LastName,
				"providerUserId": user.UserID,
				"avatarUrl":      user.AvatarURL,
			},
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
	clientCallback := getClientCallback()
	ctx.Redirect(http.StatusMovedPermanently, clientCallback+"?token="+token+"&refreshToken="+refreshToken)
}

func ProvidersAuthBegin(provider string, ctx *gin.Context) {
	utils.Logger.Error("ProvidersAuthBegin:", zap.String("provider", provider))
	q := ctx.Request.URL.Query()
	q.Add("provider", provider)
	ctx.Request.URL.RawQuery = q.Encode()
	gothic.BeginAuthHandler(ctx.Writer, ctx.Request)
}

func GothInit() {
	err := godotenv.Load()
	if err != nil {
		utils.Logger.Error("Error loading .env file")
	}

	key := os.Getenv("SLS_AUTH_SESSION_SECRET") // Replace with your SESSION_SECRET or similar
	maxAge := 86400 * 30                        // 30 days
	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(maxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = isProd // Set to true when serving over https

	gothic.Store = store

	providers := []string{"facebook", "google", "github", "bitbucket", "gitlab"}
	for _, provider := range providers {
		clientId, clientSecret, clientAuthCallback := getProviderConfiguration(provider)
		switch provider {
		case "google":
			goth.UseProviders(google.New(clientId, clientSecret, clientAuthCallback, "email", "profile"))
		case "github":
			goth.UseProviders(github.New(clientId, clientSecret, clientAuthCallback, "user:email", "profile"))
		case "facebook":
			goth.UseProviders(facebook.New(clientId, clientSecret, clientAuthCallback, "email", "profile"))
		case "bitbucket":
			goth.UseProviders(bitbucket.New(clientId, clientSecret, clientAuthCallback, "email", "profile"))
		case "gitlab":
			goth.UseProviders(gitlab.New(clientId, clientSecret, clientAuthCallback, "email", "profile"))
		}
	}
}
