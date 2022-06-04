package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func LoadPublicRoutes(router *gin.RouterGroup) {

	auth := router.Group("/auth")
	{
		auth.POST("/login", func(context *gin.Context) {
			type Body struct {
				Email    string `json:"email" binding:"required"`
				Password string `json:"password" binding:"required"`
			}
			body := Body{}
			err := context.ShouldBindJSON(&body)
			if handlerError(context, err, http.StatusBadRequest) {
				return
			}
			user, err := GetItem(toKey("user", body.Email), toKey("user", body.Email))

			if handlerError(context, err, http.StatusBadRequest) {
				return
			}
			if user != nil {
				log.Println(user)
				check, _ := VerifyPassword(body.Password, user["password"].(string))
				if check {
					userContext, err := GetUserContext(user["email"].(string))
					if handlerError(context, err, http.StatusBadRequest) {
						return
					}
					token, refreshToken, _ := GenerateToken(*userContext)
					context.JSON(http.StatusOK, gin.H{
						"token":         token,
						"refresh_token": refreshToken,
						"message":       "Login success",
					})
					return
				}
			}
			handlerError(context, errors.New("Failed to validate"), http.StatusForbidden)
		})

		auth.POST("/signup", func(context *gin.Context) {
			type Body struct {
				Email    string      `json:"email" binding:"required"`
				Password string      `json:"password" binding:"required"`
				Data     interface{} `json:"data"`
			}
			body := Body{}
			err := context.ShouldBindJSON(&body)
			if handlerError(context, err, http.StatusBadRequest) {
				return
			}
			user, _ := GetItem(toKey("user", body.Email), toKey("user", body.Email))
			if user != nil {
				if handlerError(context, errors.New("Already exists"), http.StatusBadRequest) {
					return
				}
			}
			password := HashPassword(body.Password)
			_, err = CreateUser(UserPayload{
				Email:    body.Email,
				Password: password,
				Data:     body.Data,
			})
			if handlerError(context, err, http.StatusBadRequest) {
				return
			}
			context.JSON(http.StatusOK, gin.H{
				"message": "Signup success",
				"email":   body.Email,
			})
		})

		auth.GET("/validate", func(context *gin.Context) {
			claims, valid := ValidateTokenMiddleware(context)
			if valid == false {
				handlerError(context, errors.New("Unauthorized"), http.StatusForbidden)
				return
			}
			context.JSON(http.StatusOK, claims)
		})

		auth.POST("/renew", func(context *gin.Context) {
			claims, valid := ValidateTokenMiddleware(context)
			if valid == false {
				handlerError(context, errors.New("Unauthorized"), http.StatusForbidden)
				return
			}
			userContext, err := GetUserContext(claims.Email)
			if handlerError(context, err, http.StatusBadRequest) {
				return
			}
			token, _, _ := GenerateToken(*userContext)
			context.JSON(http.StatusOK, gin.H{
				"token": token,
			})
			return
		})
	}
}
