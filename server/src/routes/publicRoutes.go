package routes

import (
	"errors"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/niradler/social-lab/src/auth"
	"github.com/niradler/social-lab/src/db"
	"github.com/niradler/social-lab/src/types"
	"github.com/niradler/social-lab/src/utils"
	"go.uber.org/zap"
)

func LoadPublicRoutes(router *gin.RouterGroup) {

	authRouter := router.Group("/auth")
	{

		authRouter.POST("/login", func(context *gin.Context) {
			type Body struct {
				Email    string `json:"email" binding:"required"`
				Password string `json:"password" binding:"required"`
			}
			body := Body{}
			err := context.ShouldBindJSON(&body)
			if utils.HandlerError(context, err, http.StatusBadRequest) {
				return
			}
			user, err := db.GetItem(db.ToKey("user", body.Email), "#")

			if utils.HandlerError(context, err, http.StatusBadRequest) {
				return
			}
			if user != nil && user["password"].(string) != "" {
				check, _ := auth.VerifyPassword(body.Password, user["password"].(string))
				if check {
					userContext, err := db.GetUserContext(user["email"].(string))
					if utils.HandlerError(context, err, http.StatusBadRequest) {
						return
					}
					token, refreshToken, _ := auth.GenerateToken(*userContext)
					context.JSON(http.StatusOK, gin.H{
						"token":         token,
						"refresh_token": refreshToken,
						"message":       "Login success",
					})
					return
				}
			}
			utils.HandlerError(context, errors.New("Failed to validate"), http.StatusForbidden)
		})

		authRouter.POST("/signup", func(context *gin.Context) {
			type Body struct {
				Email    string      `json:"email" binding:"required"`
				Password string      `json:"password" binding:"required"`
				Data     interface{} `json:"data"`
			}
			body := Body{}
			err := context.ShouldBindJSON(&body)
			if utils.HandlerError(context, err, http.StatusBadRequest) {
				return
			}
			user, _ := db.GetItem(db.ToKey("user", body.Email), "#")
			if user != nil {
				if utils.HandlerError(context, errors.New("Already exists"), http.StatusBadRequest) {
					return
				}
			}
			password := auth.HashPassword(body.Password)
			_, err = db.CreateUser(types.UserPayload{
				Email:    body.Email,
				Password: password,
				Data:     body.Data,
			})
			if utils.HandlerError(context, err, http.StatusBadRequest) {
				return
			}
			context.JSON(http.StatusOK, gin.H{
				"message": "Signup success",
				"email":   body.Email,
			})
		})

		authRouter.POST("/login/email", func(context *gin.Context) {
			type Body struct {
				Email string `json:"email" binding:"required"`
			}
			body := Body{}
			err := context.ShouldBindJSON(&body)
			if utils.HandlerError(context, err, http.StatusBadRequest) {
				return
			}
			utils.Logger.Info("Email login", zap.String("email", body.Email))
			userContext, err := db.GetUserContext(body.Email)

			if utils.HandlerError(context, err, http.StatusInternalServerError) {
				return
			}
			if userContext != nil {
				token, _, err := auth.GenerateToken(*userContext)

				err = utils.SendEmail(
					utils.EmailRequest{
						To:       userContext.Email,
						Subject:  "Passwordless Login",
						Template: "email_login.html",
						Args: map[string]string{
							"Logo":    os.Getenv("SLS_AUTH_APP_NAME"),
							"URL":     os.Getenv("SLS_AUTH_CLIENT_CALLBACK") + "?token=" + token,
							"Contact": os.Getenv("SLS_AUTH_APP_CONTACT"),
							"Company": os.Getenv("SLS_AUTH_APP_NAME"),
						},
					})
				if utils.HandlerError(context, err, http.StatusInternalServerError) {
					return
				}
			}

			context.JSON(http.StatusOK, gin.H{
				"complete": "Successfully",
			})
			return
		})

		authRouter.GET("/validate", func(context *gin.Context) {
			claims, err := auth.ValidateToken(context.Request.Header.Get("Authorization"))
			if err != nil {
				utils.HandlerError(context, err, http.StatusForbidden)
				return
			}
			context.JSON(http.StatusOK, claims)
		})

		authRouter.POST("/renew", func(context *gin.Context) {
			claims, err := auth.ValidateRefreshToken(context.Request.Header.Get("Authorization"))
			if err != nil {
				utils.HandlerError(context, err, http.StatusForbidden)
				return
			}
			userContext, err := db.GetUserContext(claims.Email)
			if utils.HandlerError(context, err, http.StatusBadRequest) {
				return
			}
			token, _, _ := auth.GenerateToken(*userContext)
			context.JSON(http.StatusOK, gin.H{
				"token": token,
			})
			return
		})

		authRouter.GET("/provider/:provider/callback", func(ctx *gin.Context) {
			utils.Logger.Info("callback auth", zap.String("provider", ctx.Param("provider")))
			auth.ProvidersAuthCallback(ctx.Param("provider"), ctx)
		})

		authRouter.GET("/provider/:provider", func(ctx *gin.Context) {
			utils.Logger.Info("start auth", zap.String("provider", ctx.Param("provider")))
			auth.ProvidersAuthBegin(ctx.Param("provider"), ctx)
		})
	}

	router.GET("/hc", func(context *gin.Context) {

		context.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})
}
