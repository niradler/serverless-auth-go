package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func LoadPublicRoutes(router *gin.RouterGroup) {

	public := router.Group("/auth")
	{
		public.POST("/login", func(context *gin.Context) {
			type Body struct {
				Email    string `json:"email" binding:"required"`
				Password string `json:"password" binding:"required"`
			}
			body := Body{}
			err := context.ShouldBindJSON(&body)
			if handlerError(context, err, http.StatusBadRequest) {
				return
			}
			user, err := GetItem(toKey("user",body.Email) , toKey("user",body.Email))
			if err != nil {
				log.Println(err)
			}
			if user != nil {
				log.Println(user)
				check, _ := VerifyPassword(body.Password, user["password"].(string))
				if check {
					token, refreshToken, _ := GenerateToken(user)
					context.JSON(http.StatusOK, gin.H{
						"token":         token,
						"refresh_token": refreshToken,
						"message":       "Login success",
					})
					return
				}
			}

			context.AbortWithStatusJSON(http.StatusBadRequest,
				gin.H{
					"error":   "ValidationError",
					"message": "Failed to validate",
				})
		})

		public.POST("/signup", func(context *gin.Context) {
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
			user, _ := GetItem(toKey("user",body.Email), toKey("user",body.Email))
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
	}
}
